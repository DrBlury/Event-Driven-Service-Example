package database

import (
	"context"
	"log/slog"
	"time"

	"drblury/event-driven-service/internal/observability"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Database struct {
	DB  *mongo.Database
	Cfg *Config
}

var tracer = otel.Tracer("drblury/event-driven-service/internal/database")

func NewDatabase(cfg *Config, logger *slog.Logger, ctx context.Context) (*Database, error) {
	db := &Database{Cfg: cfg}

	if err := db.Connect(ctx, logger); err != nil {
		return nil, err
	}

	return db, nil
}

// Connect initializes the MongoDB client and verifies connectivity.
func (db *Database) Connect(ctx context.Context, logger *slog.Logger) error {
	ctx, span := tracer.Start(ctx, "database.connect")
	defer span.End()

	log := observability.Logger(ctx, logger)

	if err := db.validateConnectState(ctx, span); err != nil {
		return err
	}

	span.SetAttributes(
		attribute.String("db.system", "mongodb"),
		attribute.String("db.name", db.Cfg.MongoDB),
		attribute.String("db.user", db.Cfg.MongoUser),
	)

	log.Info("connecting to MongoDB", "db", db.Cfg.MongoDB, "user", db.Cfg.MongoUser)

	client, err := db.connectClient(ctx, span)
	if err != nil {
		return err
	}

	db.DB = client.Database(db.Cfg.MongoDB)
	log.Info("connected to MongoDB successfully", "db", db.Cfg.MongoDB)

	return nil
}

func (db *Database) validateConnectState(ctx context.Context, span trace.Span) error {
	if db == nil {
		err := observability.Builder(ctx, "database.connect", "database_not_configured").
			Public("database configuration is invalid").
			New("database not configured")
		observability.RecordError(span, err)
		return err
	}
	if db.Cfg == nil {
		err := observability.Builder(ctx, "database.connect", "database_config_missing").
			Public("database configuration is invalid").
			New("database configuration is required")
		observability.RecordError(span, err)
		return err
	}
	if db.DB != nil {
		span.SetAttributes(attribute.Bool("database.already_connected", true))
	}
	return nil
}

func (db *Database) connectClient(ctx context.Context, span trace.Span) (*mongo.Client, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(db.Cfg.MongoURL)
	clientOpts.Auth = &options.Credential{
		Username: db.Cfg.MongoUser,
		Password: db.Cfg.MongoPassword,
	}

	if err := clientOpts.Validate(); err != nil {
		wrapped := observability.Builder(ctxWithTimeout, "database.connect", "mongo_options_invalid").
			Public("database connection configuration is invalid").
			Wrapf(err, "validate MongoDB client options")
		observability.RecordError(span, wrapped)
		return nil, wrapped
	}

	client, err := mongo.Connect(ctxWithTimeout, clientOpts)
	if err != nil {
		wrapped := observability.Builder(ctxWithTimeout, "database.connect", "mongo_connect_failed").
			Public("database connection failed").
			Wrapf(err, "connect to MongoDB database %q", db.Cfg.MongoDB)
		observability.RecordError(span, wrapped)
		return nil, wrapped
	}

	if err := client.Ping(ctxWithTimeout, nil); err != nil {
		_ = client.Disconnect(ctxWithTimeout)
		wrapped := observability.Builder(ctxWithTimeout, "database.connect", "mongo_ping_failed").
			Public("database connection check failed").
			Wrapf(err, "ping MongoDB database %q after connect", db.Cfg.MongoDB)
		observability.RecordError(span, wrapped)
		return nil, wrapped
	}

	return client, nil
}

// Close disconnects the MongoDB client.
func (db *Database) Close(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "database.close")
	defer span.End()

	if db == nil || db.DB == nil {
		return nil
	}

	client := db.DB.Client()
	if client == nil {
		return nil
	}

	if err := client.Disconnect(ctx); err != nil {
		wrapped := observability.Builder(ctx, "database.close", "mongo_disconnect_failed").
			Public("database shutdown failed").
			Wrap(err)
		observability.RecordError(span, wrapped)
		return wrapped
	}

	return nil
}

// Ping verifies that the MongoDB client is still reachable.
func (db *Database) Ping(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "database.ping")
	defer span.End()

	if db == nil {
		err := observability.Builder(ctx, "database.ping", "database_not_configured").
			Public("database is not available").
			New("database not configured")
		observability.RecordError(span, err)
		return err
	}
	if db.DB == nil {
		err := observability.Builder(ctx, "database.ping", "database_handle_missing").
			Public("database is not available").
			New("mongo database handle is nil")
		observability.RecordError(span, err)
		return err
	}

	client := db.DB.Client()
	if client == nil {
		err := observability.Builder(ctx, "database.ping", "mongo_client_missing").
			Public("database is not available").
			New("mongo client is nil")
		observability.RecordError(span, err)
		return err
	}

	err := client.Ping(ctx, nil)
	if err == nil {
		return nil
	}

	wrapped := observability.Builder(ctx, "database.ping", "mongo_ping_failed").
		Public("database health check failed").
		Wrap(err)
	observability.RecordError(span, wrapped)
	return wrapped
}
