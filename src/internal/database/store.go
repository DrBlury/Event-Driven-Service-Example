package database

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"drblury/event-driven-service/internal/domain"
)

const signupCollection = "signups"

func (db *Database) StoreOutgoingMessage(ctx context.Context, handler string, uuid string, payload string) error {
	_, err := db.DB.Collection(handler+"_outbox").InsertOne(ctx, bson.M{
		"handler": handler,
		"uuid":    uuid,
		"payload": payload,
	})
	return err
}

func (db *Database) StoreSignupInfo(ctx context.Context, info *domain.SignupInfo) error {
	_, err := db.DB.Collection(signupCollection).InsertOne(ctx, info)
	return err
}

func (db *Database) GetSignupInfoByID(ctx context.Context, id string) (*domain.SignupInfo, error) {
	var result domain.SignupInfo
	err := db.DB.Collection(signupCollection).FindOne(ctx, bson.M{"signupid": id}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, errors.New("signupInfo not found")
	}
	return &result, err
}
