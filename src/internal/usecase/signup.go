package usecase

import (
	"context"
	"drblury/poc-event-signup/internal/domain"
)

func (a AppLogic) Signup(ctx context.Context, signup *domain.Signup, token string) error {
	// validate signup
	err := a.validate(signup)
	if err != nil {
		return err
	}

	// Do something with the signup
	// store signup in database

	return nil
}
