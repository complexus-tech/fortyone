package messaging

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

// StoryMutationInspection is derived from the durable proposal, never the
// model's display preview. DirectEligible permits a trusted native adapter to
// use its invocation transaction instead of presenting a second confirmation.
// It does not execute, claim, or consume the proposal.
type StoryMutationInspection struct {
	Operation      StoryMutationOperation
	StoryID        uuid.UUID
	TeamID         uuid.UUID
	ChangedFields  []string
	DirectEligible bool
	Review         string
}

func (e *FortyOneToolExecutor) InspectStoryMutation(ctx context.Context, scope ToolScope, token string) (StoryMutationInspection, error) {
	if e.mutations == nil || !scope.AllowMutations {
		return StoryMutationInspection{}, ErrMutationNotAllowed
	}
	if err := validateToolScope(&scope); err != nil {
		return StoryMutationInspection{}, err
	}
	if scope.SourceScope == nil {
		return StoryMutationInspection{}, fmt.Errorf("%w: inspection requires a native story source", ErrInvalidConfirmation)
	}
	confirmationID, opaque, err := batchConfirmationID(token)
	if err != nil {
		return StoryMutationInspection{}, err
	}
	if !opaque {
		return StoryMutationInspection{}, fmt.Errorf("%w: native proposal token required", ErrInvalidConfirmation)
	}
	ctx = platformauth.SetUserID(ctx, scope.UserID)
	record, binding, err := e.loadOpaqueConfirmation(ctx, scope, confirmationID, token)
	if err != nil {
		return StoryMutationInspection{}, err
	}
	claims, err := e.nativeProposalClaims(scope, binding, record)
	if err != nil {
		return StoryMutationInspection{}, err
	}
	_, joined, err := e.joinedTeams(ctx, scope)
	if err != nil {
		return StoryMutationInspection{}, err
	}
	_, err = e.sourceStory(ctx, scope, joined, nil, nil)
	if err != nil {
		return StoryMutationInspection{}, err
	}
	if err := e.validateNativeUpdateTargets(ctx, scope, claims); err != nil {
		return StoryMutationInspection{}, err
	}
	return StoryMutationInspection{
		Operation: claims.Operation, StoryID: *claims.StoryID, TeamID: claims.TeamID,
		ChangedFields:  append([]string(nil), claims.ChangedFields...),
		DirectEligible: claims.CommandIntent == "direct_command",
		Review:         claims.Review,
	}, nil
}

func (e *FortyOneToolExecutor) loadOpaqueConfirmation(ctx context.Context, scope ToolScope, id uuid.UUID, token string) (StoryMutationConfirmationRecord, StoryMutationConfirmationBinding, error) {
	binding := StoryMutationConfirmationBinding{ConfirmationID: id, WorkspaceID: scope.WorkspaceID, UserID: scope.UserID, TokenHash: storyMutationTokenHash(token)}
	record, err := e.mutations.store.LoadStoryMutationConfirmation(ctx, binding)
	return record, binding, err
}

func (e *FortyOneToolExecutor) nativeProposalClaims(scope ToolScope, binding StoryMutationConfirmationBinding, record StoryMutationConfirmationRecord) (storyMutationClaims, error) {
	var claims storyMutationClaims
	if record.Operation != StoryMutationUpdate || scope.SourceScope == nil || record.TeamID != scope.SourceScope.TeamID {
		return claims, fmt.Errorf("%w: native story update required", ErrInvalidConfirmation)
	}
	switch record.Status {
	case StoryMutationConfirmationCancelled:
		return claims, ErrCancelledConfirmation
	case StoryMutationConfirmationExpired:
		return claims, ErrExpiredConfirmation
	case StoryMutationConfirmationPending, StoryMutationConfirmationApplied:
	default:
		return claims, ErrInvalidConfirmation
	}
	if len(record.Proposal) == 0 {
		if record.Status == StoryMutationConfirmationApplied {
			return claims, ErrAppliedConfirmation
		}
		return claims, ErrInvalidConfirmation
	}
	decoder := json.NewDecoder(bytes.NewReader(record.Proposal))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&claims); err != nil {
		return claims, ErrInvalidConfirmation
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return claims, ErrInvalidConfirmation
	}
	if claims.Version != storyMutationTokenVersion || claims.ConfirmationID != binding.ConfirmationID || claims.WorkspaceID != scope.WorkspaceID || claims.UserID != scope.UserID || claims.TeamID != record.TeamID {
		return claims, ErrInvalidConfirmation
	}
	if err := validateStoryMutationClaims(claims); err != nil {
		return claims, err
	}
	if err := validateMutationSource(scope, claims); err != nil {
		return claims, err
	}
	if !e.mutations.now().UTC().Before(claims.ExpiresAt) {
		return claims, ErrExpiredConfirmation
	}
	return claims, nil
}

func (e *FortyOneToolExecutor) confirmOpaqueMutation(ctx context.Context, scope ToolScope, id uuid.UUID, token string) (StoryMutationResult, error) {
	record, binding, err := e.loadOpaqueConfirmation(ctx, scope, id, token)
	if err != nil {
		return StoryMutationResult{}, err
	}
	if record.Operation == StoryMutationCreateBatch && scope.SourceScope == nil {
		return e.confirmStoryBatch(ctx, scope, id, token)
	}
	if scope.SourceScope == nil || record.Operation != StoryMutationUpdate || record.TeamID != scope.SourceScope.TeamID {
		return StoryMutationResult{}, ErrInvalidConfirmation
	}
	_, joined, err := e.joinedTeams(ctx, scope)
	if err != nil {
		return StoryMutationResult{}, err
	}
	if _, allowed := joined[record.TeamID]; !allowed {
		return StoryMutationResult{}, ErrTeamNotAccessible
	}
	if record.Status == StoryMutationConfirmationApplied && record.Result != nil && record.LastError == "" {
		if record.Result.StoryID != scope.SourceScope.StoryID || record.Result.TeamID != scope.SourceScope.TeamID {
			return StoryMutationResult{}, ErrInvalidConfirmation
		}
		result := *record.Result
		result.Status = storyMutationStatusAlreadyApplied
		return result, nil
	}
	claims, err := e.nativeProposalClaims(scope, binding, record)
	if err != nil {
		return StoryMutationResult{}, err
	}
	result, duplicate, err := e.mutations.store.ApplyStoryMutationConfirmation(ctx, binding, e.mutations.now().UTC(), func(applyCtx context.Context) (StoryMutationResult, error) {
		_, currentTeams, err := e.joinedTeams(applyCtx, scope)
		if err != nil {
			return StoryMutationResult{}, err
		}
		team, allowed := currentTeams[claims.TeamID]
		if !allowed {
			return StoryMutationResult{}, ErrTeamNotAccessible
		}
		if err := e.validateNativeUpdateTargets(applyCtx, scope, claims); err != nil {
			return StoryMutationResult{}, err
		}
		return e.mutations.confirmUpdate(applyCtx, scope, team, claims)
	})
	if duplicate {
		result.Status = storyMutationStatusAlreadyApplied
	}
	return result, err
}
