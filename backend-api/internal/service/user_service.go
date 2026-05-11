package service

import (
	"context"

	"streamloft-api/internal/interfaces"
)

type UserService struct {
	userRepo interfaces.UserRepository
}

func NewUserService(userRepo interfaces.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

type UserResponse struct {
	ID          int                      `json:"id"`
	NumericID   string                   `json:"numeric_id"`
	Name        string                  `json:"name"`
	RTMPURL     string                  `json:"rtmp_url"`
	StreamKey   *string                  `json:"stream_key"`
	Destinations []DestinationResponse    `json:"destinations"`
}

type DestinationResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	RTMPURL     string `json:"rtmp_url"`
	Configured  bool   `json:"configured"` // true if stream_key is set
}

func (s *UserService) GetUser(ctx context.Context, userID int, rtmpURL string) (*UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	destinations, err := s.userRepo.GetDestinationsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	destResp := make([]DestinationResponse, len(destinations))
	for i, d := range destinations {
		destResp[i] = DestinationResponse{
			ID:         d.ID,
			Name:       d.Destination.Name,
			RTMPURL:    d.Destination.RTMPURL,
			Configured: d.StreamKey != nil && *d.StreamKey != "",
		}
	}

	return &UserResponse{
		ID:          user.ID,
		NumericID:   user.NumericID,
		Name:        user.Name,
		RTMPURL:     rtmpURL,
		StreamKey:   user.StreamKey,
		Destinations: destResp,
	}, nil
}

func (s *UserService) UpdateBitrate(ctx context.Context, userID int, bitrate int) error {
	return s.userRepo.UpdateBitrate(ctx, userID, bitrate)
}