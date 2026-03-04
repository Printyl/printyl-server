package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gregor-gottschewski/printyl-server/internal"
	// This controls the maxprocs environment variable in container runtimes.
	// see https://martin.baillie.id/wrote/gotchas-in-the-go-network-packages-defaults/#bonus-gomaxprocs-containers-and-the-cfs
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/gregor-gottschewski/printyl-server/internal/log"
)

func main() {
	slog.SetDefault(log.New(
		log.WithLevel(os.Getenv("LOG_LEVEL")),
		log.WithSource(),
	))

	if err := run(); err != nil {
		slog.ErrorContext(context.Background(), "an error occurred", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	_, err := maxprocs.Set(maxprocs.Logger(func(s string, i ...interface{}) {
		slog.DebugContext(ctx, fmt.Sprintf(s, i...))
	}))
	if err != nil {
		return fmt.Errorf("setting max procs: %w", err)
	}

	if err := loadConfig(); err != nil {
		return err
	}

	if err := createApi(); err != nil {
		return err
	}

	// Block until we receive a termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.InfoContext(ctx, "server shutting down")

	return nil
}

// loadConfig loads application configuration
func loadConfig() error {
	if err := internal.LoadConfig(); err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	return nil
}

// createApi creates all server APIs
// Currently, only v1 API
func createApi() error {
	api := internal.NewAPI()
	if err := api.Start(); err != nil {
		return err
	}
	return nil
}
