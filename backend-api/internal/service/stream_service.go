package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"streamloft-api/internal/interfaces"
)

type LiveStreamStatus struct {
	UserID         int
	StreamKey      string
	BitrateWarning bool
	StartTime      time.Time
}

type StreamService struct {
	userRepo     interfaces.UserRepository
	broadcastRepo interfaces.BroadcastSessionRepository
	workerSvc    interfaces.WorkerManager
	liveUsers    sync.Map
}

func NewStreamService(
	userRepo interfaces.UserRepository,
	broadcastRepo interfaces.BroadcastSessionRepository,
	workerSvc interfaces.WorkerManager,
) *StreamService {
	return &StreamService{
		userRepo:      userRepo,
		broadcastRepo: broadcastRepo,
		workerSvc:     workerSvc,
	}
}

type StreamStartRequest struct {
	StreamKey     string `json:"stream_key"`
	DetectedBitrate int  `json:"detected_bitrate"`
}

type StreamStatusResponse struct {
	Status        string `json:"status"`
	BitrateWarning bool   `json:"bitrate_warning"`
}

func (s *StreamService) StartStream(ctx context.Context, streamKey string, detectedBitrate int) error {
	user, err := s.userRepo.GetByStreamKey(ctx, streamKey)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found for stream key")
	}

	bitrateWarning := false
	if user.Bitrate != nil && detectedBitrate > 0 {
		threshold := float64(*user.Bitrate) * 0.7
		if float64(detectedBitrate) < threshold {
			bitrateWarning = true
		}
	}

	s.liveUsers.Store(streamKey, LiveStreamStatus{
		UserID:         user.ID,
		StreamKey:      streamKey,
		BitrateWarning: bitrateWarning,
		StartTime:      time.Now(),
	})

	destinations, err := s.userRepo.GetDestinationsWithStreamKey(ctx, user.ID)
	if err != nil {
		return err
	}

	for _, dest := range destinations {
		_, err := s.broadcastRepo.Create(ctx, user.ID, dest.ID)
		if err != nil {
			continue
		}

		if err := s.workerSvc.StartWorker(ctx, user.ID, dest.ID); err != nil {
			continue
		}
	}

	return nil
}

func (s *StreamService) StopStream(ctx context.Context, streamKey string) error {
	user, err := s.userRepo.GetByStreamKey(ctx, streamKey)
	if err != nil {
		return err
	}
	if user == nil {
		return nil
	}

	s.liveUsers.Delete(streamKey)

	s.workerSvc.StopAllForUser(ctx, user.ID)

	sessions, err := s.broadcastRepo.GetActiveByUserID(ctx, user.ID)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, session := range sessions {
		duration := int(now.Sub(session.StartedAt).Minutes())
		s.broadcastRepo.UpdateEndTime(ctx, session.ID, now, duration)
	}

	return nil
}

func (s *StreamService) GetStatus(ctx context.Context, userID int) (*StreamStatusResponse, error) {
	var status string
	var bitrateWarning bool

	s.liveUsers.Range(func(key, value interface{}) bool {
		liveStatus := value.(LiveStreamStatus)
		if liveStatus.UserID == userID {
			status = "live"
			bitrateWarning = liveStatus.BitrateWarning
			return false
		}
		return true
	})

	if status == "" {
		status = "offline"
	}

	return &StreamStatusResponse{
		Status:        status,
		BitrateWarning: bitrateWarning,
	}, nil
}

func (s *StreamService) IsUserLive(userID int) bool {
	var isLive bool
	s.liveUsers.Range(func(key, value interface{}) bool {
		liveStatus := value.(LiveStreamStatus)
		if liveStatus.UserID == userID {
			isLive = true
			return false
		}
		return true
	})
	return isLive
}

func (s *StreamService) GetLiveStatusForUser(userID int) *LiveStreamStatus {
	var status *LiveStreamStatus
	s.liveUsers.Range(func(key, value interface{}) bool {
		liveStatus := value.(LiveStreamStatus)
		if liveStatus.UserID == userID {
			status = &liveStatus
			return false
		}
		return true
	})
	return status
}