package domain

import (
	"errors"
	"fmt"
	"strings"

	"github.com/samber/lo"
)

// Predefined errors used throughout the application
var (
	ErrorNotFound        = errors.New("not found")
	ErrorBadRequest      = errors.New("something you provided was wrong")
	ErrorUpstreamService = errors.New("upstream service error")
	ErrorNotImplemented  = errors.New("not implemented")
	ErrorInternal        = errors.New("internal error")
)

type ErrValidations struct {
	Errors []string
}

func (e ErrValidations) Error() string {
	// Join the errors with a separator and number them
	numbered := lo.Map(e.Errors, func(v string, i int) string {
		return fmt.Sprintf("%d: %s", i+1, v)
	})
	return strings.Join(numbered, " - ")
}
