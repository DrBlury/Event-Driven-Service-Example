package usecase

import (
	"context"
	"errors"

	"drblury/event-driven-service/internal/domain"
)

func (a *AppLogic) Signup(ctx context.Context, signup *domain.Signup, token string) error {
	// validate signup
	err := a.Validate(signup)
	if err != nil {
		return err
	}

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

	metadata := map[string]string{
		"source": "api.signup",
	}

	return a.PublishEvent(ctx, a.signupTopic, signup, metadata)
}
