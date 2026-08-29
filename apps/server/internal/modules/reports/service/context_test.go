package reports

import (
	"context"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

func reportTestContext(actorID uuid.UUID) context.Context {
	return platformauth.SetUserID(context.Background(), actorID)
}
