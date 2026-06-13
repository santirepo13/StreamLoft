package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apierrors "streamloft-api/internal/errors"
	"streamloft-api/internal/middleware"
	"streamloft-api/internal/service"
)

type DestinationHandler struct {
	destinationService *service.DestinationService
}

func NewDestinationHandler(destSvc *service.DestinationService) *DestinationHandler {
	return &DestinationHandler{
		destinationService: destSvc,
	}
}

type UpdateDestinationRequest struct {
	StreamKey  string `json:"stream_key"`
	BitLimited *int   `json:"bit_limited,omitempty"`
}

type ToggleDestinationRequest struct {
	Enabled int `json:"enabled"` // 1=true, 0=false
}

func (h *DestinationHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		apierrors.RespondWithError(c, apierrors.Unauthorized("Unauthorized"))
		return
	}

	destinations, err := h.destinationService.GetDestinations(c.Request.Context(), userID)
	if err != nil {
		apierrors.RespondWithError(c, apierrors.Internal("failed to get destinations: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"destinations": destinations})
}

func (h *DestinationHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		apierrors.RespondWithError(c, apierrors.Unauthorized("Unauthorized"))
		return
	}

	destID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierrors.RespondWithError(c, apierrors.BadRequest("invalid destination id"))
		return
	}

	var req UpdateDestinationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.BindJSONError(c, err)
		return
	}

	_, err = h.destinationService.UpdateStreamKey(c.Request.Context(), userID, destID, req.StreamKey)
	if err != nil {
		apierrors.RespondWithError(c, err)
		return
	}

	if req.BitLimited != nil {
		if err := h.destinationService.UpdateBitLimited(c.Request.Context(), userID, destID, *req.BitLimited); err != nil {
			if errors.Is(err, service.ErrInvalidInput) {
				apierrors.RespondWithError(c, apierrors.BadRequest("bit_limited must be 0 or 1"))
				return
			}
			apierrors.RespondWithError(c, err)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (h *DestinationHandler) Toggle(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		apierrors.RespondWithError(c, apierrors.Unauthorized("Unauthorized"))
		return
	}

	destID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierrors.RespondWithError(c, apierrors.BadRequest("invalid destination id"))
		return
	}

	var req ToggleDestinationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.BindJSONError(c, err)
		return
	}

	if err := h.destinationService.ToggleDestination(c.Request.Context(), userID, destID, req.Enabled); err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			apierrors.RespondWithError(c, apierrors.BadRequest("enabled must be 0 or 1"))
			return
		}
		apierrors.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "toggled"})
}