//go:build integration

package webhooksrepository

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repositoryFixture struct {
	repository  *Repository
	pool        *pgxpool.Pool
	workspaceID uuid.UUID
	now         time.Time
}

func newRepositoryFixture(t *testing.T) repositoryFixture {
	t.Helper()
	postgres := testkit.NewPostgres(t)
	fixture := repositoryFixture{
		repository:  New(postgres.Pool),
		pool:        postgres.Pool,
		workspaceID: uuid.New(),
		now:         time.Now().UTC().Truncate(time.Microsecond),
	}
	userID := uuid.New()
	suffix := uuid.NewString()
	if _, err := fixture.pool.Exec(t.Context(), `
		INSERT INTO users (user_id, username, email, full_name)
		VALUES ($1, $2, $3, 'Webhook Owner')
	`, userID, "webhook-"+suffix, "webhook-"+suffix+"@example.com"); err != nil {
		t.Fatalf("insert webhook owner: %v", err)
	}
	if _, err := fixture.pool.Exec(t.Context(), `
		INSERT INTO workspaces (workspace_id, name, slug, created_by)
		VALUES ($1, 'Webhook Workspace', $2, $3)
	`, fixture.workspaceID, "webhook-"+suffix, userID); err != nil {
		t.Fatalf("insert webhook workspace: %v", err)
	}
	return fixture
}

func (fixture repositoryFixture) envelope(deliveryID string) webhooks.Envelope {
	return webhooks.Envelope{
		Version:                webhooks.CurrentEnvelopeVersion,
		Provider:               integrations.ProviderKey("slack"),
		DeliveryID:             deliveryID,
		EventType:              "message.im",
		ExternalAccountID:      "T-" + fixture.workspaceID.String(),
		WorkspaceID:            fixture.workspaceID,
		InstallationID:         uuid.New(),
		InstallationGeneration: uuid.New(),
		TraceID:                "4bf92f3577b34da6a3ce929d0e0e4736",
		ReceivedAt:             fixture.now,
	}
}

func TestRepositoryDeliveryLifecycleAndDeduplication(t *testing.T) {
	t.Parallel()
	fixture := newRepositoryFixture(t)
	envelope := fixture.envelope("Ev-lifecycle")
	expiresAt := fixture.now.Add(time.Hour)

	created, inserted, err := fixture.repository.Register(t.Context(), envelope, "ciphertext-v1", expiresAt)
	if err != nil || !inserted {
		t.Fatalf("register webhook delivery = (%#v, %v, %v)", created, inserted, err)
	}
	if created.Version != envelope.Version ||
		created.Provider != envelope.Provider ||
		created.DeliveryID != envelope.DeliveryID ||
		created.EventType != envelope.EventType ||
		created.ExternalAccountID != envelope.ExternalAccountID ||
		created.WorkspaceID != envelope.WorkspaceID ||
		created.InstallationID != envelope.InstallationID ||
		created.InstallationGeneration != envelope.InstallationGeneration ||
		created.TraceID != envelope.TraceID ||
		!created.ReceivedAt.Equal(envelope.ReceivedAt) ||
		created.EncryptedPayload == nil || *created.EncryptedPayload != "ciphertext-v1" {
		t.Fatalf("created webhook delivery = %#v", created)
	}

	conflicting := envelope
	conflicting.EventType = "app_mention"
	duplicate, inserted, err := fixture.repository.Register(t.Context(), conflicting, "ciphertext-v2", expiresAt.Add(time.Hour))
	if err != nil || inserted {
		t.Fatalf("register duplicate delivery = (%#v, %v, %v)", duplicate, inserted, err)
	}
	if duplicate.ID != created.ID || duplicate.EventType != envelope.EventType {
		t.Fatalf("duplicate rewrote immutable identity: %#v", duplicate)
	}

	started, claimed, err := fixture.repository.Start(t.Context(), created.ID, fixture.now.Add(time.Second), time.Minute)
	if err != nil || !claimed || started.Status != webhooks.StatusProcessing || started.AttemptCount != 1 {
		t.Fatalf("start webhook delivery = (%#v, %v, %v)", started, claimed, err)
	}
	if _, claimed, err := fixture.repository.Start(t.Context(), created.ID, fixture.now.Add(2*time.Second), time.Minute); claimed || !errors.Is(err, webhooks.ErrLeaseBusy) {
		t.Fatalf("concurrent start = (%v, %v), want lease busy", claimed, err)
	}
	if err := fixture.repository.Complete(t.Context(), created.ID, webhooks.StatusFailed, "provider.rate_limited", fixture.now.Add(3*time.Second)); err != nil {
		t.Fatalf("fail webhook delivery: %v", err)
	}
	started, claimed, err = fixture.repository.Start(t.Context(), created.ID, fixture.now.Add(4*time.Second), time.Minute)
	if err != nil || !claimed || started.AttemptCount != 2 {
		t.Fatalf("restart failed webhook delivery = (%#v, %v, %v)", started, claimed, err)
	}
	if err := fixture.repository.Complete(t.Context(), created.ID, webhooks.StatusCompleted, "", fixture.now.Add(5*time.Second)); err != nil {
		t.Fatalf("complete webhook delivery: %v", err)
	}
	terminal, claimed, err := fixture.repository.Start(t.Context(), created.ID, fixture.now.Add(6*time.Second), time.Minute)
	if err != nil || claimed || terminal.Status != webhooks.StatusCompleted {
		t.Fatalf("start terminal webhook delivery = (%#v, %v, %v)", terminal, claimed, err)
	}
}

func TestRepositoryConcurrentRecoveryClaimsAreDisjoint(t *testing.T) {
	t.Parallel()
	fixture := newRepositoryFixture(t)
	for _, deliveryID := range []string{"Ev-recovery-a", "Ev-recovery-b"} {
		if _, _, err := fixture.repository.Register(
			t.Context(), fixture.envelope(deliveryID), "ciphertext", fixture.now.Add(time.Hour),
		); err != nil {
			t.Fatalf("register recoverable delivery: %v", err)
		}
	}
	policy := webhooks.DefaultRecoveryPolicy()
	policy.ClaimLimit = 1
	claimAt := fixture.now.Add(10 * time.Minute)

	start := make(chan struct{})
	claimedIDs := make(chan uuid.UUID, 2)
	errorsFound := make(chan error, 2)
	var wait sync.WaitGroup
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			records, err := fixture.repository.ClaimRecoverable(t.Context(), "slack", policy, claimAt)
			if err != nil {
				errorsFound <- err
				return
			}
			if len(records) != 1 {
				errorsFound <- errors.New("recovery claim did not return one delivery")
				return
			}
			claimedIDs <- records[0].ID
		}()
	}
	close(start)
	wait.Wait()
	close(errorsFound)
	for err := range errorsFound {
		t.Fatalf("claim recoverable delivery: %v", err)
	}
	close(claimedIDs)
	unique := make(map[uuid.UUID]struct{})
	for id := range claimedIDs {
		unique[id] = struct{}{}
	}
	if len(unique) != 2 {
		t.Fatalf("concurrent recovery claimed %d unique deliveries, want 2", len(unique))
	}
}

func TestRepositoryExpiresPayloadButRetainsAuditIdentity(t *testing.T) {
	t.Parallel()
	fixture := newRepositoryFixture(t)
	envelope := fixture.envelope("Ev-expire")
	envelope.ReceivedAt = fixture.now.Add(-2 * time.Hour)
	record, _, err := fixture.repository.Register(t.Context(), envelope, "ciphertext", fixture.now.Add(-time.Hour))
	if err != nil {
		t.Fatalf("register expiring webhook delivery: %v", err)
	}
	ids, err := fixture.repository.ExpirePayloads(t.Context(), fixture.now, 10)
	if err != nil || len(ids) != 1 || ids[0] != record.ID {
		t.Fatalf("expire webhook payloads = (%v, %v)", ids, err)
	}
	retained, err := fixture.repository.GetByID(t.Context(), record.ID)
	if err != nil {
		t.Fatalf("get expired webhook delivery: %v", err)
	}
	if retained.EncryptedPayload != nil || retained.DeliveryID != envelope.DeliveryID || retained.WorkspaceID != envelope.WorkspaceID {
		t.Fatalf("expired webhook audit identity = %#v", retained)
	}
}

func TestRepositoryEnforcesWorkspaceForeignKey(t *testing.T) {
	t.Parallel()
	fixture := newRepositoryFixture(t)
	envelope := fixture.envelope("Ev-invalid-workspace")
	envelope.WorkspaceID = uuid.New()
	if _, _, err := fixture.repository.Register(t.Context(), envelope, "ciphertext", fixture.now.Add(time.Hour)); err == nil {
		t.Fatal("register delivery for unknown workspace succeeded")
	}
}
