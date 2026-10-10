# syntax=docker/dockerfile:1

# ---- frontend build
FROM node:22-alpine AS web
WORKDIR /web
COPY web/package.json web/package-lock.json* ./
RUN npm ci || npm install
COPY web/ ./
RUN npm run build

# ---- backend build
ARG GO_VERSION=1.27
FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/pomosky ./cmd/server

# ---- runtime
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/pomosky /pomosky
COPY --from=web /web/dist /app/static
ENV STATIC_DIR=/app/static APP_ENV=production ADDR=:8080
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/pomosky"]
