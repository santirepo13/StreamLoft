package interfaces

import (
	"context"
	"time"

	"streamloft-api/internal/models"
)

type BroadcastSessionRepository interface {
	Create(ctx context.Context, userID, userDestinationID int) (*models.BroadcastSession, error)
	UpdateEndTime(ctx context.Context, sessionID int, endedAt time.Time, durationMinutes int) error
	GetActiveByUserID(ctx context.Context, userID int) ([]models.BroadcastSession, error)
	GetByUserID(ctx context.Context, userID int) ([]models.BroadcastSession, error)
	GetByUserIDWithDestination(ctx context.Context, userID int) ([]BroadcastSessionWithDestination, error)
	DeleteByDestinationAndDate(ctx context.Context, userID int, destinationName string, date time.Time) (int64, error)
}

type BroadcastSessionWithDestination struct {
	ID                int        `json:"id"`
	UserID            int        `json:"user_id"`
	UserDestinationID int        `json:"user_destination_id"`
	DestinationName   string     `json:"destination_name"`
	Date              time.Time  `json:"date"`
	DurationMinutes  int        `json:"duration_minutes"`
	StartedAt         time.Time  `json:"started_at"`
	EndedAt           *time.Time `json:"ended_at,omitempty"`
}