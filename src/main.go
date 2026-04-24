package main

import (
	"drblury/event-driven-service/internal/app"

	_ "github.com/drblury/protoflow/transport/aws"
	_ "github.com/drblury/protoflow/transport/channel"
	_ "github.com/drblury/protoflow/transport/io"
	_ "github.com/drblury/protoflow/transport/jetstream"
	_ "github.com/drblury/protoflow/transport/kafka"
	_ "github.com/drblury/protoflow/transport/nats"
	_ "github.com/drblury/protoflow/transport/postgres"
	_ "github.com/drblury/protoflow/transport/rabbitmq"
	_ "github.com/drblury/protoflow/transport/sqlite"
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
