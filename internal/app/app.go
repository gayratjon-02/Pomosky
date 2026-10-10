package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/config"
	"github.com/gayratjon-02/Pomosky/internal/controllers"
	"github.com/gayratjon-02/Pomosky/internal/devreload"
	"github.com/gayratjon-02/Pomosky/internal/platform/clock"
	"github.com/gayratjon-02/Pomosky/internal/platform/idgen"
	"github.com/gayratjon-02/Pomosky/internal/platform/logger"
	"github.com/gayratjon-02/Pomosky/internal/realtime"
	"github.com/gayratjon-02/Pomosky/internal/views"
)

// App is the assembled server.
type App struct {
	cfg     config.Config
	log     *slog.Logger
	handler http.Handler
	dev     *devreload.Watcher // nil unless cfg.Dev
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

	clk := clock.Real{}
	hub := realtime.NewHub(log, clk)

	ctrl := controllers.New(controllers.Deps{
		Config: cfg,
		Log:    log,
		Views:  renderer,
		Clock:  clk,
		IDs:    idgen.Generator{},
	})

	a := &App{cfg: cfg, log: log, dev: watcher}
	a.handler = a.routes(ctrl, realtime.NewHandler(hub, log))
	return a, nil
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
