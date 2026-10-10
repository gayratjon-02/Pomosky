package app

import (
	"net/http"

	"github.com/gayratjon-02/Pomosky/internal/controllers"
	"github.com/gayratjon-02/Pomosky/internal/devreload"
	"github.com/gayratjon-02/Pomosky/internal/middleware"
)

// routes registers every route and wraps the mux with the middleware stack.
func (a *App) routes(c *controllers.Controllers, ws http.Handler) http.Handler {
	mux := http.NewServeMux()

	// Pages (SSR).
	mux.HandleFunc("GET /{$}", c.Home)
	mux.HandleFunc("GET /r/{id}", c.RoomShow)
	mux.HandleFunc("GET /r/{id}/host", c.HostPanel)
	mux.HandleFunc("POST /r/new", c.RoomCreate)
	mux.HandleFunc("GET /login", c.Login)
	mux.HandleFunc("POST /login", c.LoginSubmit)
	mux.HandleFunc("GET /auth/callback", c.AuthCallback)
	mux.HandleFunc("POST /logout", c.Logout)
	mux.HandleFunc("GET /account", c.Account)
	mux.HandleFunc("GET /privacy", c.Privacy)

	// API and realtime.
	mux.HandleFunc("GET /api/weather", c.WeatherAPI)
	mux.Handle("GET /ws/{id}", ws)
	mux.HandleFunc("GET /healthz", c.Health)

	// Development only: browser auto-reload (never registered in production).
	if a.dev != nil {
		mux.Handle("GET /__dev/events", a.dev.EventsHandler())
		mux.Handle("GET /__dev/reload.js", devreload.ScriptHandler())
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
		middleware.OptionalAuth(nil), // TODO(stage 4): pass auth.Service
	)
}
