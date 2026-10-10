package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/admin"
	"github.com/gayratjon-02/Pomosky/internal/assets"
	"github.com/gayratjon-02/Pomosky/internal/auth"
	"github.com/gayratjon-02/Pomosky/internal/config"
	"github.com/gayratjon-02/Pomosky/internal/controllers"
	"github.com/gayratjon-02/Pomosky/internal/devreload"
	"github.com/gayratjon-02/Pomosky/internal/middleware"
	"github.com/gayratjon-02/Pomosky/internal/platform/clock"
	"github.com/gayratjon-02/Pomosky/internal/platform/httpx"
	"github.com/gayratjon-02/Pomosky/internal/platform/idgen"
	"github.com/gayratjon-02/Pomosky/internal/platform/logger"
	"github.com/gayratjon-02/Pomosky/internal/ports"
	"github.com/gayratjon-02/Pomosky/internal/realtime"
	"github.com/gayratjon-02/Pomosky/internal/stats"
	"github.com/gayratjon-02/Pomosky/internal/store/postgres"
	"github.com/gayratjon-02/Pomosky/internal/views"
	"github.com/gayratjon-02/Pomosky/internal/weather"
	"github.com/gayratjon-02/Pomosky/internal/weather/openmeteo"
)

// App is the assembled server.
type App struct {
	cfg      config.Config
	log      *slog.Logger
	handler  http.Handler
	dev      *devreload.Watcher // nil unless cfg.Dev
	db       *postgres.DB       // nil when DATABASE_URL is not set
	hub      *realtime.Hub
	janitor  *janitor        // nil without a database
	recorder *stats.Recorder // nil without a database
	// sessions resolves login cookies; nil means everyone is anonymous.
	sessions middleware.SessionResolver
}

// New wires the application together.
//
// Repositories, the weather service and auth are not wired yet (they are nil
// in the controller Deps) and arrive in the stages described in docs/plan.
func New(cfg config.Config) (*App, error) {
	log := logger.New(cfg.Env)

	var (
		renderer *views.Renderer
		watcher  *devreload.Watcher
		err      error
	)
	if cfg.Dev {
		// Development: templates are read from disk on every request and the
		// browser reloads when templates or built assets change.
		renderer, err = views.FromDir(cfg.DevViewsDir)
		watcher = devreload.NewWatcher(500*time.Millisecond,
			cfg.StaticDir, filepath.Join(cfg.DevViewsDir, "templates"))
	} else {
		renderer, err = views.Default()
	}
	if err != nil {
		return nil, err
	}

	// Built frontend files carry a content hash; the manifest says which is current.
	renderer.SetAssetResolver(assets.Load(cfg.StaticDir, cfg.Dev).Resolve)

	clk := clock.Real{}
	ids := idgen.Generator{}

	// Login, rooms of logged-in hosts and statistics need the database. Without
	// DATABASE_URL the server still runs: rooms, timer and reactions are anonymous.
	var (
		db       *postgres.DB
		accounts *auth.Service
		statsSvc *stats.Service
		rooms    ports.RoomRepository
		recorder *stats.Recorder
		jan      *janitor
	)
	if cfg.DatabaseURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		db, err = postgres.Open(ctx, cfg.DatabaseURL)
		if err != nil {
			return nil, err
		}
		if err = db.Migrate(ctx); err != nil {
			_ = db.Close()
			return nil, err
		}
		accounts = auth.NewService(auth.Deps{
			Users:     postgres.NewUserRepo(db),
			Sessions:  postgres.NewSessionRepo(db),
			Tokens:    postgres.NewLoginTokenRepo(db),
			Mailer:    newMailer(cfg, log),
			Clock:     clk,
			IDs:       ids,
			Providers: oauthProviders(cfg),
		}, auth.Options{BaseURL: cfg.BaseURL})
		focusRepo := postgres.NewFocusSessionRepo(db)
		statsSvc = stats.New(focusRepo, clk)
		recorder = stats.NewRecorder(focusRepo, log)
		roomRepo := postgres.NewRoomRepo(db)
		rooms = roomRepo
		jan = &janitor{log: log, clock: clk, sessions: postgres.NewSessionRepo(db), tokens: postgres.NewLoginTokenRepo(db), rooms: roomRepo}
	}

	var focusRecorder ports.FocusRecorder // stays a nil interface without a database
	if recorder != nil {
		focusRecorder = recorder
	}
	hub := realtime.NewHub(log, clk, focusRecorder)
	if cfg.RoomIdleTTL > 0 {
		hub.SetIdleTTL(cfg.RoomIdleTTL)
	}

	deps := controllers.Deps{
		Config: cfg,
		Log:    log,
		Views:  renderer,
		Clock:  clk,
		IDs:    ids,
		Rooms:  rooms,

		Live:       hub,
		HostTokens: auth.HostTokens{Salt: cfg.HostTokenSalt},
	}
	// Assigned only when set: a nil *Service inside the interface would not be nil.
	if accounts != nil {
		deps.Accounts = accounts
		deps.Stats = statsSvc
	}
	if cfg.Weather.BaseURL != "" {
		provider := openmeteo.New(cfg.Weather.BaseURL, nil)
		deps.Weather = weather.NewService(provider, cfg.Weather.CacheTTL, clk.Now)
		deps.Places = openmeteo.NewGeocoder(cfg.Weather.GeocodingURL, nil)
	}
	if db != nil {
		var counters admin.WeatherCounters // a nil interface when weather is off
		if ws, ok := deps.Weather.(*weather.Service); ok {
			counters = ws
		}
		deps.Admin = admin.New(postgres.NewAdminRepo(db), hub, counters, clk)
	}
	ctrl := controllers.New(deps)

	wsHandler, err := realtime.NewHandler(hub, log, realtime.HandlerConfig{
		Hosts: auth.HostTokens{Salt: cfg.HostTokenSalt},
		Clock: clk,
		IDs:   idgen.Generator{},

		ClientIP: func(r *http.Request) string { return httpx.ClientIP(r, cfg.TrustProxy) },
		Identify: roomIdentity,
	})
	if err != nil {
		return nil, err
	}

	a := &App{cfg: cfg, log: log, dev: watcher, db: db, hub: hub, recorder: recorder, janitor: jan}
	if accounts != nil {
		a.sessions = accounts
	}
	a.handler = a.routes(ctrl, wsHandler)
	return a, nil
}

// Close releases resources held by the app (the database pool).
func (a *App) Close() error {
	if a.recorder != nil {
		a.recorder.Close() // write the queued focus sessions before the pool goes away
	}
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// Handler returns the root HTTP handler (used by tests).
func (a *App) Handler() http.Handler { return a.handler }

// Run serves HTTP until ctx is cancelled, then shuts down gracefully.
func (a *App) Run(ctx context.Context) error {
	srv := &http.Server{
		Addr:              a.cfg.Addr,
		Handler:           a.handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	if a.dev != nil {
		a.log.Info("dev reload enabled", "watching", a.cfg.StaticDir)
		go a.dev.Run(ctx)
		srv.RegisterOnShutdown(a.dev.Close) // open event streams would block a graceful shutdown
	}

	// Focus phases that run to their end are noticed by a sweep, not by an event.
	go a.hub.RunSweeper(ctx, 5*time.Second)
	if a.janitor != nil {
		go a.janitor.run(ctx, time.Hour)
	}

	errCh := make(chan error, 1)
	go func() {
		a.log.Info("server listening", "addr", a.cfg.Addr, "env", a.cfg.Env)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		a.log.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}

// roomIdentity is what the room may show of a signed-in browser: the public
// nickname and picture, only for people who allow being listed (the same
// switch as the leaderboard). The session user was resolved by the middleware.
func roomIdentity(r *http.Request) (key, name, avatarURL string, ok bool) {
	u := middleware.UserFromContext(r.Context())
	if u == nil || !u.ShowOnLeaderboard || u.Nickname == "" {
		return "", "", "", false
	}
	return string(u.ID), u.Nickname, u.AvatarURL, true
}
