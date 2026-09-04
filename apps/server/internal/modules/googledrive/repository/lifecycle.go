package googledriverepository

import (
	"context"
	"errors"
	"strings"

	"github.com/complexus-tech/projects-api/internal/modules/googledrive/domain"
	"github.com/google/uuid"
)

const (
	providerUserLifecycleLockPrefix    = "google-drive-provider-user:"
	providerSubjectLifecycleLockPrefix = "google-drive-provider-subject:"
)

// WithinProviderUserLifecycle serializes provider-side grant mutations for a
// FortyOne user without opening a database transaction across network I/O.
func (repository *Repository) WithinProviderUserLifecycle(
	ctx context.Context,
	userID uuid.UUID,
	operation func(context.Context) error,
) error {
	if userID == uuid.Nil || operation == nil {
		return domain.ErrInvalidInput
	}
	return repository.withProviderLifecycle(ctx, providerUserLifecycleLockPrefix+userID.String(), operation)
}

// WithinProviderSubjectLifecycle is the global grant fence. Google revokes a
// subject's grant for the whole Cloud project, so this lock cannot be scoped to
// one FortyOne user or workspace.
func (repository *Repository) WithinProviderSubjectLifecycle(
	ctx context.Context,
	googleSubject string,
	operation func(context.Context) error,
) error {
	googleSubject = strings.TrimSpace(googleSubject)
	if googleSubject == "" || operation == nil {
		return domain.ErrInvalidInput
	}
	return repository.withProviderLifecycle(ctx, providerSubjectLifecycleLockPrefix+googleSubject, operation)
}

func (repository *Repository) withProviderLifecycle(
	ctx context.Context,
	lockKey string,
	operation func(context.Context) error,
) error {
	if repository == nil || repository.runProviderLock == nil {
		return errors.New("Google Drive provider lifecycle locks are unavailable")
	}
	return repository.runProviderLock(ctx, lockKey, operation)
}
