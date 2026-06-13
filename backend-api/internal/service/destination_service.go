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
	ID         int    `json:"id"`
	Name       string `json:"name"`
	RTMPURL    string `json:"rtmp_url"`
	Configured bool   `json:"configured"`
	Enabled    int    `json:"enabled"`      // 1=true, 0=false — runtime toggle
	IsActive   bool   `json:"is_active"`    // worker currently running
	StreamKey  string `json:"stream_key"`
	BitLimited int    `json:"bit_limited"`  // 1=limit to 10000kbps, 0=no limit
}

func (s *DestinationService) GetDestinations(ctx context.Context, userID int) ([]destinationResponse, error) {
	destinations, err := s.userRepo.GetDestinationsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]destinationResponse, len(destinations))
	for i, d := range destinations {
		// Decrypt stream key for display in UI
		var decryptedKey string
		if d.StreamKey != nil && *d.StreamKey != "" {
			decryptedKey, err = s.encryptor.Decrypt(*d.StreamKey)
			if err != nil {
				decryptedKey = "" // Don't expose encrypted key on error
			}
		}

		configured := d.StreamKey != nil && *d.StreamKey != ""

		result[i] = destinationResponse{
			ID:         d.ID,
			Name:       d.Destination.Name,
			RTMPURL:    d.Destination.RTMPURL,
			Configured: configured,
			Enabled:    d.Enabled,
			IsActive:   configured && d.Enabled == 1 && s.workerMgr.IsWorkerRunning(userID, d.ID),
			StreamKey:  decryptedKey,
			BitLimited: d.BitLimited,
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
	wasConfigured := oldStreamKey != nil && *oldStreamKey != ""

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

	// Auto-enable when saving a stream key
	if streamKey != "" && dest.Enabled != 1 {
		_ = s.userRepo.UpdateDestinationEnabled(ctx, userID, destinationID, 1)
	}

	isLive := s.streamChecker(userID)
	if !isLive {
		return streamKey != "", nil
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

func (s *DestinationService) UpdateBitLimited(ctx context.Context, userID, destinationID int, bitLimited int) error {
	if bitLimited != 0 && bitLimited != 1 {
		return ErrInvalidInput
	}

	dest, err := s.userRepo.GetDestinationByID(ctx, userID, destinationID)
	if err != nil {
		return err
	}
	if dest == nil {
		return ErrInvalidCredentials
	}

	if err := s.userRepo.UpdateDestinationBitLimited(ctx, userID, destinationID, bitLimited); err != nil {
		return err
	}

	// Restart worker if live to apply new bitrate limit
	isLive := s.streamChecker(userID)
	if !isLive {
		return nil
	}

	hasKey := dest.StreamKey != nil && *dest.StreamKey != ""

	if bitLimited == 1 && dest.Enabled == 1 && hasKey {
		if err := s.workerMgr.StopWorker(ctx, userID, destinationID); err != nil {
			return err
		}
		if err := s.workerMgr.StartWorker(ctx, userID, destinationID); err != nil {
			return err
		}
	} else if bitLimited == 0 && dest.Enabled == 1 && hasKey {
		if err := s.workerMgr.StopWorker(ctx, userID, destinationID); err != nil {
			return err
		}
		if err := s.workerMgr.StartWorker(ctx, userID, destinationID); err != nil {
			return err
		}
	}

	return nil
}

func (s *DestinationService) ToggleDestination(ctx context.Context, userID, destinationID int, enabled int) error {
	if enabled != 0 && enabled != 1 {
		return ErrInvalidInput
	}

	dest, err := s.userRepo.GetDestinationByID(ctx, userID, destinationID)
	if err != nil {
		return err
	}
	if dest == nil {
		return ErrInvalidCredentials
	}

	if err := s.userRepo.UpdateDestinationEnabled(ctx, userID, destinationID, enabled); err != nil {
		return err
	}

	isLive := s.streamChecker(userID)
	if !isLive {
		return nil
	}

	hasKey := dest.StreamKey != nil && *dest.StreamKey != ""

	if enabled == 1 && hasKey {
		if err := s.workerMgr.StartWorker(ctx, userID, destinationID); err != nil {
			return err
		}
	} else if enabled == 0 {
		if err := s.workerMgr.StopWorker(ctx, userID, destinationID); err != nil {
			return err
		}
	}

	return nil
}