package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"streamloft-api/internal/errors"
	"streamloft-api/internal/interfaces"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	UserIDKey           = "user_id"
)

func AuthMiddleware(validator interfaces.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			errors.RespondWithError(c, errors.Unauthorized("Authorization header required"))
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			errors.RespondWithError(c, errors.Unauthorized("Invalid authorization format"))
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, BearerPrefix)

		claims, err := validator.ValidateToken(tokenString)
		if err != nil {
			errors.RespondWithError(c, errors.Unauthorized("Invalid or expired token"))
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(int)
		if !ok {
			errors.RespondWithError(c, errors.Unauthorized("Invalid token claims"))
			c.Abort()
			return
		}

		c.Set(UserIDKey, int(userID))
		c.Next()
	}
}

func GetUserID(c *gin.Context) int {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return 0
	}
	return userID.(int)
}