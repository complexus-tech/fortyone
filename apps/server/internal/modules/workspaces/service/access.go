package workspaces

import (
	"context"

	"github.com/google/uuid"
)

// ResolveCurrentMembership reads live membership and role state. Callers must
// not cache the result for authorization decisions.
func (s *Service) ResolveCurrentMembership(
	ctx context.Context,
	slug string,
	userID uuid.UUID,
) (CurrentMembership, error) {
	return s.repo.ResolveCurrentMembership(ctx, slug, userID)
}

// RecordAccess records non-critical activity only after membership has been
// resolved successfully. Authorization does not depend on this write.
func (s *Service) RecordAccess(ctx context.Context, workspaceID, userID uuid.UUID) error {
	return s.repo.RecordAccess(ctx, workspaceID, userID)
}
