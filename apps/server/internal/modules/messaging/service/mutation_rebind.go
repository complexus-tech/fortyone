package messaging

import (
	"context"
	"fmt"
	"slices"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

// RebindStoryMutation accepts the exact change from a native proposal on behalf
// of another participant. The adapter must select a pending parent invocation
// from server state and establish explicit acceptance before calling it. Tokens
// and previousScope are provenance, never model arguments or authority to act
// as the previous requester. This method creates a proposal; ConfirmStoryMutation
// remains the sole write boundary and rechecks the responding user's access.
func (e *FortyOneToolExecutor) RebindStoryMutation(ctx context.Context, previousScope, scope ToolScope, token string) (StoryMutationConfirmation, error) {
	if e.mutations == nil || !scope.AllowMutations {
		return StoryMutationConfirmation{}, ErrMutationNotAllowed
	}
	if err := validateToolScope(&previousScope); err != nil {
		return StoryMutationConfirmation{}, err
	}
	if err := validateToolScope(&scope); err != nil {
		return StoryMutationConfirmation{}, err
	}
	if previousScope.SourceScope == nil || scope.SourceScope == nil || previousScope.WorkspaceID != scope.WorkspaceID || *previousScope.SourceScope != *scope.SourceScope {
		return StoryMutationConfirmation{}, fmt.Errorf("%w: proposal acceptance must stay in the same source story", ErrInvalidConfirmation)
	}
	id, opaque, err := batchConfirmationID(token)
	if err != nil {
		return StoryMutationConfirmation{}, err
	}
	if !opaque {
		return StoryMutationConfirmation{}, ErrInvalidConfirmation
	}
	ctx = platformauth.SetUserID(ctx, scope.UserID)
	record, binding, err := e.loadOpaqueConfirmation(ctx, previousScope, id, token)
	if err != nil {
		return StoryMutationConfirmation{}, err
	}
	claims, err := e.nativeProposalClaims(previousScope, binding, record)
	if err != nil {
		return StoryMutationConfirmation{}, err
	}
	if record.Status != StoryMutationConfirmationPending {
		return StoryMutationConfirmation{}, ErrAppliedConfirmation
	}
	_, joined, err := e.joinedTeams(ctx, scope)
	if err != nil {
		return StoryMutationConfirmation{}, err
	}
	story, err := e.sourceStory(ctx, scope, joined, nil, nil)
	if err != nil {
		return StoryMutationConfirmation{}, err
	}
	if claims.ExpectedUpdatedAt == nil || !story.UpdatedAt.Equal(*claims.ExpectedUpdatedAt) {
		return StoryMutationConfirmation{}, ErrStaleMutation
	}
	// "Me" was resolved when the original proposal was shown. Rebinding must
	// never silently change that proposed assignee to the approving participant.
	if claims.AssigneeAction == assigneeActionMe {
		claims.AssigneeAction = assigneeActionUnchanged
		if slices.Contains(claims.ChangedFields, "assignee_id") {
			claims.AssigneeAction = assigneeActionNamed
			assigneeID := claims.UserID
			claims.AssigneeID = &assigneeID
		}
	}
	claims.UserID = scope.UserID
	claims.SourceScope = scope.SourceScope
	claims.CommandIntent = "direct_command"
	claims.ConfirmationID, err = uuid.NewRandomFromReader(e.mutations.random)
	if err != nil {
		return StoryMutationConfirmation{}, fmt.Errorf("generate accepted proposal ID: %w", err)
	}
	if err := validateStoryMutationClaims(claims); err != nil {
		return StoryMutationConfirmation{}, err
	}
	if err := e.validateNativeUpdateTargets(ctx, scope, claims); err != nil {
		return StoryMutationConfirmation{}, err
	}
	team := joined[scope.SourceScope.TeamID]
	preview := StoryMutationPreview{
		StoryID: claims.StoryID, Reference: storyReference(team.Code, story.SequenceID),
		TeamID: team.ID, TeamName: team.Name, TeamCode: team.Code, Title: story.Title,
		ChangedFields:  append([]string(nil), claims.ChangedFields...),
		AssigneeAction: claims.AssigneeAction, AssigneeID: claims.AssigneeID,
	}
	if claims.Title != nil {
		preview.Title = *claims.Title
	}
	output, err := e.mutations.marshalProposal(ctx, claims, preview, fmt.Sprintf("Update %s?", preview.Reference))
	if err != nil {
		return StoryMutationConfirmation{}, err
	}
	confirmation, proposed, err := mutationConfirmationFromToolResult(output)
	if err != nil {
		return StoryMutationConfirmation{}, err
	}
	if !proposed || confirmation == nil {
		return StoryMutationConfirmation{}, ErrInvalidConfirmation
	}
	return *confirmation, nil
}
