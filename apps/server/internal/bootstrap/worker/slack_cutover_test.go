package workerbootstrap

import (
	"context"
	"errors"
	"testing"

	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	"github.com/stretchr/testify/require"
)

type slackLegacyBackfillerStub struct {
	credentialCount int
	payloadCount    int
	credentialErr   error
	payloadErr      error
	calls           []string
}

func (stub *slackLegacyBackfillerStub) BackfillLegacyCredentials(context.Context, *slack.LegacyCutover) (int, error) {
	stub.calls = append(stub.calls, "credentials")
	return stub.credentialCount, stub.credentialErr
}

func (stub *slackLegacyBackfillerStub) BackfillLegacyWebhookPayloads(context.Context, *slack.LegacyCutover) (int, error) {
	stub.calls = append(stub.calls, "webhook_payloads")
	return stub.payloadCount, stub.payloadErr
}

func TestBackfillLegacySlackDataEnforcesCredentialThenPayloadOrder(t *testing.T) {
	t.Parallel()

	stub := &slackLegacyBackfillerStub{credentialCount: 3, payloadCount: 5}
	result, err := backfillLegacySlackData(context.Background(), stub, testSlackLegacyCutover(t))
	require.NoError(t, err)
	require.Equal(t, slackLegacyBackfillResult{Credentials: 3, WebhookPayloads: 5}, result)
	require.Equal(t, []string{"credentials", "webhook_payloads"}, stub.calls)
}

func TestBackfillLegacySlackDataStopsBeforePayloadsWhenCredentialsFail(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("credential cutover failed")
	stub := &slackLegacyBackfillerStub{credentialErr: wantErr}
	result, err := backfillLegacySlackData(context.Background(), stub, testSlackLegacyCutover(t))
	require.ErrorIs(t, err, wantErr)
	require.Equal(t, slackLegacyBackfillResult{}, result)
	require.Equal(t, []string{"credentials"}, stub.calls)
}

func TestBackfillLegacySlackDataReturnsCompletedCredentialCountOnPayloadFailure(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("payload cutover failed")
	stub := &slackLegacyBackfillerStub{credentialCount: 3, payloadErr: wantErr}
	result, err := backfillLegacySlackData(context.Background(), stub, testSlackLegacyCutover(t))
	require.ErrorIs(t, err, wantErr)
	require.Equal(t, slackLegacyBackfillResult{Credentials: 3}, result)
	require.Equal(t, []string{"credentials", "webhook_payloads"}, stub.calls)
}

func TestBackfillLegacySlackDataRequiresExplicitCutover(t *testing.T) {
	t.Parallel()

	_, err := backfillLegacySlackData(context.Background(), &slackLegacyBackfillerStub{}, nil)
	require.Error(t, err)
}

func testSlackLegacyCutover(t *testing.T) *slack.LegacyCutover {
	t.Helper()
	cutover, err := slack.NewLegacyCutover("bootstrap-test-legacy-slack-key")
	require.NoError(t, err)
	return cutover
}
