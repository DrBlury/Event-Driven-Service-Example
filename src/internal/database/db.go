package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	DB  *mongo.Database
	Cfg *Config
}

func NewDatabase(cfg *Config, logger *slog.Logger) (*Database, error) {
	logger.Info("Connecting to MongoDB", "url", cfg.MongoURL, "db", cfg.MongoDB, "user", cfg.MongoUser)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoURI := cfg.MongoURL
	if cfg.MongoUser != "" && cfg.MongoPassword != "" {
		mongoURI = fmt.Sprintf("mongodb://%s:%s@%s", cfg.MongoUser, cfg.MongoPassword, cfg.MongoURL)
	}

	clientOpts := options.Client().ApplyURI(mongoURI)
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
