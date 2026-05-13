package database

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/attribute"

	"drblury/event-driven-service/internal/domain"
	"drblury/event-driven-service/internal/observability"
)

const exampleCollection = "example-records"

func (db *Database) StoreOutgoingMessage(ctx context.Context, handler string, uuid string, payload string) error {
	ctx, span := tracer.Start(ctx, "database.store_outgoing_message")
	defer span.End()

	collection := handler + "_outbox"
	span.SetAttributes(
		attribute.String("db.collection.name", collection),
		attribute.String("messaging.handler", handler),
		attribute.String("messaging.message.id", uuid),
	)

	if db == nil || db.DB == nil {
		err := observability.Builder(ctx, "database.store_outgoing_message", "database_not_configured").
			Public("database is not available").
			New("mongo database handle is nil")
		observability.RecordError(span, err)
		return err
	}

	_, err := db.DB.Collection(collection).InsertOne(ctx, bson.M{
		"handler": handler,
		"uuid":    uuid,
		"payload": payload,
	})
	if err == nil {
		return nil
	}

	wrapped := observability.Builder(ctx, "database.store_outgoing_message", "mongo_insert_failed").
		Public("outgoing message could not be stored").
		With("collection", collection, "handler", handler, "message_uuid", uuid).
		Wrap(err)
	observability.RecordError(span, wrapped)
	return wrapped
}

func (db *Database) StoreExampleRecord(ctx context.Context, record *domain.ExampleRecord) error {
	ctx, span := tracer.Start(ctx, "database.store_example_record")
	defer span.End()

	if record == nil {
		err := observability.Builder(ctx, "database.store_example_record", "example_record_required").
			Public("example payload is invalid").
			New("example record is required")
		observability.RecordError(span, err)
		return err
	}

	span.SetAttributes(
		attribute.String("db.collection.name", exampleCollection),
		attribute.String("example.record_id", record.GetRecordId()),
	)

	if db == nil || db.DB == nil {
		err := observability.Builder(ctx, "database.store_example_record", "database_not_configured").
			Public("database is not available").
			With("record_id", record.GetRecordId()).
			New("mongo database handle is nil")
		observability.RecordError(span, err)
		return err
	}

	_, err := db.DB.Collection(exampleCollection).InsertOne(ctx, record)
	if err == nil {
		return nil
	}

	wrapped := observability.Builder(ctx, "database.store_example_record", "mongo_insert_failed").
		Public("example record could not be stored").
		With("collection", exampleCollection, "record_id", record.GetRecordId()).
		Wrap(err)
	observability.RecordError(span, wrapped)
	return wrapped
}

func (db *Database) GetExampleRecordByID(ctx context.Context, id string) (*domain.ExampleRecord, error) {
	ctx, span := tracer.Start(ctx, "database.get_example_record_by_id")
	defer span.End()

	var result domain.ExampleRecord
	span.SetAttributes(
		attribute.String("db.collection.name", exampleCollection),
		attribute.String("example.record_id", id),
	)

	if db == nil || db.DB == nil {
		err := observability.Builder(ctx, "database.get_example_record_by_id", "database_not_configured").
			Public("database is not available").
			With("record_id", id).
			New("mongo database handle is nil")
		observability.RecordError(span, err)
		return nil, err
	}

	err := db.DB.Collection(exampleCollection).FindOne(ctx, bson.M{"record_id": id}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		wrapped := observability.Builder(ctx, "database.get_example_record_by_id", "example_record_not_found").
			Public("example record was not found").
			With("record_id", id).
			Wrapf(domain.ErrorNotFound, "example record %q not found", id)
		observability.RecordError(span, wrapped)
		return nil, wrapped
	}
	if err != nil {
		wrapped := observability.Builder(ctx, "database.get_example_record_by_id", "mongo_query_failed").
			Public("example record lookup failed").
			With("record_id", id, "collection", exampleCollection).
			Wrap(err)
		observability.RecordError(span, wrapped)
		return nil, wrapped
	}
	return &result, nil
}
