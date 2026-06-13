package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"streamloft-api/internal/config"
	"streamloft-api/internal/middleware"
	"streamloft-api/internal/service"
)

type UserHandler struct {
	userService *service.UserService
	cfg         *config.Config
}

func NewUserHandler(userService *service.UserService, cfg *config.Config) *UserHandler {
	return &UserHandler{
		userService: userService,
		cfg:         cfg,
	}
}

func (h *UserHandler) GetUser(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		http.Error(c.Writer, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), userID, h.cfg.RTMPURL)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, user)
}