package events

import (
	"fmt"
	"log/slog"

	"drblury/event-driven-service/internal/domain"
	"drblury/event-driven-service/internal/observability"

	"buf.build/go/protovalidate"
	"github.com/samber/lo"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Validator struct {
	validator protovalidate.Validator
}

func NewValidator() (*Validator, error) {
	validator, err := protovalidate.New()
	if err != nil {
		wrapped := observability.Builder(nil, "events.validator", "protovalidate_init_failed").
			Public("event validation could not be initialized").
			Wrap(err)
		slog.Error("error creating protovalidate validator", "error", wrapped)
		return nil, wrapped
	}
	return &Validator{
		validator: validator,
	}, nil
}

func (v Validator) Validate(a any) error {
	if err := v.validator.Validate(a.(protoreflect.ProtoMessage)); err != nil {
		errMessages := lo.Map(err.(*protovalidate.ValidationError).Violations, func(violation *protovalidate.Violation, _ int) string {
			return fmt.Sprintf("%s %s", violation.Proto.GetField(), violation.Proto.GetMessage())
		})
		slog.Error(
			"validation error",
			"error",
			observability.Builder(nil, "events.validator", "event_validation_failed").
				With("violations", errMessages).
				Wrap(err),
		)
		return domain.ErrValidations{Errors: errMessages}
	}
	return nil
}
