package service

import (
	"context"
	"fmt"
	"sort"
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

type BroadcastGroupResponse struct {
	DestinationName string `json:"destination_name"`
	Date            string `json:"date"`
	TotalMinutes    int    `json:"total_minutes"`
}

func (s *BroadcastService) DeleteBroadcastGroup(ctx context.Context, userID int, destinationName, dateStr string) (int64, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return 0, fmt.Errorf("invalid date format: %w", err)
	}

	return s.broadcastRepo.DeleteByDestinationAndDate(ctx, userID, destinationName, date)
}

func (s *BroadcastService) GetBroadcasts(ctx context.Context, userID int) ([]BroadcastGroupResponse, error) {
	sessions, err := s.broadcastRepo.GetByUserIDWithDestination(ctx, userID)
	if err != nil {
		return nil, err
	}

	today := toEst(time.Now()).Format("2006-01-02")
	groups := make(map[string]*BroadcastGroupResponse)

	for _, session := range sessions {
		dateStr := toEst(session.Date).Format("2006-01-02")
		if dateStr == today {
			continue
		}

		key := session.DestinationName + "|" + dateStr
		if g, ok := groups[key]; ok {
			g.TotalMinutes += session.DurationMinutes
		} else {
			groups[key] = &BroadcastGroupResponse{
				DestinationName: session.DestinationName,
				Date:            dateStr,
				TotalMinutes:    session.DurationMinutes,
			}
		}
	}

	result := make([]BroadcastGroupResponse, 0, len(groups))
	for _, g := range groups {
		result = append(result, *g)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Date > result[j].Date
	})

	return result, nil
}