package slackdomain

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestHumanSlackCommandsRequireExplicitActorAndTenant(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	workspaceID := uuid.New()
	actorID := uuid.New()
	installationID := uuid.New()
	generation := uuid.New()

	tests := map[string]struct {
		validate func() error
	}{
		"missing actor": {
			validate: func() error {
				return (WorkspaceActorQuery{WorkspaceID: workspaceID}).Validate()
			},
		},
		"missing workspace": {
			validate: func() error {
				return (WorkspaceActorQuery{ActorID: actorID}).Validate()
			},
		},
		"missing install time": {
			validate: func() error {
				return (UpsertInstallationCommand{
					WorkspaceID: workspaceID,
					ActorID:     actorID,
					Installation: OAuthInstallation{
						SlackTeamID:       "T1",
						InstallGeneration: generation,
					},
				}).Validate()
			},
		},
		"missing sync generation": {
			validate: func() error {
				return (SyncChannelsCommand{
					WorkspaceID:    workspaceID,
					ActorID:        actorID,
					InstallationID: installationID,
					Now:            now,
				}).Validate()
			},
		},
		"invalid request log limit": {
			validate: func() error {
				return (ListRequestLogsQuery{WorkspaceID: workspaceID, ActorID: actorID, Limit: 201}).Validate()
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			require.ErrorIs(t, test.validate(), ErrInvalidInput)
		})
	}
}

func TestSlackMutationCommandsValidateCompleteIntent(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	workspaceID := uuid.New()
	actorID := uuid.New()
	installationID := uuid.New()
	generation := uuid.New()

	require.NoError(t, (SyncChannelsCommand{
		WorkspaceID:            workspaceID,
		ActorID:                actorID,
		InstallationID:         installationID,
		InstallationGeneration: generation,
		Channels: []ChannelUpsert{{
			SlackChannelID: " C1 ",
			Name:           " general ",
		}},
		Now: now,
	}).Validate())
	require.ErrorIs(t, (UpdateAgentSettingsCommand{
		WorkspaceID: workspaceID,
		ActorID:     actorID,
		Guidance:    strings.Repeat("x", MaxAgentGuidanceRunes+1),
		Now:         now,
	}).Validate(), ErrInvalidInput)
	require.ErrorIs(t, (ReplaceChannelAudienceCommand{
		WorkspaceID:            workspaceID,
		ActorID:                actorID,
		InstallationID:         installationID,
		InstallationGeneration: generation,
		SlackChannelID:         "C1",
		Configured:             true,
		TeamIDs:                []uuid.UUID{uuid.Nil},
		Now:                    now,
	}).Validate(), ErrInvalidInput)
}
