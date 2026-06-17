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

type StreamEvent struct {
	Type           string `json:"type"`
	Status         string `json:"status"`
	BitrateWarning bool   `json:"bitrate_warning"`
}

type StreamService struct {
	userRepo      interfaces.UserRepository
	broadcastRepo interfaces.BroadcastSessionRepository
	workerSvc     interfaces.WorkerManager
	liveUsers     sync.Map // streamKey → LiveStreamStatus
	userLiveKeys  sync.Map // userID → streamKey
	subscribers   sync.Map
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
	StreamKey      string `json:"stream_key"`
	DetectedBitrate int   `json:"detected_bitrate"`
}

type StreamStatusResponse struct {
	Status         string `json:"status"`
	BitrateWarning bool   `json:"bitrate_warning"`
}

func (s *StreamService) StartStream(ctx context.Context, streamKey string, detectedBitrate int) (int, error) {
	log.Printf("STREAM_CALLBACK: Received /stream/start callback with stream_key=%s, detected_bitrate=%d", streamKey, detectedBitrate)

	user, err := s.userRepo.GetByStreamKey(ctx, streamKey)
	if err != nil {
		log.Printf("STREAM_CALLBACK_ERROR: Failed to get user by stream_key=%s: %v", streamKey, err)
		return 0, err
	}
	if user == nil {
		log.Printf("STREAM_CALLBACK_ERROR: User not found for stream_key=%s", streamKey)
		return 0, fmt.Errorf("user not found for stream key")
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

	if oldStreamKey, ok := s.userLiveKeys.Load(user.ID); ok {
		s.liveUsers.Delete(oldStreamKey)
		s.userLiveKeys.Delete(user.ID)
		log.Printf("STREAM_CALLBACK_CLEANUP: Removed stale live entry for user %d (old stream key: %s)", user.ID, oldStreamKey)
	}

	s.liveUsers.Store(streamKey, LiveStreamStatus{
		UserID:         user.ID,
		StreamKey:      streamKey,
		BitrateWarning: bitrateWarning,
		StartTime:      time.Now(),
	})
	s.userLiveKeys.Store(user.ID, streamKey)

	s.Notify(user.ID, StreamEvent{Type: "start", Status: "live", BitrateWarning: bitrateWarning})

	log.Printf("STREAM_CALLBACK_USER_LIVE: Marked user %d as live with stream_key=%s, bitrate_warning=%t",
		user.ID, streamKey, bitrateWarning)

	return user.ID, nil
}

func (s *StreamService) StartForwardingWorkers(ctx context.Context, userID int) {
	time.Sleep(500 * time.Millisecond)

	destinations, err := s.userRepo.GetDestinationsWithStreamKey(ctx, userID)
	if err != nil {
		log.Printf("STREAM_CALLBACK_ERROR: Failed to get destinations for user %d: %v", userID, err)
		return
	}

	log.Printf("STREAM_CALLBACK_DESTINATIONS: Found %d destinations with stream keys for user %d", len(destinations), userID)

	activeSessions, err := s.broadcastRepo.GetActiveByUserID(ctx, userID)
	if err != nil {
		log.Printf("STREAM_CALLBACK_WARNING: Failed to get active sessions for cleanup: %v", err)
	} else {
		now := time.Now()
		for _, session := range activeSessions {
			duration := int(now.Sub(session.StartedAt).Minutes())
			if err := s.broadcastRepo.UpdateEndTime(ctx, session.ID, now, duration); err != nil {
				log.Printf("STREAM_CALLBACK_WARNING: Failed to close stale session %d: %v", session.ID, err)
			} else {
				log.Printf("STREAM_CALLBACK_CLEANUP: Closed stale session %d with duration %d minutes", session.ID, duration)
			}
		}
	}

	for _, dest := range destinations {
		_, err := s.broadcastRepo.Create(ctx, userID, dest.ID)
		if err != nil {
			log.Printf("STREAM_CALLBACK_BROADCAST_ERROR: Failed to create broadcast session for user %d, destination %d: %v",
				userID, dest.ID, err)
			continue
		}

		if err := s.workerSvc.StartWorker(ctx, userID, dest.ID); err != nil {
			log.Printf("STREAM_CALLBACK_WORKER_ERROR: Failed to start worker for user %d, destination %d: %v",
				userID, dest.ID, err)
			continue
		}

		log.Printf("STREAM_CALLBACK_WORKER_SUCCESS: Started forwarding worker for user %d, destination %d (%s)",
			userID, dest.ID, dest.Destination.Name)
	}

	log.Printf("STREAM_CALLBACK_COMPLETE: Successfully started %d forwarding workers for user %d",
		len(destinations), userID)
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
	s.userLiveKeys.Delete(user.ID)
	s.Notify(user.ID, StreamEvent{Type: "stop", Status: "offline", BitrateWarning: false})
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
	streamKey, ok := s.userLiveKeys.Load(userID)
	if !ok {
		return &StreamStatusResponse{Status: "offline", BitrateWarning: false}, nil
	}

	ls, ok := s.liveUsers.Load(streamKey)
	if !ok {
		return &StreamStatusResponse{Status: "offline", BitrateWarning: false}, nil
	}

	liveStatus := ls.(LiveStreamStatus)
	return &StreamStatusResponse{
		Status:         "live",
		BitrateWarning: liveStatus.BitrateWarning,
	}, nil
}

func (s *StreamService) IsUserLive(userID int) bool {
	_, ok := s.userLiveKeys.Load(userID)
	return ok
}

func (s *StreamService) GetLiveStatusForUser(userID int) *LiveStreamStatus {
	streamKey, ok := s.userLiveKeys.Load(userID)
	if !ok {
		return nil
	}
	ls, ok := s.liveUsers.Load(streamKey)
	if !ok {
		return nil
	}
	status := ls.(LiveStreamStatus)
	return &status
}

func (s *StreamService) Subscribe(userID int) chan StreamEvent {
	ch := make(chan StreamEvent, 64)
	s.subscribers.Store(userID, ch)
	return ch
}

func (s *StreamService) Unsubscribe(userID int) {
	if ch, ok := s.subscribers.LoadAndDelete(userID); ok {
		close(ch.(chan StreamEvent))
	}
}

func (s *StreamService) Notify(userID int, event StreamEvent) {
	if ch, ok := s.subscribers.Load(userID); ok {
		select {
		case ch.(chan StreamEvent) <- event:
		default:
		}
	}
}
