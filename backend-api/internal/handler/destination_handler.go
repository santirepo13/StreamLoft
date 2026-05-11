package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"streamloft-api/internal/errors"
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
	StreamKey string `json:"stream_key"`
}

func (h *DestinationHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		errors.RespondWithError(c, errors.Unauthorized("Unauthorized"))
		return
	}

	destinations, err := h.destinationService.GetDestinations(c.Request.Context(), userID)
	if err != nil {
		errors.RespondWithError(c, errors.Internal("failed to get destinations: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"destinations": destinations})
}

func (h *DestinationHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		errors.RespondWithError(c, errors.Unauthorized("Unauthorized"))
		return
	}

	destID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		errors.RespondWithError(c, errors.BadRequest("invalid destination id"))
		return
	}

	var req UpdateDestinationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.BindJSONError(c, err)
		return
	}

	_, err = h.destinationService.UpdateStreamKey(c.Request.Context(), userID, destID, req.StreamKey)
	if err != nil {
		errors.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}