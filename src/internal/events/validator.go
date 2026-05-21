package events

import (
	"drblury/event-driven-service/internal/domain"
	"errors"
	"fmt"
	"log/slog"

	"buf.build/go/protovalidate"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/samber/lo"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Validator struct {
	validator protovalidate.Validator
}

func NewValidator() (*Validator, error) {
	validator, err := protovalidate.New()
	if err != nil {
		logger.Error("error creating protovalidate validator", err)
		return nil, err
	}
	return &Validator{
		validator: validator,
	}, nil
}

func (v Validator) Validate(a any) error {
	if err := v.validator.Validate(a.(protoreflect.ProtoMessage)); err != nil {
		// log the error
		slog.With("error", err).Error("validation error")

		var valErr *protovalidate.ValidationError
		if errors.As(err, &valErr) {
			errMessages := lo.Map(valErr.Violations, func(violation *protovalidate.Violation, _ int) string {
				return fmt.Sprintf("%s %s", violation.Proto.GetField(), violation.Proto.GetMessage())
			})
			return domain.ErrValidations{Errors: errMessages}
		}

		// CompilationError or any unexpected error type — surface it as-is
		return fmt.Errorf("validator setup error: %w", err)
	}
	return nil
}
