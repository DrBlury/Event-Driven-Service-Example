package usecase

import (
	"drblury/poc-event-signup/internal/domain"
	"fmt"
	"log/slog"

	"buf.build/go/protovalidate"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type AppLogic struct {
	log       *slog.Logger
	validator protovalidate.Validator
}

func NewAppLogic(
	logger *slog.Logger,
) AppLogic {
	v, err := protovalidate.New()
	if err != nil {
		slog.With("error", err).Error("failed to create validator")
	}

	return AppLogic{
		log:       logger,
		validator: v,
	}
}

func (a AppLogic) validate(msg protoreflect.ProtoMessage) error {
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
