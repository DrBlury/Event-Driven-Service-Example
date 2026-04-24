package database

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	DB  *mongo.Database
	Cfg *Config
}

func NewDatabase(cfg *Config, logger *slog.Logger, ctx context.Context) (*Database, error) {
	db := &Database{Cfg: cfg}

	if err := db.Connect(ctx, logger); err != nil {
		return nil, err
	}

	return db, nil
}

// Connect initializes the MongoDB client and verifies connectivity.
func (db *Database) Connect(ctx context.Context, logger *slog.Logger) error {
	if db == nil {
		return errors.New("database not configured")
	}
	if db.Cfg == nil {
		return errors.New("database configuration is required")
	}
	if db.DB != nil {
		return nil
	}

	log := logger
	if log == nil {
		log = slog.Default()
	}

	log.Info("Connecting to MongoDB", "url", db.Cfg.MongoURL, "db", db.Cfg.MongoDB, "user", db.Cfg.MongoUser)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(db.Cfg.MongoURL)
	clientOpts.Auth = &options.Credential{
		Username: db.Cfg.MongoUser,
		Password: db.Cfg.MongoPassword,
	}

	if err := clientOpts.Validate(); err != nil {
		log.Error("MongoDB client options validation failed", "error", err)
		return err
	}

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Error("MongoDB connection failed", "error", err)
		return err
	}
	if err := client.Ping(ctx, nil); err != nil {
		log.Error("MongoDB ping failed", "error", err)
		_ = client.Disconnect(ctx)
		return err
	}

	db.DB = client.Database(db.Cfg.MongoDB)
	log.Info("Connected to MongoDB successfully")

	return nil
}

// Close disconnects the MongoDB client.
func (db *Database) Close(ctx context.Context) error {
	if db == nil || db.DB == nil {
		return nil
	}

	client := db.DB.Client()
	if client == nil {
		return nil
	}

	return client.Disconnect(ctx)
}

// Ping verifies that the MongoDB client is still reachable.
func (db *Database) Ping(ctx context.Context) error {
	if db == nil {
		return errors.New("database not configured")
	}
	if db.DB == nil {
		return errors.New("mongo database handle is nil")
	}

	client := db.DB.Client()
	if client == nil {
		return errors.New("mongo client is nil")
	}

	return client.Ping(ctx, nil)
}
