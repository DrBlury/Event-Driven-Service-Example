package database

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"drblury/poc-event-signup/internal/domain"
)

const signupCollection = "signups"

func (db *Database) StoreIncomingMessage(ctx context.Context, topic string, uuid string, payload string) error {
	_, err := db.DB.Collection(topic).InsertOne(ctx, bson.M{
		"topic":   topic,
		"uuid":    uuid,
		"payload": payload,
	})
	return err
}

func (db *Database) SetIncomingMessageProcessed(ctx context.Context, topic string, uuid string) error {
	_, err := db.DB.Collection(topic).UpdateOne(ctx, bson.M{"uuid": uuid}, bson.M{"$set": bson.M{"processed": true}})
	return err
}

func (db *Database) StoreOutgoingMessage(ctx context.Context, topic string, uuid string, payload string) error {
	_, err := db.DB.Collection(topic+"_outbox").InsertOne(ctx, bson.M{
		"topic":   topic,
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
