package service

import (
	"context"
	"time"

	"streamloft-api/internal/interfaces"
)

var estLocation *time.Location

func init() {
	var err error
	estLocation, err = time.LoadLocation("America/New_York")
	if err != nil {
		estLocation = time.UTC
	}
}

func toEst(t time.Time) time.Time {
	return t.In(estLocation)
}

type BroadcastService struct {
	broadcastRepo interfaces.BroadcastSessionRepository
}

func NewBroadcastService(broadcastRepo interfaces.BroadcastSessionRepository) *BroadcastService {
	return &BroadcastService{
		broadcastRepo: broadcastRepo,
	}
}

type BroadcastResponse struct {
	ID                int        `json:"id"`
	DestinationName   string     `json:"destination_name"`
	Date              string     `json:"date"`
	StartedAt         string     `json:"started_at"`
	EndedAt           *string    `json:"ended_at,omitempty"`
	DurationMinutes  int        `json:"duration_minutes"`
}

func (s *BroadcastService) GetBroadcasts(ctx context.Context, userID int) ([]BroadcastResponse, error) {
	sessions, err := s.broadcastRepo.GetByUserIDWithDestination(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]BroadcastResponse, len(sessions))
	for i, session := range sessions {
		result[i] = BroadcastResponse{
			ID:               session.ID,
			DestinationName:  session.DestinationName,
			Date:             toEst(session.Date).Format("2006-01-02"),
			StartedAt:        toEst(session.StartedAt).Format("2006-01-02T15:04:05-07:00"),
			DurationMinutes:  session.DurationMinutes,
		}
		if session.EndedAt != nil {
			endedAt := toEst(*session.EndedAt).Format("2006-01-02T15:04:05-07:00")
			result[i].EndedAt = &endedAt
		}
	}

	return result, nil
}