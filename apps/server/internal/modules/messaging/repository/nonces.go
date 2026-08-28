package messagingrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	messagingsql "github.com/complexus-tech/projects-api/internal/modules/messaging/repository/sqlc"
	"github.com/google/uuid"
)

type NonceInput struct {
	Provider            string
	Purpose             string
	NonceHash           []byte
	WorkspaceID         uuid.UUID
	UserID              *uuid.UUID
	ExternalWorkspaceID string
	ExternalUserID      string
	Payload             json.RawMessage
	ExpiresAt           time.Time
}

type NonceRecord struct {
	ID                  uuid.UUID
	Provider            string
	Purpose             string
	WorkspaceID         uuid.UUID
	UserID              *uuid.UUID
	ExternalWorkspaceID *string
	ExternalUserID      *string
	Payload             json.RawMessage
	ExpiresAt           time.Time
	ConsumedAt          *time.Time
}

type NonceConsumeInput struct {
	Provider    string
	Purpose     string
	NonceHash   []byte
	WorkspaceID *uuid.UUID
	UserID      *uuid.UUID
	Now         time.Time
}

func (repository *Repository) CreateNonce(ctx context.Context, input NonceInput) error {
	if !repository.configured() {
		return errors.New("messaging repository is not configured")
	}
	payload := input.Payload
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}
	if err := repository.queries.CreateNonce(ctx, messagingsql.CreateNonceParams{
		Provider:            input.Provider,
		Purpose:             input.Purpose,
		NonceHash:           input.NonceHash,
		WorkspaceID:         input.WorkspaceID,
		UserID:              input.UserID,
		ExternalWorkspaceID: input.ExternalWorkspaceID,
		ExternalUserID:      input.ExternalUserID,
		Payload:             payload,
		ExpiresAt:           input.ExpiresAt,
	}); err != nil {
		return fmt.Errorf("create messaging nonce: %w", err)
	}
	return nil
}

// ConsumeNonce atomically marks a nonce used and returns its bound identity.
func (repository *Repository) ConsumeNonce(ctx context.Context, input NonceConsumeInput) (NonceRecord, error) {
	if !repository.configured() {
		return NonceRecord{}, errors.New("messaging repository is not configured")
	}
	row, err := repository.queries.ConsumeNonce(ctx, messagingsql.ConsumeNonceParams{
		ConsumedAt:  &input.Now,
		UserID:      input.UserID,
		Provider:    input.Provider,
		Purpose:     input.Purpose,
		NonceHash:   input.NonceHash,
		WorkspaceID: input.WorkspaceID,
	})
	if err != nil {
		return NonceRecord{}, err
	}
	return NonceRecord{
		ID: row.ID, Provider: row.Provider, Purpose: row.Purpose,
		WorkspaceID: row.WorkspaceID, UserID: row.UserID,
		ExternalWorkspaceID: row.ExternalWorkspaceID, ExternalUserID: row.ExternalUserID,
		Payload: json.RawMessage(row.Payload), ExpiresAt: row.ExpiresAt, ConsumedAt: row.ConsumedAt,
	}, nil
}
