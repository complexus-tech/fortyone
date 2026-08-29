package slack

import (
	"context"
	"encoding/hex"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type mockNonceStore struct {
	mu      sync.Mutex
	records map[string]nonceRecord
}

func newMockNonceStore() *mockNonceStore {
	return &mockNonceStore{records: make(map[string]nonceRecord)}
}

func nonceStoreKey(provider, purpose string, digest []byte) string {
	return provider + ":" + purpose + ":" + hex.EncodeToString(digest)
}

func (m *mockNonceStore) CreateNonce(_ context.Context, input nonceInput) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.records == nil {
		m.records = make(map[string]nonceRecord)
	}
	var externalWorkspaceID *string
	if input.ExternalWorkspaceID != "" {
		value := input.ExternalWorkspaceID
		externalWorkspaceID = &value
	}
	var externalUserID *string
	if input.ExternalUserID != "" {
		value := input.ExternalUserID
		externalUserID = &value
	}
	m.records[nonceStoreKey(input.Provider, input.Purpose, input.NonceHash)] = nonceRecord{
		ID:                  uuid.New(),
		Provider:            input.Provider,
		Purpose:             input.Purpose,
		WorkspaceID:         input.WorkspaceID,
		UserID:              input.UserID,
		ExternalWorkspaceID: externalWorkspaceID,
		ExternalUserID:      externalUserID,
		Payload:             append([]byte(nil), input.Payload...),
		ExpiresAt:           input.ExpiresAt,
	}
	return nil
}

func (m *mockNonceStore) ConsumeNonce(_ context.Context, input nonceConsumeInput) (nonceRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := nonceStoreKey(input.Provider, input.Purpose, input.NonceHash)
	record, ok := m.records[key]
	if !ok || record.ConsumedAt != nil || !input.Now.Before(record.ExpiresAt) {
		return nonceRecord{}, errMessagingRecordNotFound
	}
	if input.WorkspaceID != nil && record.WorkspaceID != *input.WorkspaceID {
		return nonceRecord{}, errMessagingRecordNotFound
	}
	if input.UserID != nil && record.UserID != nil && *record.UserID != *input.UserID {
		return nonceRecord{}, errMessagingRecordNotFound
	}
	if record.UserID == nil && input.UserID != nil {
		boundUserID := *input.UserID
		record.UserID = &boundUserID
	}
	consumedAt := input.Now
	record.ConsumedAt = &consumedAt
	m.records[key] = record
	return record, nil
}

func TestSlackOAuthStateRejectsPurposeAndWorkspaceMismatchWithoutCrossConsumption(t *testing.T) {
	t.Parallel()

	workspaceID, userID := uuid.New(), uuid.New()
	service := newTestService(&mockRepo{}, &mockRequestStore{}, &mockStoryService{}, Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "https://api.example.com/integrations/slack/setup",
	})
	session, err := service.CreateInstallSession(context.Background(), workspaceID, userID, "acme")
	require.NoError(t, err)
	installURL, err := url.Parse(session.InstallURL)
	require.NoError(t, err)
	state := installURL.Query().Get("state")

	_, err = service.consumeNonce(context.Background(), slackNoncePurposeAccount, state, nil, nil)
	require.Error(t, err)
	wrongWorkspace := uuid.New()
	_, err = service.consumeNonce(context.Background(), slackNoncePurposeOAuth, state, &wrongWorkspace, nil)
	require.Error(t, err)

	record, err := service.consumeNonce(context.Background(), slackNoncePurposeOAuth, state, &workspaceID, &userID)
	require.NoError(t, err)
	require.Equal(t, workspaceID, record.WorkspaceID)
	require.Equal(t, userID, *record.UserID)
}

func TestSlackOAuthStateHasExactlyOneConcurrentConsumer(t *testing.T) {
	t.Parallel()

	workspaceID, userID := uuid.New(), uuid.New()
	service := newTestService(&mockRepo{}, &mockRequestStore{}, &mockStoryService{}, Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "https://api.example.com/integrations/slack/setup",
	})
	session, err := service.CreateInstallSession(context.Background(), workspaceID, userID, "acme")
	require.NoError(t, err)
	installURL, err := url.Parse(session.InstallURL)
	require.NoError(t, err)
	state := installURL.Query().Get("state")

	const consumers = 32
	var successes atomic.Int32
	var wait sync.WaitGroup
	wait.Add(consumers)
	for range consumers {
		go func() {
			defer wait.Done()
			if _, consumeErr := service.consumeNonce(context.Background(), slackNoncePurposeOAuth, state, nil, nil); consumeErr == nil {
				successes.Add(1)
			}
		}()
	}
	wait.Wait()
	require.EqualValues(t, 1, successes.Load())
}
