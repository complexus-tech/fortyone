//go:build integration

package chatrepository

import (
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestEventLeasesCannotBeRenewedOrCompletedAfterExpiryOrReplacement(t *testing.T) {
	f := newFixture(t)
	f.create(t, "Lease test", nil)
	claimed, err := f.repo.ClaimEvents(f.ctx, 1)
	require.NoError(t, err)
	require.Len(t, claimed, 1)
	first := claimed[0]
	require.NoError(t, f.repo.RenewEventLease(f.ctx, first.ID, first.LeaseToken))
	require.ErrorIs(t, f.repo.RenewEventLease(f.ctx, first.ID, uuid.New()), domain.ErrConflict)
	_, err = f.pool.Exec(f.ctx, `UPDATE chat_outbox SET leased_until=now()-interval '1 second' WHERE id=$1`, first.ID)
	require.NoError(t, err)
	require.ErrorIs(t, f.repo.RenewEventLease(f.ctx, first.ID, first.LeaseToken), domain.ErrConflict)
	require.ErrorIs(t, f.repo.CompleteEvent(f.ctx, first.ID, first.LeaseToken), domain.ErrConflict)
	claimed, err = f.repo.ClaimEvents(f.ctx, 1)
	require.NoError(t, err)
	require.Len(t, claimed, 1)
	replacement := claimed[0]
	require.Equal(t, first.ID, replacement.ID)
	require.NotEqual(t, first.LeaseToken, replacement.LeaseToken)
	require.ErrorIs(t, f.repo.RenewEventLease(f.ctx, first.ID, first.LeaseToken), domain.ErrConflict)
	require.ErrorIs(t, f.repo.ReleaseEvent(f.ctx, first.ID, first.LeaseToken), domain.ErrConflict)
	require.NoError(t, f.repo.RenewEventLease(f.ctx, replacement.ID, replacement.LeaseToken))
	require.NoError(t, f.repo.CompleteEvent(f.ctx, replacement.ID, replacement.LeaseToken))
	require.ErrorIs(t, f.repo.RenewEventLease(f.ctx, replacement.ID, replacement.LeaseToken), domain.ErrConflict)
}

func TestReadEventRecipientIsOnlyCurrentAuthorizedReader(t *testing.T) {
	f := newFixture(t)
	event := domain.Event{Type: "read.updated", WorkspaceID: f.workspace, ConversationID: f.room, ActorID: f.author}
	recipients, err := f.repo.ReadRecipientIDs(f.ctx, event)
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{f.author}, recipients)
	event.ActorID = f.outsider
	recipients, err = f.repo.ReadRecipientIDs(f.ctx, event)
	require.NoError(t, err)
	require.Empty(t, recipients)
	event.ActorID, event.WorkspaceID = f.author, uuid.New()
	recipients, err = f.repo.ReadRecipientIDs(f.ctx, event)
	require.NoError(t, err)
	require.Empty(t, recipients)
	event.WorkspaceID = f.workspace
	_, err = f.pool.Exec(f.ctx, `DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, f.team, f.author)
	require.NoError(t, err)
	recipients, err = f.repo.ReadRecipientIDs(f.ctx, event)
	require.NoError(t, err)
	require.Empty(t, recipients)
}
