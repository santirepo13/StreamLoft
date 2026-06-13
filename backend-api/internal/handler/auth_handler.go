package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"streamloft-api/internal/config"
	"streamloft-api/internal/errors"
	"streamloft-api/internal/models"
	"streamloft-api/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
	cfg         *config.Config
}

func NewAuthHandler(authService *service.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		cfg:         cfg,
	}
}

type LoginRequest struct {
	NumericID  string `json:"numeric_id" binding:"required"`
	MachineID  string `json:"machine_id" binding:"required"`
}

type LoginResponse struct {
	NumericID    string                `json:"numeric_id"`
	Name         string                `json:"name"`
	RTMPURL      string                `json:"rtmp_url"`
	StreamKey    *string               `json:"stream_key"`
	Bitrate      *int                  `json:"bitrate"`
	AccessToken  string                `json:"access_token"`
	RefreshToken string                `json:"refresh_token"`
	Destinations []DestinationResponse `json:"destinations"`
}

type UserResponse struct {
	ID        int    `json:"id"`
	NumericID string `json:"numeric_id"`
	Name      string `json:"name"`
}

type DestinationResponse struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	RTMPURL      string `json:"rtmp_url"`
	StreamKey    *string `json:"stream_key,omitempty"`
	Configured   bool   `json:"configured"`
	BitLimited   int    `json:"bit_limited"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.BindJSONError(c, err)
		return
	}

	result, err := h.authService.Login(c.Request.Context(), req.NumericID, req.MachineID)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			errors.RespondWithError(c, errors.Unauthorized("Invalid ID"))
			return
		}
		errors.RespondWithError(c, err)
		return
	}

	destinations := make([]DestinationResponse, len(result.Destinations))
	for i, d := range result.Destinations {
		destinations[i] = DestinationResponse{
			ID:         d.ID,
			Name:       d.Destination.Name,
			RTMPURL:    d.Destination.RTMPURL,
			StreamKey:  d.StreamKey,
			Configured: d.StreamKey != nil && *d.StreamKey != "",
			BitLimited: d.BitLimited,
		}
	}

	c.JSON(http.StatusOK, LoginResponse{
		NumericID:    result.User.NumericID,
		Name:         result.User.Name,
		RTMPURL:      h.cfg.RTMPURL,
		StreamKey:    result.User.StreamKey,
		Bitrate:      result.User.Bitrate,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		Destinations: destinations,
	})
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.BindJSONError(c, err)
		return
	}

	accessToken, err := h.authService.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if err == service.ErrInvalidToken {
			errors.RespondWithError(c, errors.Unauthorized("Invalid or expired token"))
			return
		}
		errors.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, RefreshResponse{AccessToken: accessToken})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		errors.RespondWithError(c, errors.Unauthorized("Missing authorization header"))
		return
	}

	accessToken := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		accessToken = authHeader[7:]
	}

	user, err := h.authService.ValidateAccessToken(c.Request.Context(), accessToken)
	if err != nil {
		errors.RespondWithError(c, errors.Unauthorized("Invalid token"))
		return
	}

	machineID := c.GetHeader("X-Machine-ID")
	if machineID == "" {
		errors.RespondWithError(c, errors.BadRequest("Missing machine ID"))
		return
	}

	if err := h.authService.Logout(c.Request.Context(), user.ID, machineID); err != nil {
		errors.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func mapDestinations(destinations []models.UserDestination) []DestinationResponse {
	result := make([]DestinationResponse, len(destinations))
	for i, d := range destinations {
		result[i] = DestinationResponse{
			ID:         d.ID,
			Name:       d.Destination.Name,
			RTMPURL:    d.Destination.RTMPURL,
			Configured: d.StreamKey != nil && *d.StreamKey != "",
			BitLimited: d.BitLimited,
		}
		if d.StreamKey != nil && *d.StreamKey != "" {
			result[i].StreamKey = d.StreamKey
		}
	}
	return result
}