//go:build integration

package figmarepository

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"sync"
	"testing"
	"time"

	figmaprovider "github.com/complexus-tech/projects-api/internal/modules/figma"
	figmadomain "github.com/complexus-tech/projects-api/internal/modules/figma/domain"
	figma "github.com/complexus-tech/projects-api/internal/modules/figma/service"
	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	webhooksrepository "github.com/complexus-tech/projects-api/internal/platform/webhooks/repository"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type figmaRepositoryFixture struct {
	database   *testkit.Postgres
	repository *Repository
	workspaceA uuid.UUID
	workspaceB uuid.UUID
	userA      uuid.UUID
	userB      uuid.UUID
	storyA     uuid.UUID
}

func newFigmaRepositoryFixture(t *testing.T) figmaRepositoryFixture {
	t.Helper()
	database := testkit.NewPostgres(t)
	fixture := figmaRepositoryFixture{
		database: database, repository: New(database.Pool),
		workspaceA: uuid.New(), workspaceB: uuid.New(),
		userA: uuid.New(), userB: uuid.New(), storyA: uuid.New(),
	}
	teamA, teamB, suffix := uuid.New(), uuid.New(), uuid.NewString()
	ctx := t.Context()
	_, err := database.Pool.Exec(ctx, `
		INSERT INTO users (user_id, username, email, full_name)
		VALUES
			($1, $2, $3, 'Figma repository A'),
			($4, $5, $6, 'Figma repository B')
	`, fixture.userA, "figma-a-"+suffix, "figma-a-"+suffix+"@example.test",
		fixture.userB, "figma-b-"+suffix, "figma-b-"+suffix+"@example.test")
	require.NoError(t, err)
	_, err = database.Pool.Exec(ctx, `
		INSERT INTO workspaces (workspace_id, name, slug, created_by)
		VALUES
			($1, 'Figma repository A', $2, $3),
			($4, 'Figma repository B', $5, $6)
	`, fixture.workspaceA, "figma-a-"+suffix, fixture.userA,
		fixture.workspaceB, "figma-b-"+suffix, fixture.userB)
	require.NoError(t, err)
	_, err = database.Pool.Exec(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, 'admin'), ($3, $4, 'admin')
	`, fixture.workspaceA, fixture.userA, fixture.workspaceB, fixture.userB)
	require.NoError(t, err)
	_, err = database.Pool.Exec(ctx, `
		INSERT INTO teams (team_id, name, workspace_id, code, color)
		VALUES
			($1, 'Figma A', $2, $3, '#000000'),
			($4, 'Figma B', $5, $6, '#111111')
	`, teamA, fixture.workspaceA, "FGA"+suffix[:5],
		teamB, fixture.workspaceB, "FGB"+suffix[:5])
	require.NoError(t, err)
	_, err = database.Pool.Exec(ctx, `
		INSERT INTO stories (id, team_id, title, workspace_id)
		VALUES ($1, $2, 'Figma-linked story', $3)
	`, fixture.storyA, teamA, fixture.workspaceA)
	require.NoError(t, err)
	return fixture
}

func (fixture figmaRepositoryFixture) connection(
	workspaceID, userID uuid.UUID,
) figmadomain.Connection {
	return figmadomain.Connection{
		ID: uuid.New(), WorkspaceID: workspaceID, FigmaUserID: uuid.NewString(),
		CredentialPayload: "vault.v2.test-envelope", CredentialVersion: 2,
		InstallationGeneration: uuid.New(), Scopes: []string{"file_content:read"},
		ExpiresAt: time.Now().UTC().Add(time.Hour), ConnectedByUserID: userID,
		IsActive: true,
	}
}

func TestFigmaConnectionMutationIsAtomicAndGenerationFenced(t *testing.T) {
	fixture := newFigmaRepositoryFixture(t)
	ctx := t.Context()
	connection := fixture.connection(fixture.workspaceA, fixture.userA)
	created, err := fixture.repository.UpsertConnection(ctx, connection)
	require.NoError(t, err)
	require.Equal(t, connection.ID, created.ID)
	require.Equal(t, connection.InstallationGeneration, created.InstallationGeneration)

	webhook := figmadomain.Webhook{
		ConnectionID: created.ID, FileKey: "product-file", EventType: figma.EventFileUpdate,
		FigmaWebhookID: 91001, PasscodeHash: "redacted-passcode-digest", IsActive: true,
	}
	require.NoError(t, fixture.repository.SaveWebhook(ctx, webhook))

	rejected := fixture.connection(fixture.workspaceA, fixture.userB)
	_, err = fixture.repository.UpsertConnection(ctx, rejected)
	require.ErrorIs(t, err, figmadomain.ErrNotFound)
	retained, err := fixture.repository.GetConnection(ctx, fixture.workspaceA)
	require.NoError(t, err)
	require.Equal(t, created.ID, retained.ID, "failed reconnect must roll back deactivation")
	_, err = fixture.repository.GetWebhook(ctx, webhook.FigmaWebhookID)
	require.NoError(t, err, "failed reconnect must retain the existing webhook grant")

	const wrongGenerationPayload = "vault.v2.wrong-generation"
	replaced, err := fixture.repository.UpdateConnectionCredential(
		ctx, created.ID, uuid.New(), created.CredentialPayload,
		wrongGenerationPayload, created.ExpiresAt.Add(time.Hour),
	)
	require.NoError(t, err)
	require.False(t, replaced)
	replaced, err = fixture.repository.UpdateConnectionCredential(
		ctx, created.ID, created.InstallationGeneration, created.CredentialPayload,
		"vault.v2.refreshed", created.ExpiresAt.Add(time.Hour),
	)
	require.NoError(t, err)
	require.True(t, replaced)
	replaced, err = fixture.repository.UpdateConnectionCredential(
		ctx, created.ID, created.InstallationGeneration, created.CredentialPayload,
		"vault.v2.stale-refresh", created.ExpiresAt.Add(2*time.Hour),
	)
	require.NoError(t, err)
	require.False(t, replaced, "stale refresh must lose the optimistic CAS")

	reconnected := fixture.connection(fixture.workspaceA, fixture.userA)
	current, err := fixture.repository.UpsertConnection(ctx, reconnected)
	require.NoError(t, err)
	require.Equal(t, reconnected.InstallationGeneration, current.InstallationGeneration)
	_, err = fixture.repository.GetCurrentWebhook(
		ctx, created.ID, created.InstallationGeneration, webhook.FigmaWebhookID,
	)
	require.ErrorIs(t, err, figmadomain.ErrNotFound)

	_, err = fixture.database.Pool.Exec(ctx, `SET enable_seqscan = off`)
	require.NoError(t, err)
	rows, err := fixture.database.Pool.Query(ctx, `
		EXPLAIN (COSTS OFF)
		SELECT id
		FROM figma_connections
		WHERE workspace_id = $1
		  AND installation_generation = $2
		  AND is_active
	`, fixture.workspaceA, current.InstallationGeneration)
	require.NoError(t, err)
	var plan strings.Builder
	for rows.Next() {
		var line string
		require.NoError(t, rows.Scan(&line))
		plan.WriteString(line)
	}
	require.NoError(t, rows.Err())
	rows.Close()
	_, _ = fixture.database.Pool.Exec(ctx, `RESET enable_seqscan`)
	require.Contains(t, plan.String(), "figma_connections_active_generation")
}

func TestFigmaStoryLinksCannotCrossWorkspaceBoundaries(t *testing.T) {
	fixture := newFigmaRepositoryFixture(t)
	ctx := t.Context()
	nodeID, nodeName := "12:34", "Checkout"
	link := figmadomain.StoryLink{
		WorkspaceID: fixture.workspaceA, StoryID: fixture.storyA,
		CreatedByUserID: fixture.userA,
		Artifact: figmadomain.Artifact{
			FileKey: "product-file", NodeID: &nodeID,
			OriginalURL:  "https://www.figma.com/design/product?node-id=12-34",
			CanonicalURL: "https://www.figma.com/design/product?node-id=12-34",
			FileName:     "Product", NodeName: &nodeName,
		},
	}
	created, err := fixture.repository.UpsertStoryLink(ctx, link)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, created.ID)
	require.NotNil(t, created.StoryLinkID)

	rejected := link
	rejected.WorkspaceID = fixture.workspaceB
	rejected.CreatedByUserID = fixture.userB
	_, err = fixture.repository.UpsertStoryLink(ctx, rejected)
	require.ErrorIs(t, err, figmadomain.ErrNotFound)
	var genericCount int
	require.NoError(t, fixture.database.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM story_links WHERE story_id = $1
	`, fixture.storyA).Scan(&genericCount))
	require.Equal(t, 1, genericCount, "failed cross-tenant upsert must roll back its generic link")

	_, err = fixture.repository.GetStoryLink(ctx, fixture.workspaceB, created.ID)
	require.ErrorIs(t, err, figmadomain.ErrNotFound)
	crossTenantUpdate := created
	crossTenantUpdate.WorkspaceID = fixture.workspaceB
	require.ErrorIs(t, fixture.repository.UpdateStoryLink(ctx, crossTenantUpdate), figmadomain.ErrNotFound)
	_, err = fixture.repository.DeleteStoryLink(ctx, fixture.workspaceB, created.ID)
	require.ErrorIs(t, err, figmadomain.ErrNotFound)

	deleted, err := fixture.repository.DeleteStoryLink(ctx, fixture.workspaceA, created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, deleted.ID)
	require.NoError(t, fixture.database.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM story_links WHERE story_id = $1
	`, fixture.storyA).Scan(&genericCount))
	require.Zero(t, genericCount)
}

type concurrentFigmaQueue struct {
	mutex    sync.Mutex
	payloads []tasks.FigmaWebhookPayload
}

func (queue *concurrentFigmaQueue) EnqueueFigmaWebhook(
	_ context.Context,
	payload tasks.FigmaWebhookPayload,
) error {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	queue.payloads = append(queue.payloads, payload)
	return nil
}

func TestFigmaWebhookGatewayPersistsOneEncryptedTenantBoundDelivery(t *testing.T) {
	fixture := newFigmaRepositoryFixture(t)
	ctx := t.Context()
	connection, err := fixture.repository.UpsertConnection(
		ctx,
		fixture.connection(fixture.workspaceA, fixture.userA),
	)
	require.NoError(t, err)
	const webhookID int64 = 92001
	const passcode = "provider-passcode-never-persisted"
	require.NoError(t, fixture.repository.SaveWebhook(ctx, figmadomain.Webhook{
		ConnectionID: connection.ID, FileKey: "product-file",
		EventType: figma.EventFileUpdate, FigmaWebhookID: webhookID,
		PasscodeHash: figmaPasscodeDigest(passcode), IsActive: true,
	}))

	queue := &concurrentFigmaQueue{}
	runtime, err := figma.NewWebhookRuntime(fixture.repository, queue, figma.Config{
		WebhookPayloadSecret: "dedicated-test-payload-secret",
	})
	require.NoError(t, err)
	catalog, err := integrations.NewRegistry(figmaprovider.ProviderDescriptor())
	require.NoError(t, err)
	runtimes, err := webhooks.NewRuntimeRegistry(catalog, runtime.Registration)
	require.NoError(t, err)
	inbox := webhooksrepository.New(fixture.database.Pool)
	gateway, err := webhooks.NewGateway(inbox, runtimes, webhooks.Config{})
	require.NoError(t, err)
	body := []byte(`{
		"event_type":"FILE_UPDATE",
		"file_key":"product-file",
		"file_name":"Product",
		"passcode":"` + passcode + `",
		"timestamp":"2026-08-28T12:00:00Z",
		"webhook_id":92001
	}`)

	const deliveries = 8
	start := make(chan struct{})
	receipts := make(chan webhooks.Receipt, deliveries)
	errorsFound := make(chan error, deliveries)
	var waitGroup sync.WaitGroup
	for range deliveries {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			receipt, receiveErr := gateway.Receive(context.Background(), figmaprovider.ProviderKey, webhooks.SignedRequest{
				Method: "POST", Body: body,
			})
			if receiveErr != nil {
				errorsFound <- receiveErr
				return
			}
			receipts <- receipt
		}()
	}
	close(start)
	waitGroup.Wait()
	close(errorsFound)
	for receiveErr := range errorsFound {
		require.NoError(t, receiveErr)
	}
	close(receipts)
	var inboxID uuid.UUID
	for receipt := range receipts {
		require.True(t, receipt.Queued)
		if inboxID == uuid.Nil {
			inboxID = receipt.ID
		}
		require.Equal(t, inboxID, receipt.ID)
	}
	require.NotEqual(t, uuid.Nil, inboxID)
	record, err := inbox.GetByID(ctx, inboxID)
	require.NoError(t, err)
	require.Equal(t, fixture.workspaceA, record.WorkspaceID)
	require.Equal(t, connection.ID, record.InstallationID)
	require.Equal(t, connection.InstallationGeneration, record.InstallationGeneration)
	require.NotNil(t, record.EncryptedPayload)
	require.NotContains(t, *record.EncryptedPayload, passcode)
	require.NotContains(t, *record.EncryptedPayload, "product-file")

	queue.mutex.Lock()
	require.Len(t, queue.payloads, deliveries)
	for _, payload := range queue.payloads {
		require.Equal(t, inboxID, payload.InboxID)
	}
	queue.mutex.Unlock()
}

func figmaPasscodeDigest(value string) string {
	// The production verifier stores the hexadecimal SHA-256 digest; computing
	// it here keeps the integration fixture independent of unexported helpers.
	digest := sha256.Sum256([]byte(value))
	return base64.RawURLEncoding.EncodeToString(digest[:])
}

var _ figma.WebhookQueue = (*concurrentFigmaQueue)(nil)
