package handler

import (
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
	StreamKey     string `json:"stream_key"`
	DetectedBitrate int  `json:"detected_bitrate"`
}

type UpdateBitrateRequest struct {
	Bitrate int `json:"bitrate" binding:"required"`
}

func (h *StreamHandler) Start(c *gin.Context) {
	log.Printf("STREAM_HANDLER_START: Received POST /stream/start request")
	
	var req StreamStartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("STREAM_HANDLER_ERROR: Failed to parse JSON: %v", err)
		errors.BindJSONError(c, err)
		return
	}

	log.Printf("STREAM_HANDLER_INFO: Processing stream_key=%s, detected_bitrate=%d", req.StreamKey, req.DetectedBitrate)

	if req.StreamKey == "" {
		log.Printf("STREAM_HANDLER_ERROR: stream_key is required")
		errors.RespondWithError(c, errors.BadRequest("stream_key is required"))
		return
	}

	if err := h.streamService.StartStream(c.Request.Context(), req.StreamKey, req.DetectedBitrate); err != nil {
		log.Printf("STREAM_HANDLER_ERROR: Failed to start stream: %v", err)
		errors.RespondWithError(c, errors.Internal("failed to start stream: "+err.Error()))
		return
	}

	log.Printf("STREAM_HANDLER_SUCCESS: Stream processing completed successfully")
	c.JSON(http.StatusOK, gin.H{"status": "started"})
}

func (h *StreamHandler) Stop(c *gin.Context) {
	log.Printf("STREAM_HANDLER_STOP: Received POST /stream/stop request")
	
	var req StreamStartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("STREAM_HANDLER_ERROR: Failed to parse JSON: %v", err)
		errors.BindJSONError(c, err)
		return
	}

	log.Printf("STREAM_HANDLER_INFO: Processing stream_key=%s for stop", req.StreamKey)

	if req.StreamKey == "" {
		log.Printf("STREAM_HANDLER_ERROR: stream_key is required")
		errors.RespondWithError(c, errors.BadRequest("stream_key is required"))
		return
	}

	if err := h.streamService.StopStream(c.Request.Context(), req.StreamKey); err != nil {
		log.Printf("STREAM_HANDLER_ERROR: Failed to stop stream: %v", err)
		errors.RespondWithError(c, errors.Internal("failed to stop stream: "+err.Error()))
		return
	}

	log.Printf("STREAM_HANDLER_SUCCESS: Stream stop processing completed successfully")
	c.JSON(http.StatusOK, gin.H{"status": "stopped"})
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