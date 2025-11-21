package events

import (
	"drblury/event-driven-service/internal/domain"
	"fmt"
	"log/slog"

	"buf.build/go/protovalidate"
	"github.com/bytedance/gopkg/util/logger"
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
		var errMessages []string
		for _, violation := range err.(*protovalidate.ValidationError).Violations {
			errMessage := fmt.Sprintf("%s %s", violation.Proto.GetField(), violation.Proto.GetMessage())
			errMessages = append(errMessages, errMessage)
		}
		return domain.ErrValidations{Errors: errMessages}
	}
	return nil
}
