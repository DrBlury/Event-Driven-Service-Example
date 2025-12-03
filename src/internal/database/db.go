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
	logger.Info("Connecting to MongoDB", "url", cfg.MongoURL, "db", cfg.MongoDB, "user", cfg.MongoUser)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(cfg.MongoURL)

	clientOpts.Auth = &options.Credential{
		Username: cfg.MongoUser,
		Password: cfg.MongoPassword,
	}

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		logger.Error("MongoDB connection failed", "error", err)
		return nil, err
	}
	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		logger.Error("MongoDB ping failed", "error", err)
		return nil, err
	}
	logger.Info("Connected to MongoDB successfully")
	return &Database{
		DB:  client.Database(cfg.MongoDB),
		Cfg: cfg,
	}, nil
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
