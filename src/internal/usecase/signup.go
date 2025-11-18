package usecase

import (
	"context"
	"errors"

	"drblury/event-driven-service/internal/domain"

	"github.com/drblury/protoflow"
)

func (a *AppLogic) Signup(ctx context.Context, signup *domain.Signup, token string) error {
	// Do something with the signup
	// store signup in database

	return nil
}

func (a *AppLogic) EmitSignupEvent(ctx context.Context, signup *domain.Signup) error {
	if a == nil {
		return errors.New("applogic is nil")
	}
	if signup == nil {
		return errors.New("signup payload is required")
	}
	if a.signupTopic == "" {
		return errors.New("signup topic not configured")
	}

	metadata := protoflow.Metadata{
		"source": "api.signup",
	}

	return a.eventProducer.PublishProto(ctx, a.signupTopic, signup, metadata)
}
