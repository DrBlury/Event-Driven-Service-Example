package database

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"drblury/event-driven-service/internal/domain"
)

const exampleCollection = "example-records"

func (db *Database) StoreOutgoingMessage(ctx context.Context, handler string, uuid string, payload string) error {
	_, err := db.DB.Collection(handler+"_outbox").InsertOne(ctx, bson.M{
		"handler": handler,
		"uuid":    uuid,
		"payload": payload,
	})
	return err
}

func (db *Database) StoreExampleRecord(ctx context.Context, record *domain.ExampleRecord) error {
	if record == nil {
		return errors.New("example record is required")
	}
	_, err := db.DB.Collection(exampleCollection).InsertOne(ctx, record)
	return err
}

func (db *Database) GetExampleRecordByID(ctx context.Context, id string) (*domain.ExampleRecord, error) {
	var result domain.ExampleRecord
	err := db.DB.Collection(exampleCollection).FindOne(ctx, bson.M{"record_id": id}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, errors.New("example record not found")
	}
	return &result, err
}
