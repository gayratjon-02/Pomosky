# Environment variables

Local development reads `.env` (copy `.env.example`); a variable already set in the shell wins.
Production reads `deploy/.env` through docker compose (create it with `deploy/scripts/init-env.sh`).
Neither file is ever committed. A test fails if the code reads a variable that is not listed here.

## Basics

| Variable | Default | Meaning |
|---|---|---|
| `APP_ENV` | `development` | `production` turns on the strict checks below and secure cookies |
| `ADDR` | `:8080` | Listen address (`PORT` is used if `ADDR` is not set) |
| `PORT` | `8080` | Only used to build the default `ADDR` |
| `BASE_URL` | `http://localhost:8080` | Public address, used in emailed links, share links and the Google redirect |
| `STATIC_DIR` | `web/dist` | Folder with the built frontend |
| `TRUST_PROXY` | `false` | Believe `X-Forwarded-For`. `true` only behind a proxy you control (Caddy), or clients can fake their IP and dodge rate limits |

## Required in production

| Variable | Meaning |
|---|---|
| `DATABASE_URL` | PostgreSQL address. Compose builds it from `POSTGRES_PASSWORD` |
| `HOST_TOKEN_SALT` | Long random text mixed into host-token hashes. Changing it locks hosts out of running rooms |
| `POSTGRES_PASSWORD` | Compose only: password of the database container |
| `DOMAIN` | Compose only: the domain Caddy gets a certificate for |

## Admin

| Variable | Meaning |
|---|---|
| `ADMIN_EMAILS` | Comma separated emails of the operators. A logged-in user whose verified email is in the list can open `/admin`; everyone else gets a normal 404 |
| `ADMIN_HOST` | Host that serves `/admin` and `/api/admin/*`, for example `dash.example.com`. When set, everywhere else answers 404 and the root of that host redirects to `/admin`. Empty: `/admin` works on every host. It needs its own DNS record and Caddy block |
| `COOKIE_DOMAIN` | Domain of the login cookie, for example `.example.com`, so one login also works on `ADMIN_HOST`. Empty: the cookie belongs to the one host that set it |
| `GEOIP_FILE` | Path of the DB-IP "IP to Country Lite" CSV (plain or `.gz`), for example `/opt/pomosky/geoip/dbip-country-lite.csv.gz`. Fills in the country of visits when no `CF-IPCountry` header exists. Empty: countries need a proxy. Refresh monthly with `deploy/bare/update-geoip.sh`. Data by DB-IP.com, CC BY 4.0 |
| `VAPID_PUBLIC_KEY`, `VAPID_PRIVATE_KEY` | Web Push keys; both are needed. They switch on alerts for a finished phase that arrive even when the page is closed or the phone is locked. Create a pair with `go run ./cmd/vapid`. Empty: alerts only show while the page is open in a background tab |
| `VAPID_SUBJECT` | Contact the push services may use, `mailto:you@example.com` or an `https:` URL. Defaults to `BASE_URL` |

## Rooms

| Variable | Default | Meaning |
|---|---|---|
| `ROOM_IDLE_TTL` | `10m` | How long a room with nobody connected stays alive |

## Login (optional, needs `DATABASE_URL`)

| Variable | Meaning |
|---|---|
| `OAUTH_GOOGLE_CLIENT_ID`, `OAUTH_GOOGLE_CLIENT_SECRET` | Google sign-in. Register the redirect URI `BASE_URL/auth/google/callback` |
| `MAIL_FROM`, `SMTP_HOST`, `SMTP_PORT` (587), `SMTP_USERNAME`, `SMTP_PASSWORD` | Email sign-in links over SMTP with STARTTLS. In development without SMTP the link is printed in the log; in production without SMTP email login is hidden |

## Weather

| Variable | Default | Meaning |
|---|---|---|
| `OPEN_METEO_BASE` | `https://api.open-meteo.com` | Forecast API |
| `OPEN_METEO_GEOCODING_BASE` | `https://geocoding-api.open-meteo.com` | City search |
| `WEATHER_CACHE_TTL` | `10m` | How long a place's weather is reused |

The free Open-Meteo API is for non-commercial use only (no ads, no subscriptions), under
10 000 calls a day, and its data is CC BY 4.0 (the pages credit Open-Meteo.com). A commercial
product needs their paid plan and API key.

## Development only

| Variable | Meaning |
|---|---|
| `DEV_RELOAD` | Read templates from disk and reload the browser on changes. The server refuses to start with it in production |
| `DEV_VIEWS_DIR` | Where the templates are read from (default `internal/views`) |
