package events

import (
	"drblury/poc-event-signup/internal/domain"
	"time"
)

type demoEvent struct {
	ID   int          `json:"id"`
	Date *domain.Date `json:"date"`
}

type processedDemoEvent struct {
	ProcessedID int          `json:"processed_id"`
	Time        time.Time    `json:"time"`
	Date        *domain.Date `json:"date"`
}

type signupEvent struct {
	ID     int            `json:"id"`
	Signup *domain.Signup `json:"signup"`
}

type processedSignupEvent struct {
	ProcessedID    int       `json:"processed_id"`
	Time           time.Time `json:"time"`
	SuccessMessage string    `json:"success_message"`
	ErrorMessage   string    `json:"error_message"`
}
