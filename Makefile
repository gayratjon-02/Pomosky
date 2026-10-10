.PHONY: run build test vet fmt web-install web-dev web-build migrate-up migrate-down sqlc up down

run: web-build
	go run ./cmd/server

build:
	go build -o bin/pomosky ./cmd/server

test:
	go test ./...

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

# Requires sqlc: https://sqlc.dev
sqlc:
	sqlc generate

up:
	cd deploy && docker compose up -d --build

down:
	cd deploy && docker compose down
