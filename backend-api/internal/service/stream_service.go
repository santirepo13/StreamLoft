package service

import (
	"context"
	"fmt"
	"log"
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
	// Log callback verification
	log.Printf("STREAM_CALLBACK: Received /stream/start callback with stream_key=%s, detected_bitrate=%d", streamKey, detectedBitrate)
	
	user, err := s.userRepo.GetByStreamKey(ctx, streamKey)
	if err != nil {
		log.Printf("STREAM_CALLBACK_ERROR: Failed to get user by stream_key=%s: %v", streamKey, err)
		return err
	}
	if user == nil {
		log.Printf("STREAM_CALLBACK_ERROR: User not found for stream_key=%s", streamKey)
		return fmt.Errorf("user not found for stream key")
	}
	
	log.Printf("STREAM_CALLBACK_SUCCESS: Found user_id=%d, name=%s for stream_key=%s", user.ID, user.Name, streamKey)

	bitrateWarning := false
	if user.Bitrate != nil && detectedBitrate > 0 {
		threshold := float64(*user.Bitrate) * 0.7
		if float64(detectedBitrate) < threshold {
			bitrateWarning = true
			log.Printf("STREAM_CALLBACK_BITRATE_WARNING: User %d detected bitrate %d < threshold %f (configured: %d)", 
				user.ID, detectedBitrate, threshold, *user.Bitrate)
		}
	}

	s.liveUsers.Store(streamKey, LiveStreamStatus{
		UserID:         user.ID,
		StreamKey:      streamKey,
		BitrateWarning: bitrateWarning,
		StartTime:      time.Now(),
	})

	log.Printf("STREAM_CALLBACK_USER_LIVE: Marked user %d as live with stream_key=%s, bitrate_warning=%t", 
		user.ID, streamKey, bitrateWarning)

	destinations, err := s.userRepo.GetDestinationsWithStreamKey(ctx, user.ID)
	if err != nil {
		log.Printf("STREAM_CALLBACK_ERROR: Failed to get destinations for user %d: %v", user.ID, err)
		return err
	}

	log.Printf("STREAM_CALLBACK_DESTINATIONS: Found %d destinations with stream keys for user %d", len(destinations), user.ID)
	
	for _, dest := range destinations {
		_, err := s.broadcastRepo.Create(ctx, user.ID, dest.ID)
		if err != nil {
			log.Printf("STREAM_CALLBACK_BROADCAST_ERROR: Failed to create broadcast session for user %d, destination %d: %v", 
				user.ID, dest.ID, err)
			continue
		}

		if err := s.workerSvc.StartWorker(ctx, user.ID, dest.ID); err != nil {
			log.Printf("STREAM_CALLBACK_WORKER_ERROR: Failed to start worker for user %d, destination %d: %v", 
				user.ID, dest.ID, err)
			continue
		}
		
		log.Printf("STREAM_CALLBACK_WORKER_SUCCESS: Started forwarding worker for user %d, destination %d (%s)", 
			user.ID, dest.ID, dest.Destination.Name)
	}

	log.Printf("STREAM_CALLBACK_COMPLETE: Successfully processed stream start for user %d, started %d forwarding workers", 
		user.ID, len(destinations))
	
	return nil
}

func (s *StreamService) StopStream(ctx context.Context, streamKey string) error {
	log.Printf("STREAM_STOP: Received stream stop for stream_key=%s", streamKey)
	
	user, err := s.userRepo.GetByStreamKey(ctx, streamKey)
	if err != nil {
		log.Printf("STREAM_STOP_ERROR: Failed to get user by stream_key=%s: %v", streamKey, err)
		return err
	}
	if user == nil {
		log.Printf("STREAM_STOP_WARNING: User not found for stream_key=%s (may have already been cleaned up)", streamKey)
		return nil
	}

	log.Printf("STREAM_STOP_SUCCESS: Found user_id=%d for stream_key=%s", user.ID, streamKey)

	s.liveUsers.Delete(streamKey)
	log.Printf("STREAM_STOP_USER_LIVE: Removed user %d from live streams", user.ID)

	s.workerSvc.StopAllForUser(ctx, user.ID)
	log.Printf("STREAM_STOP_WORKERS: Stopped all workers for user %d", user.ID)

	sessions, err := s.broadcastRepo.GetActiveByUserID(ctx, user.ID)
	if err != nil {
		log.Printf("STREAM_STOP_ERROR: Failed to get active sessions for user %d: %v", user.ID, err)
		return err
	}

	log.Printf("STREAM_STOP_SESSIONS: Found %d active sessions for user %d to close", len(sessions), user.ID)
	
	now := time.Now()
	for _, session := range sessions {
		duration := int(now.Sub(session.StartedAt).Minutes())
		err := s.broadcastRepo.UpdateEndTime(ctx, session.ID, now, duration)
		if err != nil {
			log.Printf("STREAM_STOP_SESSION_ERROR: Failed to update session %d: %v", session.ID, err)
			continue
		}
		log.Printf("STREAM_STOP_SESSION_SUCCESS: Closed session %d with duration %d minutes", session.ID, duration)
	}

	log.Printf("STREAM_STOP_COMPLETE: Successfully processed stream stop for user %d", user.ID)
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