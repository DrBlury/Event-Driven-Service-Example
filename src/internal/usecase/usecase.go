package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"drblury/event-driven-service/internal/database"
	"drblury/event-driven-service/internal/domain"

	"buf.build/go/protovalidate"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type AppLogic struct {
	db        *database.Database
	log       *slog.Logger
	validator protovalidate.Validator
}

func NewAppLogic(
	db *database.Database,
	logger *slog.Logger,
) *AppLogic {
	v, err := protovalidate.New()
	if err != nil {
		slog.With("error", err).Error("failed to create validator")
	}

	return &AppLogic{
		db:        db,
		log:       logger,
		validator: v,
	}
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
