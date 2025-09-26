package events

import (
	"drblury/poc-event-signup/internal/domain"
	"time"
)

type event struct {
	ID   int          `json:"id"`
	Date *domain.Date `json:"date"`
}

type processedEvent struct {
	ProcessedID int          `json:"processed_id"`
	Time        time.Time    `json:"time"`
	Date        *domain.Date `json:"date"`
}
