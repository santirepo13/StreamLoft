package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"streamloft-api/internal/errors"
	"streamloft-api/internal/middleware"
	"streamloft-api/internal/service"
)

type BroadcastHandler struct {
	broadcastService *service.BroadcastService
}

func NewBroadcastHandler(broadcastSvc *service.BroadcastService) *BroadcastHandler {
	return &BroadcastHandler{
		broadcastService: broadcastSvc,
	}
}

func (h *BroadcastHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		errors.RespondWithError(c, errors.Unauthorized("Unauthorized"))
		return
	}

	broadcasts, err := h.broadcastService.GetBroadcasts(c.Request.Context(), userID)
	if err != nil {
		errors.RespondWithError(c, errors.Internal("failed to get broadcasts: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"broadcasts": broadcasts})
}