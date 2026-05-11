package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"streamloft-api/internal/config"
	"streamloft-api/internal/crypto"
	"streamloft-api/internal/interfaces"
	"streamloft-api/internal/models"
)

type AuthService struct {
	userRepo     interfaces.UserRepository
	sessionRepo  interfaces.SessionRepository
	machineRepo  interfaces.MachineRepository
	encryptor    *crypto.Encryptor
	tokenService interfaces.TokenService
	tokenExpiry  time.Time
}

func NewAuthService(
	userRepo interfaces.UserRepository,
	sessionRepo interfaces.SessionRepository,
	machineRepo interfaces.MachineRepository,
	enc *crypto.Encryptor,
	tokenSvc interfaces.TokenService,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		sessionRepo:  sessionRepo,
		machineRepo:  machineRepo,
		encryptor:    enc,
		tokenService: tokenSvc,
		tokenExpiry:  cfg.TokenExpiry(),
	}
}

type LoginResult struct {
	User         *models.User
	AccessToken  string
	RefreshToken string
	Destinations []models.UserDestination
}

func (s *AuthService) Login(ctx context.Context, numericID, machineID string) (*LoginResult, error) {
	user, err := s.userRepo.GetByNumericID(ctx, numericID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if user.StreamKey == "" {
		streamKey := generateStreamKey()
		if err := s.userRepo.UpdateStreamKey(ctx, user.ID, streamKey); err != nil {
			return nil, err
		}
		user.StreamKey = streamKey
	}

	_, err = s.machineRepo.Upsert(ctx, machineID, user.ID)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.tokenService.GenerateToken(user.ID, machineID, s.tokenExpiry)
	if err != nil {
		return nil, err
	}

	refreshToken := generateTokenString()

	encAccessToken, err := s.encryptor.Encrypt(accessToken)
	if err != nil {
		return nil, err
	}
	encRefreshToken, err := s.encryptor.Encrypt(refreshToken)
	if err != nil {
		return nil, err
	}

	_, err = s.sessionRepo.Create(ctx, user.ID, machineID, encAccessToken, encRefreshToken, s.tokenExpiry)
	if err != nil {
		return nil, err
	}

	destinations, err := s.userRepo.GetDestinationsByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Destinations: destinations,
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (string, error) {
	session, err := s.sessionRepo.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", err
	}
	if session == nil {
		return "", ErrInvalidToken
	}

	newAccessToken, err := s.tokenService.RefreshToken(refreshToken, s.tokenExpiry)
	if err != nil {
		return "", err
	}

	encAccessToken, err := s.encryptor.Encrypt(newAccessToken)
	if err != nil {
		return "", err
	}

	if err := s.sessionRepo.UpdateAccessToken(ctx, session.ID, encAccessToken, s.tokenExpiry); err != nil {
		return "", err
	}

	return newAccessToken, nil
}

func (s *AuthService) Logout(ctx context.Context, userID int, machineID string) error {
	return s.sessionRepo.DeleteByUserAndMachine(ctx, userID, machineID)
}

func (s *AuthService) ValidateAccessToken(ctx context.Context, accessToken string) (*models.User, error) {
	claims, err := s.tokenService.ValidateToken(accessToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	userID := int(claims["user_id"].(float64))
	machineID := claims["machine_id"].(string)

	session, err := s.sessionRepo.GetByAccessToken(ctx, accessToken)
	if err != nil || session == nil {
		return nil, ErrInvalidToken
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	_ = machineID

	return user, nil
}

func generateStreamKey() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func generateTokenString() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}