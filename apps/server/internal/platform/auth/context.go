package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type contextKey string

const userIDKey contextKey = "userID"

func SetUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func GetUserID(ctx context.Context) (uuid.UUID, error) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok {
		return uuid.UUID{}, errors.New("userID not found in context")
	}
	return id, nil
}
