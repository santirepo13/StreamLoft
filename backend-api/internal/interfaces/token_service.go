package interfaces

import "time"

type TokenService interface {
	GenerateToken(userID int, machineID string, expiry time.Time) (string, error)
	ValidateToken(tokenString string) (map[string]interface{}, error)
	RefreshToken(oldTokenString string, newExpiry time.Time) (string, error)
}