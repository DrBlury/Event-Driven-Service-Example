package domain

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrorNotFound           = errors.New("not found")
	ErrorMulti              = errors.New("something went wrong internally")
	ErrorBadRequest         = errors.New("something you provided was wrong")
	ErrorInvalidToken       = errors.New("invalid token")
	ErrorInvalidCredentials = errors.New("invalid credentials for login")
	ErrorUpstreamService    = errors.New("upstream service error")
	ErrorMissingParameter   = errors.New("missing parameter")
	ErrorMissingToken       = errors.New("missing token")
	ErrorTripicaBusiness    = errors.New("upstream business error")
	ErrorNotImplemented     = errors.New("not implemented")
	ErrorSignupFailed       = errors.New("signup failed")
)

type ErrSignupFailed struct {
	Message string
}

func (e ErrSignupFailed) Error() string {
	return fmt.Sprintf("%s: %s", ErrorSignupFailed.Error(), e.Message)
}

type ErrValidations struct {
	Errors []string
}

func (e ErrValidations) Error() string {
	// Join the errors with a separator and number them
	for i, v := range e.Errors {
		e.Errors[i] = fmt.Sprintf("%d: %s", i+1, v)
	}
	return strings.Join(e.Errors, " - ")
}

type ErrTripicaBusiness struct {
	Message string
	Details string
}

func (e ErrTripicaBusiness) Error() string {
	return fmt.Sprintf("%s: %s: %s", ErrorTripicaBusiness.Error(), e.Message, e.Details)
}
