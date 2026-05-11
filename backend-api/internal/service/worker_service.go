package service

import (
	"context"

	"streamloft-api/internal/crypto"
	"streamloft-api/internal/interfaces"
	"streamloft-api/internal/workers"
)

type WorkerService struct {
	userRepo     interfaces.UserRepository
	forwardingMgr *workers.ForwardingManager
	encryptor    *crypto.Encryptor
	srsURL       string
}

func NewWorkerService(
	userRepo interfaces.UserRepository,
	enc *crypto.Encryptor,
	srsURL string,
) *WorkerService {
	return &WorkerService{
		userRepo:     userRepo,
		forwardingMgr: workers.NewForwardingManager(srsURL),
		encryptor:    enc,
		srsURL:       srsURL,
	}
}

func (s *WorkerService) StartWorker(ctx context.Context, userID, userDestinationID int) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil || user.StreamKey == "" {
		return nil
	}

	destinations, err := s.userRepo.GetDestinationsWithStreamKey(ctx, userID)
	if err != nil {
		return err
	}

	for _, dest := range destinations {
		if dest.ID == userDestinationID && dest.StreamKey != "" {
			decryptedKey, err := s.encryptor.Decrypt(dest.StreamKey)
			if err != nil {
				return err
			}

			targetURL := dest.Destination.RTMPURL + "/" + decryptedKey
			return s.forwardingMgr.StartWorker(ctx, userID, userDestinationID, user.StreamKey, targetURL)
		}
	}

	return nil
}

func (s *WorkerService) StopWorker(ctx context.Context, userID, userDestinationID int) error {
	return s.forwardingMgr.StopWorker(ctx, userID, userDestinationID)
}

func (s *WorkerService) StopAllForUser(ctx context.Context, userID int) error {
	return s.forwardingMgr.StopAllForUser(ctx, userID)
}

func (s *WorkerService) IsWorkerRunning(userID, userDestinationID int) bool {
	return s.forwardingMgr.IsWorkerRunning(userID, userDestinationID)
}

var _ interfaces.WorkerManager = (*WorkerService)(nil)