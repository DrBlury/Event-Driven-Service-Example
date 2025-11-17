package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"drblury/event-driven-service/internal/database"
	"drblury/event-driven-service/internal/domain"
	"drblury/event-driven-service/pkg/events"

	"buf.build/go/protovalidate"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type AppLogic struct {
	db            *database.Database
	log           *slog.Logger
	validator     protovalidate.Validator
	eventProducer events.Producer
	signupTopic   string
}

func NewAppLogic(
	db *database.Database,
	logger *slog.Logger,
	eventsCfg *events.Config,
) *AppLogic {
	v, err := protovalidate.New()
	if err != nil {
		slog.With("error", err).Error("failed to create validator")
	}

	signupTopic := ""
	if eventsCfg != nil {
		signupTopic = eventsCfg.ConsumeQueueSignup
	}

	return &AppLogic{
		db:          db,
		log:         logger,
		validator:   v,
		signupTopic: signupTopic,
	}
}

// SetEventProducer wires the event producer used by PublishEvent.
func (a *AppLogic) SetEventProducer(producer events.Producer) {
	if a == nil {
		return
	}
	a.eventProducer = producer
}

// PublishEvent emits the supplied payload to the configured topic with optional metadata.
func (a *AppLogic) PublishEvent(ctx context.Context, topic string, payload proto.Message, metadata events.Metadata) error {
	if a == nil {
		return errors.New("applogic is nil")
	}
	if payload == nil {
		return errors.New("event payload is required")
	}
	if topic == "" {
		return errors.New("event topic is required")
	}
	if a.eventProducer == nil {
		return errors.New("event producer is not configured")
	}

	return a.eventProducer.PublishProto(ctx, topic, payload, metadata)
}

func (a *AppLogic) SignupTopic() string {
	if a == nil {
		return ""
	}
	return a.signupTopic
}

func (a AppLogic) Validate(msg protoreflect.ProtoMessage) error {
	if err := a.validator.Validate(msg); err != nil {
		// log the error
		slog.With("error", err).Error("validation error")
		var errMessages []string
		for _, violation := range err.(*protovalidate.ValidationError).Violations {
			errMessage := fmt.Sprintf("%s %s", violation.Proto.GetField(), violation.Proto.GetMessage())
			errMessages = append(errMessages, errMessage)
		}
		return domain.ErrValidations{Errors: errMessages}
	}
	return nil
}

// DatabaseProbe ensures the backing database remains reachable for readiness checks.
func (a *AppLogic) DatabaseProbe(ctx context.Context) error {
	if a == nil {
		return errors.New("applogic is nil")
	}
	if a.db == nil {
		return errors.New("database not configured")
	}
	return a.db.Ping(ctx)
}
