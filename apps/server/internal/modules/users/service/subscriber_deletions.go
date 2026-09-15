package users

import (
	"context"
	"errors"
	"time"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	"github.com/complexus-tech/projects-api/pkg/brevo"
	"github.com/google/uuid"
)

type SubscriberCleanupRepository interface {
	Claim(context.Context, time.Time, time.Time) (usersdomain.SubscriberDeletion, bool, error)
	Renew(context.Context, uuid.UUID, uuid.UUID, time.Time, time.Time) (bool, error)
	Complete(context.Context, uuid.UUID, uuid.UUID) error
	Retry(context.Context, uuid.UUID, uuid.UUID, time.Time, time.Time) error
	WithinSubscriberLifecycle(context.Context, string, func(context.Context) error) error
	ActiveAccount(context.Context, string) (usersdomain.AccountSubscriber, bool, error)
}

type SubscriberContactProvider interface {
	DeleteAccountContact(context.Context, string) error
	CreateOrUpdateContact(context.Context, brevo.CreateOrUpdateContactRequest) (*brevo.CreateOrUpdateContactResponse, error)
}

type SubscriberCleanup struct {
	repository SubscriberCleanupRepository
	provider   SubscriberContactProvider
	now        func() time.Time
}

func NewSubscriberCleanup(repository SubscriberCleanupRepository, provider SubscriberContactProvider) *SubscriberCleanup {
	return &SubscriberCleanup{repository: repository, provider: provider, now: time.Now}
}

func (cleanup *SubscriberCleanup) Dispatch(ctx context.Context) (int, error) {
	completed := 0
	var failures error
	for range 25 {
		if err := ctx.Err(); err != nil {
			return completed, errors.Join(failures, err)
		}
		now := cleanup.now().UTC()
		job, claimed, err := cleanup.repository.Claim(ctx, now, now.Add(2*time.Minute))
		if err != nil || !claimed {
			return completed, errors.Join(failures, err)
		}
		err = cleanup.repository.WithinSubscriberLifecycle(ctx, job.Email, func(lockedCtx context.Context) error {
			now := cleanup.now().UTC()
			current, err := cleanup.repository.Renew(lockedCtx, job.ID, job.LeaseToken, now, now.Add(2*time.Minute))
			if err != nil {
				return err
			}
			if !current {
				return errors.New("subscriber cleanup claim expired")
			}
			providerCtx, cancel := context.WithTimeout(lockedCtx, 20*time.Second)
			defer cancel()
			return cleanup.provider.DeleteAccountContact(providerCtx, job.Email)
		})
		finishCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		if err == nil {
			err = cleanup.repository.Complete(finishCtx, job.ID, job.LeaseToken)
			if err == nil {
				completed++
			}
		} else {
			now = cleanup.now().UTC()
			retryErr := cleanup.repository.Retry(finishCtx, job.ID, job.LeaseToken, now.Add(subscriberRetryDelay(job.AttemptCount)), now)
			// Provider errors may contain email addresses or response bodies.
			err = errors.Join(errors.New("subscriber cleanup remains pending"), retryErr)
		}
		cancel()
		failures = errors.Join(failures, err)
	}
	return completed, failures
}

// Queued profile data is not authoritative: deletion may have committed since
// enqueue. The shared email lock also fences an update already in flight when
// the deletion transaction starts. New accounts wait for older provider cleanup.
func (cleanup *SubscriberCleanup) UpdateActiveAccount(ctx context.Context, update usersdomain.SubscriberUpdate) error {
	return cleanup.repository.WithinSubscriberLifecycle(ctx, update.Email, func(lockedCtx context.Context) error {
		account, found, err := cleanup.repository.ActiveAccount(lockedCtx, update.Email)
		if err != nil || !found {
			return err
		}
		if update.UserID != uuid.Nil && update.UserID != account.UserID {
			return nil
		}
		if account.CleanupPending {
			return errors.New("subscriber update is waiting for account cleanup")
		}
		providerCtx, cancel := context.WithTimeout(lockedCtx, 20*time.Second)
		defer cancel()
		attributes := brevo.ContactAttributes{}
		for key, value := range update.Attributes {
			attributes[key] = value
		}
		attributes["NAME"] = account.FullName
		_, err = cleanup.provider.CreateOrUpdateContact(providerCtx, brevo.CreateOrUpdateContactRequest{
			Email:      account.Email,
			Attributes: attributes,
			ListIDs:    update.ListIDs,
		})
		if err != nil {
			return errors.New("subscriber update failed")
		}
		return nil
	})
}

func subscriberRetryDelay(attempt int) time.Duration {
	delay := time.Minute
	for current := 1; current < attempt && delay < time.Hour; current++ {
		delay *= 2
	}
	return min(delay, time.Hour)
}
