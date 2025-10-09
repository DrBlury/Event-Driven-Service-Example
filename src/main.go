package main

import (
	"drblury/event-driven-service/internal/app"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
)

// Application metadata that is set at compile time.
// nolint
var (
	version     string
	buildDate   string
	description = "CPC contract facade"
	commitHash  string
	commitDate  string
)

// main just loads config and inits logger. Rest is done in app.Run.
func main() {
	appCfg, err := app.LoadConfig(
		version,
		buildDate,
		description,
		commitHash,
		commitDate,
	)
	if err != nil {
		fmt.Printf("could not load config: %s", err.Error())
		os.Exit(1)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	err = app.Run(
		appCfg,
		quit,
	)

	if err != nil {
		slog.Error("error running app", "error", err)
	}
}
