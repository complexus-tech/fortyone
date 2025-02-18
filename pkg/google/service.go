package google

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/api/idtoken"
)

var (
	ErrInvalidToken  = errors.New("invalid google token")
	ErrTokenExpired  = errors.New("token expired")
	ErrEmailMismatch = errors.New("email mismatch")
)

// Service handles Google OAuth operations
type Service struct {
	clientID  string
	validator *idtoken.Validator
}

// NewService creates a new Google service instance
func NewService(clientID string) (*Service, error) {
	validator, err := idtoken.NewValidator(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create validator: %w", err)
	}

	return &Service{
		clientID:  clientID,
		validator: validator,
	}, nil
}

// VerifyToken validates a Google ID token and ensures it matches the provided email
func (s *Service) VerifyToken(ctx context.Context, token string, email string) (*idtoken.Payload, error) {
	payload, err := s.validator.Validate(ctx, token, s.clientID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	// Verify email matches
	if tokenEmail, ok := payload.Claims["email"].(string); !ok || tokenEmail != email {
		return nil, ErrEmailMismatch
	}

	return payload, nil
}
