package service

import (
	"context"

	"streamloft-api/internal/crypto"
	"streamloft-api/internal/interfaces"
)

type DestinationService struct {
	userRepo     interfaces.UserRepository
	workerMgr    interfaces.WorkerManager
	streamChecker func(userID int) bool
	encryptor    *crypto.Encryptor
}

func NewDestinationService(
	userRepo interfaces.UserRepository,
	workerMgr interfaces.WorkerManager,
	streamChecker func(userID int) bool,
	enc *crypto.Encryptor,
) *DestinationService {
	return &DestinationService{
		userRepo:      userRepo,
		workerMgr:    workerMgr,
		streamChecker: streamChecker,
		encryptor:    enc,
	}
}

type destinationResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	RTMPURL     string `json:"rtmp_url"`
	Configured  bool   `json:"configured"`
}

func (s *DestinationService) GetDestinations(ctx context.Context, userID int) ([]destinationResponse, error) {
	destinations, err := s.userRepo.GetDestinationsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]destinationResponse, len(destinations))
	for i, d := range destinations {
		result[i] = destinationResponse{
			ID:         d.ID,
			Name:       d.Destination.Name,
			RTMPURL:    d.Destination.RTMPURL,
			Configured: d.StreamKey != "",
		}
	}

	return result, nil
}

func (s *DestinationService) UpdateStreamKey(ctx context.Context, userID, destinationID int, streamKey string) (bool, error) {
	dest, err := s.userRepo.GetDestinationByID(ctx, userID, destinationID)
	if err != nil {
		return false, err
	}
	if dest == nil {
		return false, ErrInvalidCredentials
	}

	oldStreamKey := dest.StreamKey
	wasConfigured := oldStreamKey != ""

	if streamKey != "" {
		encrypted, err := s.encryptor.Encrypt(streamKey)
		if err != nil {
			return false, err
		}
		streamKey = encrypted
	}

	if err := s.userRepo.UpdateDestinationStreamKey(ctx, userID, destinationID, streamKey); err != nil {
		return false, err
	}

	isLive := s.streamChecker(userID)
	if !isLive {
		return wasConfigured, nil
	}

	if streamKey != "" && !wasConfigured {
		if err := s.workerMgr.StartWorker(ctx, userID, destinationID); err != nil {
			return false, err
		}
	} else if streamKey == "" && wasConfigured {
		if err := s.workerMgr.StopWorker(ctx, userID, destinationID); err != nil {
			return false, err
		}
	}

	return streamKey != "", nil
}