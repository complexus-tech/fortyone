package slack

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func uuidSubset(required, available []uuid.UUID) bool {
	if len(required) == 0 {
		return false
	}
	set := make(map[uuid.UUID]struct{}, len(available))
	for _, id := range available {
		if id != uuid.Nil {
			set[id] = struct{}{}
		}
	}
	for _, id := range required {
		if _, ok := set[id]; !ok {
			return false
		}
	}
	return true
}

func (p *EventProcessor) slackChannelDeliveryAuthorizationCurrent(
	ctx context.Context,
	workspaceID uuid.UUID,
	installation slackWorkspaceRecord,
	userID uuid.UUID,
	channelID string,
	externalRecipientUserID string,
	providerPayload SlackProviderPayload,
) (bool, error) {
	authorization := providerPayload.Authorization
	channelID = strings.TrimSpace(channelID)
	if authorization == nil {
		// A DM actor with no joined teams has no workspace data scope to freeze.
		// The assistant path already revalidates the linked actor and recipient;
		// an unscoped public-channel delivery is never allowed.
		return strings.HasPrefix(strings.ToUpper(channelID), "D"), nil
	}
	if authorization.ActorUserID == nil || *authorization.ActorUserID != userID {
		return false, nil
	}
	memberships, ok := p.repo.(eventTeamMembershipRepository)
	if !ok {
		return false, errors.New("slack team membership repository is not configured")
	}
	actorTeams, err := memberships.ListWorkspaceTeamsForUser(ctx, workspaceID, userID)
	if err != nil {
		return false, err
	}
	if !uuidSubset(authorization.AllowedTeamIDs, slackTeamRecordIDs(actorTeams)) {
		return false, nil
	}
	if strings.HasPrefix(strings.ToUpper(channelID), "D") {
		externalRecipientUserID = strings.TrimSpace(externalRecipientUserID)
		if externalRecipientUserID == "" {
			return false, nil
		}
		recipientUserID, err := p.repo.FindLinkedUserIDBySlackUser(ctx, workspaceID, installation.SlackTeamID, externalRecipientUserID)
		if err != nil {
			return false, err
		}
		if recipientUserID == nil || *recipientUserID == uuid.Nil {
			return false, nil
		}
		recipientTeams, err := memberships.ListWorkspaceTeamsForUser(ctx, workspaceID, *recipientUserID)
		if err != nil {
			return false, err
		}
		return uuidSubset(authorization.AllowedTeamIDs, slackTeamRecordIDs(recipientTeams)), nil
	}
	if authorization.Scope == slackDeliveryAuthorizationScopeActorMembership {
		return true, nil
	}
	if channelID == "" {
		return false, nil
	}
	repository, ok := p.repo.(eventAssistantChannelAudienceRepository)
	if !ok {
		return false, errors.New("slack assistant channel audience repository is not configured")
	}
	currentScope, err := repository.GetAuthorizedAssistantChannelTeamScope(
		ctx,
		workspaceID,
		installation.ID,
		channelID,
		userID,
	)
	if err != nil {
		return false, err
	}
	if !uuidSubset(authorization.AllowedTeamIDs, currentScope.AllowedTeamIDs) {
		return false, nil
	}
	if len(authorization.SharedTeamIDs) > 0 && !uuidSubset(authorization.SharedTeamIDs, currentScope.SharedTeamIDs) {
		return false, nil
	}
	return true, nil
}

func assistantSettingsAllowDelivery(settings CoreSlackAgentSettings, payload SlackProviderPayload) bool {
	return true
}

func validateOutboundInstallation(record outboundDeliveryRecord, installation slackWorkspaceRecord) error {
	externalWorkspaceID := strings.TrimSpace(record.ExternalWorkspaceID)
	if externalWorkspaceID == "" {
		return errors.New("slack outbound delivery has no external workspace binding")
	}
	if installation.WorkspaceID != record.WorkspaceID {
		return fmt.Errorf(
			"slack installation workspace mismatch for external workspace %q: delivery workspace %s, installation workspace %s",
			externalWorkspaceID,
			record.WorkspaceID,
			installation.WorkspaceID,
		)
	}
	if strings.TrimSpace(installation.SlackTeamID) != externalWorkspaceID {
		return fmt.Errorf(
			"slack installation team mismatch: delivery team %q, installation team %q",
			externalWorkspaceID,
			installation.SlackTeamID,
		)
	}
	if record.InstallGeneration == nil || *record.InstallGeneration == uuid.Nil || *record.InstallGeneration != installation.InstallGeneration {
		return fmt.Errorf("slack installation generation mismatch for external workspace %q", externalWorkspaceID)
	}
	if !installation.IsActive {
		return fmt.Errorf("slack installation for external workspace %q is inactive", externalWorkspaceID)
	}
	return nil
}
