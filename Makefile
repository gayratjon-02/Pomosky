.PHONY: run dev build test test-db vet fmt web-install web-dev web-build migrate-up migrate-down sqlc up down

run: web-build
	go run ./cmd/server

# Live development, stop with Ctrl+C: Go rebuilds on change (air), the frontend
# rebuilds (vite --watch), templates are read from disk and the browser reloads itself.
# Set ADDR to change the port, e.g. `ADDR=:8081 make dev`.
dev:
	@test -d web/node_modules || (cd web && npm install)
	@trap 'kill 0' INT TERM EXIT; \
	(cd web && npm run watch) & \
	DEV_RELOAD=true go run github.com/air-verse/air@v1.67.4 -c .air.toml & \
	wait

build:
	go build -o bin/pomosky ./cmd/server

test:
	go test ./...

# Also runs the tests that need PostgreSQL. They empty the tables of the database, so use a
# throwaway one, e.g. TEST_DATABASE_URL=postgres://user@localhost:5432/pomosky_test?sslmode=disable
test-db:
	@test -n "$(TEST_DATABASE_URL)" || (echo "set TEST_DATABASE_URL"; exit 1)
	go test -p 1 ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

web-install:
	cd web && npm install

web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build

# Requires the goose CLI: https://github.com/pressly/goose
migrate-up:
	goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir migrations postgres "$(DATABASE_URL)" down

# Regenerates internal/store/postgres/sqlcgen from the SQL in queries/.
sqlc:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0 generate

up:
	cd deploy && docker compose up -d --build

down:
	cd deploy && docker compose down
