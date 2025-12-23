package services

import (
	"context"
	"database/sql"
	"time"

	"go-reasonable-api/app/errors"
	"go-reasonable-api/app/interfaces/repositories"
	"go-reasonable-api/app/interfaces/services"
	"go-reasonable-api/db/sqlcgen"
	"go-reasonable-api/support/config"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultAuthTokenTTL is the default time-to-live for authentication tokens
	// when not explicitly configured
	DefaultAuthTokenTTL = 365 * 24 * time.Hour // 1 year
)

type SessionService struct {
	config        *config.Config
	userRepo      repositories.UserRepository
	authTokenRepo repositories.AuthTokenRepository
}

func NewSessionService(
	cfg *config.Config,
	userRepo repositories.UserRepository,
	authTokenRepo repositories.AuthTokenRepository,
) *SessionService {
	return &SessionService{
		config:        cfg,
		userRepo:      userRepo,
		authTokenRepo: authTokenRepo,
	}
}

func (s *SessionService) Create(ctx context.Context, email, password string) (*sqlcgen.User, string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if eris.Is(err, sql.ErrNoRows) {
			return nil, "", errors.ErrInvalidCredentials
		}
		return nil, "", eris.Wrap(err, "failed to get user by email")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", errors.ErrInvalidCredentials
	}

	// Cancel scheduled deletion if user logs in
	if user.DeletionScheduledAt != nil {
		if err := s.userRepo.CancelDeletion(ctx, user.ID); err != nil {
			return nil, "", eris.Wrap(err, "failed to cancel deletion")
		}
		user.DeletionScheduledAt = nil
	}

	token, err := s.generateToken(ctx, user.ID)
	if err != nil {
		return nil, "", eris.Wrap(err, "failed to generate token")
	}

	return user, token, nil
}

func (s *SessionService) CreateForUser(ctx context.Context, userID uuid.UUID) (string, error) {
	return s.generateToken(ctx, userID)
}

func (s *SessionService) Delete(ctx context.Context, token string) error {
	tokenHash := HashToken(token)
	if err := s.authTokenRepo.RevokeByHash(ctx, tokenHash); err != nil {
		return eris.Wrap(err, "failed to revoke token")
	}
	return nil
}

func (s *SessionService) ValidateToken(ctx context.Context, token string) (*sqlcgen.AuthToken, error) {
	tokenHash := HashToken(token)

	authToken, err := s.authTokenRepo.GetByHash(ctx, tokenHash)
	if err != nil {
		if eris.Is(err, sql.ErrNoRows) {
			return nil, errors.ErrInvalidToken
		}
		return nil, eris.Wrap(err, "failed to get auth token")
	}

	if authToken.RevokedAt != nil {
		return nil, errors.ErrTokenRevoked
	}

	if time.Now().UTC().After(authToken.ExpiresAt) {
		return nil, errors.ErrTokenExpired
	}

	return authToken, nil
}

func (s *SessionService) generateToken(ctx context.Context, userID uuid.UUID) (string, error) {
	token, err := GenerateSecureToken(32)
	if err != nil {
		return "", eris.Wrap(err, "failed to generate secure token")
	}

	tokenHash := HashToken(token)

	now := time.Now().UTC()
	ttl := s.config.Auth.AuthTokenTTL
	if ttl == 0 {
		ttl = DefaultAuthTokenTTL
	}
	expiresAt := now.Add(ttl)

	_, err = s.authTokenRepo.Create(ctx, userID, tokenHash, expiresAt)
	if err != nil {
		return "", eris.Wrap(err, "failed to create auth token")
	}

	return token, nil
}

var _ services.SessionService = (*SessionService)(nil)
