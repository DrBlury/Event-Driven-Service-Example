package main

import (
	"drblury/event-driven-service/internal/app"

	// Import all protoflow transports for auto-registration
	_ "github.com/drblury/protoflow/transport/transports"
)

// Application metadata that is set at compile time.
var (
	version     string
	buildDate   string
	description = "application template"
	commitHash  string
	commitDate  string
)

func main() {
	app.New(app.Metadata{
		Version:     version,
		BuildDate:   buildDate,
		Description: description,
		CommitHash:  commitHash,
		CommitDate:  commitDate,
	}, nil).Run()
}
