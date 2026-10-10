package app

import (
	"net/http"
	"strings"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/assets"
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
	mux.HandleFunc("POST /account/leaderboard", c.AccountLeaderboard)
	mux.HandleFunc("POST /account/nickname", c.AccountNickname)
	mux.HandleFunc("GET /privacy", c.Privacy)

	// API and realtime.
	mux.HandleFunc("GET /api/weather", c.WeatherAPI)
	// Place search is limited per client: a burst of 5, then one a second.
	placeLimit := middleware.RateLimit(ratelimit.New(1, 5, time.Now), a.cfg.TrustProxy)
	mux.Handle("GET /api/places", placeLimit(http.HandlerFunc(c.PlacesAPI)))
	mux.HandleFunc("GET /api/stats", c.StatsAPI)
	mux.HandleFunc("GET /api/stats/demo", c.DemoStats)
	mux.HandleFunc("GET /api/stats/community", c.CommunityAPI)
	mux.HandleFunc("GET /api/stats/leaderboard", c.LeaderboardAPI)
	mux.HandleFunc("GET /admin", c.Admin)
	mux.HandleFunc("GET /api/admin/overview", c.AdminOverviewAPI)
	mux.HandleFunc("GET /api/admin/users", c.AdminUsersAPI)
	mux.Handle("GET /ws/{id}", ws)
	mux.HandleFunc("GET /healthz", c.Health)
	mux.HandleFunc("GET /robots.txt", c.Robots)

	// Development only: browser auto-reload (never registered in production).
	if a.dev != nil {
		mux.Handle("GET /__dev/events", a.dev.EventsHandler())
		mux.Handle("GET /__dev/reload.js", devreload.ScriptHandler())
		mux.HandleFunc("GET /__dev/sky", c.SkyLab)
		mux.HandleFunc("GET /__dev/mobile", c.MobileLab)
	}

	// Static assets built from web/.
	static := a.staticHandler()
	mux.Handle("GET /static/", static)

	// Everything else.
	mux.HandleFunc("/", c.NotFound)

	return middleware.Chain(mux,
		middleware.Recover(a.log),
		middleware.RequestID(),
		middleware.Logging(a.log),
		a.securityHeaders(),
		middleware.OptionalAuth(a.sessions),
		middleware.CSRF(a.cfg.IsProduction()),
	)
}

// securityHeaders is the strict policy, except that in development the site may
// frame itself so the mobile preview can show pages at phone width.
func (a *App) securityHeaders() middleware.Middleware {
	if a.dev != nil {
		return middleware.SecurityHeaders(middleware.DevCSP, true)
	}
	return middleware.SecurityHeaders(middleware.DefaultCSP)
}

// staticHandler serves the built frontend. Files with a content hash in their
// name never change, so browsers may keep them for a year. Everything else
// (favicon, share image) is revalidated. Dot files such as the Vite manifest
// are not served. In development nothing is cached.
func (a *App) staticHandler() http.Handler {
	files := http.StripPrefix("/static/", http.FileServer(http.Dir(a.cfg.StaticDir)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rel := strings.TrimPrefix(r.URL.Path, "/static/")
		for _, part := range strings.Split(rel, "/") {
			if strings.HasPrefix(part, ".") {
				http.NotFound(w, r)
				return
			}
		}
		switch {
		case a.dev != nil:
			// Development: never reuse an old build of the frontend.
			w.Header().Set("Cache-Control", "no-cache")
		case assets.Hashed(rel):
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		default:
			w.Header().Set("Cache-Control", "no-cache")
		}
		files.ServeHTTP(w, r)
	})
}
