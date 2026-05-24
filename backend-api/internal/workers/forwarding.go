package workers

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

type ForwardingWorker struct {
	process       *exec.Cmd
	isRunning     bool
	startTime     time.Time
	restartCount  int
	lastHealthCheck time.Time
	streamKey        string
	destinationURL   string
	lastError        error
	stderrBuf        *strings.Builder // Captures ffmpeg stderr
}

type ForwardingManager struct {
	workers map[string]*ForwardingWorker
	mu      sync.RWMutex
	srsURL  string
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewForwardingManager(srsURL string) *ForwardingManager {
	ctx, cancel := context.WithCancel(context.Background())
	manager := &ForwardingManager{
		workers: make(map[string]*ForwardingWorker),
		srsURL:  srsURL,
		ctx:     ctx,
		cancel:  cancel,
	}
	
	// Start worker monitoring
	go manager.MonitorWorkers(30 * time.Second) // Check every 30 seconds
	
	return manager
}

func (m *ForwardingManager) workerKey(userID, userDestinationID int) string {
	return fmt.Sprintf("%d-%d", userID, userDestinationID)
}

// handleWorkerCrash manages automatic restart of crashed workers
func (m *ForwardingManager) handleWorkerCrash(key string, userID, userDestinationID int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	worker, exists := m.workers[key]
	if !exists {
		return
	}

	worker.isRunning = false

	if err != nil {
		// Build error details including captured stderr from ffmpeg
		errorFields := log.Error().Int("user_id", userID).Int("destination_id", userDestinationID).Err(err)
		if worker.stderrBuf != nil {
			stderr := worker.stderrBuf.String()
			if stderr != "" {
				errorFields = errorFields.Str("ffmpeg_stderr", stderr)
			}
		}
		errorFields.Msg("forwarding worker crashed")

		// Update worker error state
		worker.lastError = err

		// Check if we should restart this worker
		if worker.restartCount < 3 { // Max 3 restart attempts
			worker.restartCount++
			log.Info().
				Int("user_id", userID).
				Int("destination_id", userDestinationID).
				Int("restart_count", worker.restartCount).
				Msg("attempting to restart crashed worker")

			// Schedule restart after a delay (exponential backoff)
			go m.scheduleWorkerRestart(key, userID, userDestinationID, worker.restartCount)
		} else {
			log.Error().
				Int("user_id", userID).
				Int("destination_id", userDestinationID).
				Int("restart_count", worker.restartCount).
				Msg("max restart attempts reached, not restarting worker")
		}
	} else {
		log.Info().Int("user_id", userID).Int("destination_id", userDestinationID).Msg("forwarding worker stopped gracefully")
	}
}

// scheduleWorkerRestart restarts a worker after a delay
func (m *ForwardingManager) scheduleWorkerRestart(key string, userID, userDestinationID int, attempt int) {
	// Exponential backoff: 1s, 2s, 4s for attempts 1, 2, 3
	delay := time.Duration(attempt) * time.Second
	
	time.Sleep(delay)
	
	log.Info().
		Int("user_id", userID).
		Int("destination_id", userDestinationID).
		Int("attempt", attempt).
		Msg("restarting worker after crash")
	
	// Get the worker to access stored information
	m.mu.Lock()
	worker, exists := m.workers[key]
	if !exists {
		m.mu.Unlock()
		return
	}
	
	// Store the original parameters for restart
	streamKey := worker.streamKey
	destinationURL := worker.destinationURL
	
	m.mu.Unlock()
	
	// Restart the worker
	restartErr := m.StartWorker(context.Background(), userID, userDestinationID, streamKey, destinationURL)
	if restartErr != nil {
		log.Error().
			Int("user_id", userID).
			Int("destination_id", userDestinationID).
			Err(restartErr).
			Msg("failed to restart worker after crash")
	}
}

func (m *ForwardingManager) StartWorker(ctx context.Context, userID int, userDestinationID int, streamKey, destinationRTMPURL string) error {
	key := m.workerKey(userID, userDestinationID)

	m.mu.Lock()
	// Check if a worker entry exists and appears to be running
	if w := m.workers[key]; w != nil && w.isRunning {
		// Double-check if the process is actually still alive (avoid race condition)
		if w.process != nil && w.process.Process != nil {
			if err := w.process.Process.Signal(syscall.Signal(0)); err == nil {
				// Process is alive, worker truly is running
				m.mu.Unlock()
				log.Info().Int("user_id", userID).Int("destination_id", userDestinationID).Msg("worker already running")
				return nil
			}
			// Process dead, clean up stale entry
			delete(m.workers, key)
		} else {
			// No valid process, clean up
			delete(m.workers, key)
		}
	}

	log.Info().
		Int("user_id", userID).
		Int("destination_id", userDestinationID).
		Str("destination", destinationRTMPURL).
		Msg("starting forwarding worker")

	inputURL := fmt.Sprintf("%s/live/%s.flv", m.srsURL, streamKey)

	// Capture stderr to diagnose failures
	stderrBuf := new(strings.Builder)

cmd := exec.Command(
		"ffmpeg",
		"-re",
		"-fflags", "+genpts+igndts+nobuffer",
		"-flags", "low_delay",
		"-probesize", "1M",
		"-analyzeduration", "1M",
		"-i", inputURL,
		"-c:v", "copy",
		"-c:a", "copy",
		"-f", "flv",
		"-flvflags", "no_duration_filesize",
		// Append query params — use & if URL already has a query string, ? otherwise
		destinationRTMPURL+func() string {
			if strings.Contains(destinationRTMPURL, "?") {
				return "&chunk_size=4096&tcp_nodelay=1&rtmp_live=1"
			}
			return "?chunk_size=4096&tcp_nodelay=1&rtmp_live=1"
		}(),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = io.MultiWriter(os.Stderr, stderrBuf)

	if err := cmd.Start(); err != nil {
		m.mu.Unlock()
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	worker := &ForwardingWorker{
		process:        cmd,
		isRunning:      true,
		startTime:      time.Now(),
		restartCount:   0,
		lastHealthCheck: time.Now(),
		streamKey:       streamKey,
		destinationURL:  destinationRTMPURL,
		stderrBuf:       stderrBuf,
	}

	m.workers[key] = worker
	m.mu.Unlock()

	go func() {
		err := cmd.Wait()
		m.handleWorkerCrash(key, userID, userDestinationID, err)
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
		// Try graceful termination first
		if err := worker.process.Process.Signal(syscall.SIGTERM); err != nil {
			log.Warn().Int("user_id", userID).Int("destination_id", userDestinationID).Err(err).Msg("SIGTERM failed, using SIGKILL")
			// If SIGTERM fails, use SIGKILL
			if err := worker.process.Process.Kill(); err != nil {
				log.Error().Int("user_id", userID).Int("destination_id", userDestinationID).Err(err).Msg("failed to kill worker process")
			}
		} else {
			// Wait for graceful termination
			time.Sleep(5 * time.Second)
			// Check if process is still running
			if worker.process.Process.Signal(syscall.Signal(0)) == nil {
				log.Warn().Int("user_id", userID).Int("destination_id", userDestinationID).Msg("process did not terminate gracefully, using SIGKILL")
				worker.process.Process.Kill()
			}
		}
	}

	delete(m.workers, key)
	log.Info().Int("user_id", userID).Int("destination_id", userDestinationID).Msg("forwarding worker stopped")
	return nil
}

func (m *ForwardingManager) StopAllWorkers() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Info().Int("worker_count", len(m.workers)).Msg("stopping all workers")

	for key, worker := range m.workers {
		if worker.process != nil && worker.process.Process != nil {
			worker.process.Process.Signal(syscall.SIGTERM)
		}
		delete(m.workers, key)
	}

	// Give processes time to terminate gracefully
	time.Sleep(5 * time.Second)

	// Check if any processes are still running and force kill them
	m.mu.Lock()
	for key, worker := range m.workers {
		if worker.process != nil && worker.process.Process != nil {
			if err := worker.process.Process.Signal(syscall.Signal(0)); err == nil {
				worker.process.Process.Kill()
				delete(m.workers, key)
			}
		}
	}
	m.mu.Unlock()

	log.Info().Msg("all workers stopped")
	return nil
}

func (m *ForwardingManager) StopAllForUser(ctx context.Context, userID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for key, worker := range m.workers {
		var wUserID, wDestID int
		fmt.Sscanf(key, "%d-%d", &wUserID, &wDestID)
		if wUserID == userID {
			if worker.process != nil && worker.process.Process != nil {
				worker.process.Process.Signal(syscall.SIGTERM)
			}
			delete(m.workers, key)
			log.Info().Int("user_id", userID).Int("destination_id", wDestID).Msg("stopped worker for user")
		}
	}

	return nil
}

func (m *ForwardingManager) IsWorkerRunning(userID, userDestinationID int) bool {
	key := m.workerKey(userID, userDestinationID)
	m.mu.RLock()
	defer m.mu.RUnlock()

	worker, exists := m.workers[key]
	if !exists {
		return false
	}
	
	// Check if process is actually still running
	if worker.process != nil && worker.process.Process != nil {
		if err := worker.process.Process.Signal(syscall.Signal(0)); err != nil {
			// Process is dead, update state
			m.mu.RUnlock()
			m.mu.Lock()
			worker.isRunning = false
			m.mu.Unlock()
			m.mu.RLock()
			// Trigger crash handling
			go m.handleWorkerCrash(key, userID, userDestinationID, fmt.Errorf("process died unexpectedly"))
			return false
		}
	}
	
	return worker.isRunning
}

// MonitorWorkers periodically checks worker health and handles crashes
func (m *ForwardingManager) MonitorWorkers(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		m.checkWorkerHealth()
	}
}

// checkWorkerHealth proactively checks all workers for crashes
func (m *ForwardingManager) checkWorkerHealth() {
	m.mu.RLock()
	
	workersToCheck := make([]string, 0, len(m.workers))
	for key, worker := range m.workers {
		if worker.isRunning && worker.process != nil && worker.process.Process != nil {
			workersToCheck = append(workersToCheck, key)
		}
	}
	
	m.mu.RUnlock()

	// Check each worker
	for _, key := range workersToCheck {
		var userID, userDestinationID int
		fmt.Sscanf(key, "%d-%d", &userID, &userDestinationID)
		
		if !m.IsWorkerRunning(userID, userDestinationID) {
			// Worker is not running, crash handling will be triggered by IsWorkerRunning
			log.Warn().Int("user_id", userID).Int("destination_id", userDestinationID).Msg("worker detected as not running during health check")
		}
	}
	
	// Clean up dead workers
	m.CleanupDeadWorkers()
}

// GetWorkerStats returns comprehensive statistics about all workers
func (m *ForwardingManager) GetWorkerStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	stats := make(map[string]interface{})
	
	totalWorkers := len(m.workers)
	runningWorkers := 0
	crashedWorkers := 0
	totalRestarts := 0
	
	workerDetails := make([]map[string]interface{}, 0)
	
	for key, worker := range m.workers {
		var userID, userDestinationID int
		fmt.Sscanf(key, "%d-%d", &userID, &userDestinationID)
		
	workerInfo := map[string]interface{}{
		"user_id":          userID,
		"destination_id":   userDestinationID,
		"is_running":       worker.isRunning,
		"start_time":       worker.startTime,
		"uptime":           time.Since(worker.startTime),
		"restart_count":    worker.restartCount,
		"last_health_check": worker.lastHealthCheck,
		"stream_key":       worker.streamKey,
		"destination_url":  worker.destinationURL,
		"last_error":       worker.lastError,
	}

		if worker.stderrBuf != nil {
			workerInfo["ffmpeg_stderr"] = worker.stderrBuf.String()
		}
		
		if worker.isRunning {
			runningWorkers++
		} else {
			crashedWorkers++
		}
		
		totalRestarts += worker.restartCount
		workerDetails = append(workerDetails, workerInfo)
	}
	
	stats["total_workers"] = totalWorkers
	stats["running_workers"] = runningWorkers
	stats["crashed_workers"] = crashedWorkers
	stats["total_restarts"] = totalRestarts
	stats["worker_details"] = workerDetails
	stats["last_health_check"] = time.Now()
	
	return stats
}

// GetWorkerStatus returns status for a specific worker
func (m *ForwardingManager) GetWorkerStatus(userID, userDestinationID int) (map[string]interface{}, bool) {
	key := m.workerKey(userID, userDestinationID)
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	worker, exists := m.workers[key]
	if !exists {
		return nil, false
	}
	
	status := map[string]interface{}{
		"user_id":          userID,
		"destination_id":   userDestinationID,
		"is_running":       worker.isRunning,
		"start_time":       worker.startTime,
		"uptime":           time.Since(worker.startTime),
		"restart_count":    worker.restartCount,
		"last_health_check": worker.lastHealthCheck,
		"stream_key":       worker.streamKey,
		"destination_url":  worker.destinationURL,
		"last_error":       worker.lastError,
	}

	if worker.stderrBuf != nil {
		status["ffmpeg_stderr"] = worker.stderrBuf.String()
	}
	
	return status, true
}

// UpdateWorkerHealth updates the health check time and error state for a worker
func (m *ForwardingManager) UpdateWorkerHealth(key string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if worker, exists := m.workers[key]; exists {
		worker.lastHealthCheck = time.Now()
		worker.lastError = err
	}
}

// CleanupDeadWorkers removes workers that are no longer running and have no active processes
func (m *ForwardingManager) CleanupDeadWorkers() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	keysToDelete := make([]string, 0)
	
	for key, worker := range m.workers {
		if !worker.isRunning && (worker.process == nil || worker.process.Process == nil || worker.process.Process.Signal(syscall.Signal(0)) != nil) {
			keysToDelete = append(keysToDelete, key)
		}
	}
	
	for _, key := range keysToDelete {
		delete(m.workers, key)
		log.Info().Str("worker_key", key).Msg("cleaned up dead worker")
	}
}