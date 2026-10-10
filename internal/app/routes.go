package app

import (
	"net/http"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/controllers"
	"github.com/gayratjon-02/Pomosky/internal/devreload"
	"github.com/gayratjon-02/Pomosky/internal/middleware"
	"github.com/gayratjon-02/Pomosky/internal/platform/ratelimit"
)

// routes registers every route and wraps the mux with the middleware stack.
func (a *App) routes(c *controllers.Controllers, ws http.Handler) http.Handler {
	mux := http.NewServeMux()

	// Pages (SSR).
	mux.HandleFunc("GET /{$}", c.Home)
	mux.HandleFunc("GET /r/{id}", c.RoomShow)
	mux.HandleFunc("GET /r/{id}/host", c.HostPanel)
	// Creating a room is limited per client: 3 at once, then one every 10 seconds.
	createRoom := middleware.RateLimit(ratelimit.New(0.1, 3, time.Now), a.cfg.TrustProxy)
	mux.Handle("POST /r/new", createRoom(http.HandlerFunc(c.RoomCreate)))
	mux.HandleFunc("GET /login", c.Login)
	// Login attempts are limited per client: a burst of 5, then one every 5 seconds.
	loginLimit := middleware.RateLimit(ratelimit.New(0.2, 5, time.Now), a.cfg.TrustProxy)
	mux.Handle("POST /login", loginLimit(http.HandlerFunc(c.LoginSubmit)))
	mux.HandleFunc("GET /auth/callback", c.AuthCallback)
	mux.HandleFunc("GET /auth/google/callback", c.GoogleCallback)
	mux.Handle("POST /auth/callback", loginLimit(http.HandlerFunc(c.AuthConfirm)))
	mux.HandleFunc("POST /logout", c.Logout)
	mux.HandleFunc("GET /account", c.Account)
	mux.HandleFunc("POST /account/delete", c.AccountDelete)
	mux.HandleFunc("GET /privacy", c.Privacy)

	// API and realtime.
	mux.HandleFunc("GET /api/weather", c.WeatherAPI)
	mux.HandleFunc("GET /api/stats", c.StatsAPI)
	mux.Handle("GET /ws/{id}", ws)
	mux.HandleFunc("GET /healthz", c.Health)

	// Development only: browser auto-reload (never registered in production).
	if a.dev != nil {
		mux.Handle("GET /__dev/events", a.dev.EventsHandler())
		mux.Handle("GET /__dev/reload.js", devreload.ScriptHandler())
		mux.HandleFunc("GET /__dev/sky", c.SkyLab)
		mux.HandleFunc("GET /__dev/account", c.DevAccount)
		mux.HandleFunc("GET /__dev/api/stats", c.DevStats)
	}

	// Static assets built from web/.
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(a.cfg.StaticDir))))

	// Everything else.
	mux.HandleFunc("/", c.NotFound)

	return middleware.Chain(mux,
		middleware.Recover(a.log),
		middleware.RequestID(),
		middleware.Logging(a.log),
		middleware.SecurityHeaders(middleware.DefaultCSP),
		middleware.OptionalAuth(a.sessions),
		middleware.CSRF(a.cfg.IsProduction()),
	)
}
