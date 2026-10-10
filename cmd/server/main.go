// Command server runs the Pomosky web server.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	_ "time/tzdata" // time zone names for statistics, also on minimal images

	"github.com/gayratjon-02/Pomosky/internal/app"
	"github.com/gayratjon-02/Pomosky/internal/config"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	// A .env file is a local-development convenience; production gets its
	// environment from the deployment (docker compose env_file, systemd).
	if os.Getenv("APP_ENV") != "production" {
		if err := config.LoadDotenv(".env"); err != nil {
			return err
		}
	}
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	a, err := app.New(cfg)
	if err != nil {
		return err
	}
	defer a.Close()
	return a.Run(ctx)
}
