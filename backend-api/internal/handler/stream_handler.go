package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"streamloft-api/internal/config"
	"streamloft-api/internal/errors"
	"streamloft-api/internal/middleware"
	"streamloft-api/internal/service"
)

type StreamHandler struct {
	streamService *service.StreamService
	userService   *service.UserService
	workerService *service.WorkerService
	cfg           *config.Config
}

func NewStreamHandler(streamSvc *service.StreamService, userSvc *service.UserService, workerSvc *service.WorkerService, cfg *config.Config) *StreamHandler {
	return &StreamHandler{
		streamService: streamSvc,
		userService:   userSvc,
		workerService: workerSvc,
		cfg:           cfg,
	}
}

type StreamStartRequest struct {
	StreamKey       string `json:"stream_key"`
	Stream           string `json:"stream"`
	DetectedBitrate int    `json:"detected_bitrate"`
}

type UpdateBitrateRequest struct {
	Bitrate int `json:"bitrate" binding:"required"`
}

func (h *StreamHandler) Start(c *gin.Context) {
	log.Printf("STREAM_HANDLER_START: Received POST /stream/start request")
	
	// Try to get stream key from query params first (SRS callback format: ?stream=KEY)
	streamKey := c.Query("stream")
	detectedBitrate := 0
	
	// Also check JSON body for backward compatibility or other callers
	var req StreamStartRequest
	if err := c.ShouldBindJSON(&req); err == nil {
		if streamKey == "" && req.StreamKey != "" {
			streamKey = req.StreamKey
		}
		if streamKey == "" && req.Stream != "" {
			streamKey = req.Stream
		}
		if req.DetectedBitrate > 0 {
			detectedBitrate = req.DetectedBitrate
		}
	}

	log.Printf("STREAM_HANDLER_INFO: Processing stream_key=%s, detected_bitrate=%d", streamKey, detectedBitrate)

	if streamKey == "" {
		log.Printf("STREAM_HANDLER_ERROR: stream_key is required")
		errors.RespondWithError(c, errors.BadRequest("stream_key is required"))
		return
	}

	if err := h.streamService.StartStream(c.Request.Context(), streamKey, detectedBitrate); err != nil {
		log.Printf("STREAM_HANDLER_ERROR: Failed to start stream: %v", err)
		errors.RespondWithError(c, errors.Internal("failed to start stream: "+err.Error()))
		return
	}

	log.Printf("STREAM_HANDLER_SUCCESS: Stream processing completed successfully")
	c.JSON(http.StatusOK, gin.H{"code": 0})
}

func (h *StreamHandler) Stop(c *gin.Context) {
	log.Printf("STREAM_HANDLER_STOP: Received POST /stream/stop request")
	
	// Try to get stream key from query params first (SRS callback format: ?stream=KEY)
	streamKey := c.Query("stream")
	
	// Also check JSON body for backward compatibility or other callers
	var req StreamStartRequest
	if err := c.ShouldBindJSON(&req); err == nil {
		if streamKey == "" && req.StreamKey != "" {
			streamKey = req.StreamKey
		}
		if streamKey == "" && req.Stream != "" {
			streamKey = req.Stream
		}
	}

	log.Printf("STREAM_HANDLER_INFO: Processing stream_key=%s for stop", streamKey)

	if streamKey == "" {
		log.Printf("STREAM_HANDLER_ERROR: stream_key is required")
		errors.RespondWithError(c, errors.BadRequest("stream_key is required"))
		return
	}

	if err := h.streamService.StopStream(c.Request.Context(), streamKey); err != nil {
		log.Printf("STREAM_HANDLER_ERROR: Failed to stop stream: %v", err)
		errors.RespondWithError(c, errors.Internal("failed to stop stream: "+err.Error()))
		return
	}

	log.Printf("STREAM_HANDLER_SUCCESS: Stream stop processing completed successfully")
	c.JSON(http.StatusOK, gin.H{"code": 0})
}

func (h *StreamHandler) Status(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		errors.RespondWithError(c, errors.Unauthorized("Unauthorized"))
		return
	}

	status, err := h.streamService.GetStatus(c.Request.Context(), userID)
	if err != nil {
		errors.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, status)
}

func (h *StreamHandler) UpdateBitrate(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		errors.RespondWithError(c, errors.Unauthorized("Unauthorized"))
		return
	}

	var req UpdateBitrateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.BindJSONError(c, err)
		return
	}

	if req.Bitrate <= 0 {
		errors.RespondWithError(c, errors.BadRequest("Bitrate must be positive"))
		return
	}

	if err := h.userService.UpdateBitrate(c.Request.Context(), userID, req.Bitrate); err != nil {
		errors.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"bitrate": req.Bitrate})
}

func (h *StreamHandler) Events(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	ch := h.streamService.Subscribe(userID)
	defer h.streamService.Unsubscribe(userID)

	c.Stream(func(w io.Writer) bool {
		select {
		case event, ok := <-ch:
			if !ok {
				return false
			}
			data, _ := json.Marshal(event)
			fmt.Fprintf(w, "data: %s\n\n", data)
			return true
		case <-c.Request.Context().Done():
			return false
		}
	})
}