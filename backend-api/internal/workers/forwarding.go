package workers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type ForwardingWorker struct {
	process   *exec.Cmd
	isRunning bool
	startTime time.Time
}

type ForwardingManager struct {
	workers map[string]*ForwardingWorker
	mu      sync.RWMutex
	srsURL  string
}

func NewForwardingManager(srsURL string) *ForwardingManager {
	return &ForwardingManager{
		workers: make(map[string]*ForwardingWorker),
		srsURL:  srsURL,
	}
}

func (m *ForwardingManager) workerKey(userID, userDestinationID int) string {
	return fmt.Sprintf("%d-%d", userID, userDestinationID)
}

func (m *ForwardingManager) StartWorker(ctx context.Context, userID int, userDestinationID int, streamKey, destinationRTMPURL string) error {
	key := m.workerKey(userID, userDestinationID)

	m.mu.Lock()
	if m.workers[key] != nil && m.workers[key].isRunning {
		m.mu.Unlock()
		log.Info().Int("user_id", userID).Int("destination_id", userDestinationID).Msg("worker already running")
		return nil
	}

	log.Info().
		Int("user_id", userID).
		Int("destination_id", userDestinationID).
		Str("destination", destinationRTMPURL).
		Msg("starting forwarding worker")

	inputURL := fmt.Sprintf("%s/%s", m.srsURL, streamKey)

	cmd := exec.CommandContext(ctx,
		"ffmpeg",
		"-i", inputURL,
		"-c:v", "copy",
		"-c:a", "copy",
		"-f", "flv",
		destinationRTMPURL,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		m.mu.Unlock()
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	m.workers[key] = &ForwardingWorker{
		process:   cmd,
		isRunning: true,
		startTime: time.Now(),
	}
	m.mu.Unlock()

	go func() {
		err := cmd.Wait()
		m.mu.Lock()
		if m.workers[key] != nil {
			m.workers[key].isRunning = false
			if err != nil {
				log.Error().Int("user_id", userID).Int("destination_id", userDestinationID).Err(err).Msg("forwarding worker stopped with error")
			} else {
				log.Info().Int("user_id", userID).Int("destination_id", userDestinationID).Msg("forwarding worker stopped")
			}
		}
		m.mu.Unlock()
	}()

	log.Info().Int("user_id", userID).Int("destination_id", userDestinationID).Msg("forwarding worker started successfully")
	return nil
}

func (m *ForwardingManager) StopWorker(ctx context.Context, userID int, userDestinationID int) error {
	key := m.workerKey(userID, userDestinationID)

	m.mu.Lock()
	defer m.mu.Unlock()

	worker, exists := m.workers[key]
	if !exists || !worker.isRunning {
		log.Info().Int("user_id", userID).Int("destination_id", userDestinationID).Msg("worker not running")
		return nil
	}

	if worker.process != nil && worker.process.Process != nil {
		if err := worker.process.Process.Kill(); err != nil {
			log.Error().Int("user_id", userID).Int("destination_id", userDestinationID).Err(err).Msg("failed to kill worker process")
		}
	}

	delete(m.workers, key)
	log.Info().Int("user_id", userID).Int("destination_id", userDestinationID).Msg("forwarding worker stopped")
	return nil
}

func (m *ForwardingManager) StopAllForUser(ctx context.Context, userID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for key, worker := range m.workers {
		if worker.isRunning {
			var wUserID, wDestID int
			fmt.Sscanf(key, "%d-%d", &wUserID, &wDestID)
			if wUserID == userID {
				if worker.process != nil && worker.process.Process != nil {
					worker.process.Process.Kill()
				}
				delete(m.workers, key)
				log.Info().Int("user_id", userID).Int("destination_id", wDestID).Msg("stopped worker for user")
			}
		}
	}

	return nil
}

func (m *ForwardingManager) IsWorkerRunning(userID, userDestinationID int) bool {
	key := m.workerKey(userID, userDestinationID)
	m.mu.RLock()
	defer m.mu.RUnlock()

	worker, exists := m.workers[key]
	return exists && worker.isRunning
}