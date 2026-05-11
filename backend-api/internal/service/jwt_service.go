package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"streamloft-api/internal/interfaces"
)

type Claims struct {
	UserID    int    `json:"user_id"`
	MachineID string `json:"machine_id"`
	jwt.RegisteredClaims
}

type JWTService struct {
	secret []byte
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{secret: []byte(secret)}
}

func (s *JWTService) GenerateToken(userID int, machineID string, expiry time.Time) (string, error) {
	claims := Claims{
		UserID:    userID,
		MachineID: machineID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:   "streamloft-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *JWTService) ValidateToken(tokenString string) (map[string]interface{}, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, ErrTokenExpired
	}

	return map[string]interface{}{
		"user_id":    claims.UserID,
		"machine_id": claims.MachineID,
	}, nil
}

func (s *JWTService) RefreshToken(oldTokenString string, newExpiry time.Time) (string, error) {
	claims, err := s.ValidateToken(oldTokenString)
	if err != nil {
		return "", err
	}

	userID := int(claims["user_id"].(int))
	machineID := claims["machine_id"].(string)

	return s.GenerateToken(userID, machineID, newExpiry)
}

var _ interfaces.TokenService = (*JWTService)(nil)