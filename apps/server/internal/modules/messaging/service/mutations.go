package messaging

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

const (
	storyMutationConfirmationKind = "story_mutation_confirmation_required"
	storyMutationTokenVersion     = 1
	storyMutationConfirmationTTL  = 10 * time.Minute
	maximumStoryMutationTokenSize = 2_000
	maximumStoryTitleRunes        = 255

	storyMutationStatusApplied        = "applied"
	storyMutationStatusAlreadyApplied = "already_applied"

	assigneeActionUnchanged  = "unchanged"
	assigneeActionMe         = "me"
	assigneeActionUnassigned = "unassigned"
)

var storyPriorities = map[string]struct{}{
	"No Priority": {},
	"Low":         {},
	"Medium":      {},
	"High":        {},
	"Urgent":      {},
}

type storyMutationExecutor struct {
	stories StoryMutationService
	store   StoryMutationConfirmationStore
	key     []byte
	now     func() time.Time
	random  io.Reader
}

type storyMutationClaims struct {
	Version           int                    `json:"v"`
	ConfirmationID    uuid.UUID              `json:"i"`
	Operation         StoryMutationOperation `json:"o"`
	WorkspaceID       uuid.UUID              `json:"w"`
	UserID            uuid.UUID              `json:"u"`
	TeamID            uuid.UUID              `json:"t"`
	StoryID           *uuid.UUID             `json:"s,omitempty"`
	ExpectedUpdatedAt *time.Time             `json:"e,omitempty"`
	Title             *string                `json:"n,omitempty"`
	Priority          *string                `json:"p,omitempty"`
	AssigneeAction    string                 `json:"a"`
	ExpiresAt         time.Time              `json:"x"`
}

type storyMutationConfirmationToolResult struct {
	Kind         string                    `json:"kind"`
	Confirmation StoryMutationConfirmation `json:"confirmation"`
}

func newStoryMutationExecutor(
	storiesService StoryMutationService,
	secret string,
	store StoryMutationConfirmationStore,
) *storyMutationExecutor {
	key := sha256.Sum256([]byte("fortyone:messaging:story-mutation:v1\x00" + secret))
	return &storyMutationExecutor{
		stories: storiesService,
		store:   store,
		key:     append([]byte(nil), key[:]...),
		now:     time.Now,
		random:  rand.Reader,
	}
}

func (m *storyMutationExecutor) proposeCreate(
	ctx context.Context,
	executor *FortyOneToolExecutor,
	scope ToolScope,
	raw json.RawMessage,
) (json.RawMessage, error) {
	var args struct {
		TeamID   string  `json:"team_id"`
		Title    string  `json:"title"`
		Priority *string `json:"priority"`
		Assignee string  `json:"assignee"`
	}
	if err := decodeToolArguments(raw, &args, "team_id", "title", "priority", "assignee"); err != nil {
		return nil, err
	}

	_, joinedByID, err := executor.joinedTeams(ctx, scope)
	if err != nil {
		return nil, err
	}
	teamID, err := parseAccessibleTeamID(args.TeamID, joinedByID)
	if err != nil {
		return nil, err
	}
	title, err := normalizedStoryTitle(args.Title)
	if err != nil {
		return nil, err
	}
	priority, err := normalizedStoryPriority(args.Priority, "No Priority")
	if err != nil {
		return nil, err
	}
	if args.Assignee != assigneeActionMe && args.Assignee != assigneeActionUnassigned {
		return nil, fmt.Errorf("%w: assignee must be me or unassigned", ErrInvalidToolArguments)
	}

	team := joinedByID[teamID]
	confirmationID, err := uuid.NewRandomFromReader(m.random)
	if err != nil {
		return nil, fmt.Errorf("generate story mutation confirmation ID: %w", err)
	}
	now := m.now().UTC()
	claims := storyMutationClaims{
		Version:        storyMutationTokenVersion,
		ConfirmationID: confirmationID,
		Operation:      StoryMutationCreate,
		WorkspaceID:    scope.WorkspaceID,
		UserID:         scope.UserID,
		TeamID:         teamID,
		Title:          &title,
		Priority:       &priority,
		AssigneeAction: args.Assignee,
		ExpiresAt:      now.Add(storyMutationConfirmationTTL),
	}
	return m.marshalProposal(ctx, claims, StoryMutationPreview{
		TeamID:         team.ID,
		TeamName:       team.Name,
		TeamCode:       strings.ToUpper(team.Code),
		Title:          title,
		Priority:       &priority,
		AssigneeAction: args.Assignee,
	}, fmt.Sprintf("Create %q in %s (%s)?", title, team.Name, strings.ToUpper(team.Code)))
}

func (m *storyMutationExecutor) proposeUpdate(
	ctx context.Context,
	executor *FortyOneToolExecutor,
	scope ToolScope,
	raw json.RawMessage,
) (json.RawMessage, error) {
	var args struct {
		StoryID        *string `json:"story_id"`
		StoryReference *string `json:"story_reference"`
		Title          *string `json:"title"`
		Priority       *string `json:"priority"`
		Assignee       string  `json:"assignee"`
	}
	if err := decodeToolArguments(raw, &args, "story_id", "story_reference", "title", "priority", "assignee"); err != nil {
		return nil, err
	}
	if (args.StoryID == nil) == (args.StoryReference == nil) {
		return nil, fmt.Errorf("%w: provide exactly one of story_id or story_reference", ErrInvalidToolArguments)
	}
	if args.Assignee != assigneeActionUnchanged && args.Assignee != assigneeActionMe && args.Assignee != assigneeActionUnassigned {
		return nil, fmt.Errorf("%w: assignee must be unchanged, me, or unassigned", ErrInvalidToolArguments)
	}

	_, joinedByID, err := executor.joinedTeams(ctx, scope)
	if err != nil {
		return nil, err
	}
	story, err := m.resolveUpdateStory(ctx, scope, joinedByID, args.StoryID, args.StoryReference)
	if err != nil {
		return nil, err
	}
	team, allowed := joinedByID[story.Team]
	if !allowed || story.Workspace != scope.WorkspaceID {
		return nil, fmt.Errorf("%w: %s", ErrTeamNotAccessible, story.Team)
	}

	var title *string
	if args.Title != nil {
		value, err := normalizedStoryTitle(*args.Title)
		if err != nil {
			return nil, err
		}
		title = &value
	}
	var priority *string
	if args.Priority != nil {
		value, err := normalizedStoryPriority(args.Priority, "")
		if err != nil {
			return nil, err
		}
		priority = &value
	}

	changedFields := proposedChangedFields(story, scope.UserID, title, priority, args.Assignee)
	if len(changedFields) == 0 {
		return nil, fmt.Errorf("%w: update_story must contain at least one effective change", ErrInvalidToolArguments)
	}

	confirmationID, err := uuid.NewRandomFromReader(m.random)
	if err != nil {
		return nil, fmt.Errorf("generate story mutation confirmation ID: %w", err)
	}
	now := m.now().UTC()
	expectedUpdatedAt := story.UpdatedAt.UTC()
	claims := storyMutationClaims{
		Version:           storyMutationTokenVersion,
		ConfirmationID:    confirmationID,
		Operation:         StoryMutationUpdate,
		WorkspaceID:       scope.WorkspaceID,
		UserID:            scope.UserID,
		TeamID:            story.Team,
		StoryID:           &story.ID,
		ExpectedUpdatedAt: &expectedUpdatedAt,
		Title:             title,
		Priority:          priority,
		AssigneeAction:    args.Assignee,
		ExpiresAt:         now.Add(storyMutationConfirmationTTL),
	}
	previewTitle := story.Title
	if title != nil {
		previewTitle = *title
	}
	reference := storyReference(team.Code, story.SequenceID)
	storyIDCopy := story.ID
	return m.marshalProposal(ctx, claims, StoryMutationPreview{
		StoryID:        &storyIDCopy,
		Reference:      reference,
		TeamID:         team.ID,
		TeamName:       team.Name,
		TeamCode:       strings.ToUpper(team.Code),
		Title:          previewTitle,
		Priority:       priority,
		AssigneeAction: args.Assignee,
		ChangedFields:  changedFields,
	}, fmt.Sprintf("Update %s?", reference))
}

func (m *storyMutationExecutor) resolveUpdateStory(
	ctx context.Context,
	scope ToolScope,
	joinedByID map[uuid.UUID]teams.CoreTeam,
	storyIDRaw, storyReferenceRaw *string,
) (stories.CoreSingleStory, error) {
	if storyIDRaw != nil {
		storyID, err := parseRequiredUUID(*storyIDRaw, "story_id")
		if err != nil {
			return stories.CoreSingleStory{}, err
		}
		story, err := m.stories.Get(ctx, storyID, scope.WorkspaceID)
		if err != nil {
			return stories.CoreSingleStory{}, fmt.Errorf("load story for update proposal: %w", err)
		}
		return story, nil
	}

	reference, teamCode, err := normalizedStoryReference(*storyReferenceRaw)
	if err != nil {
		return stories.CoreSingleStory{}, err
	}
	var expectedTeamID uuid.UUID
	for teamID, team := range joinedByID {
		if strings.EqualFold(team.Code, teamCode) {
			expectedTeamID = teamID
			break
		}
	}
	if expectedTeamID == uuid.Nil {
		return stories.CoreSingleStory{}, fmt.Errorf("%w: team code %s", ErrTeamNotAccessible, teamCode)
	}
	story, err := m.stories.QueryByRef(ctx, scope.WorkspaceID, reference)
	if err != nil {
		return stories.CoreSingleStory{}, fmt.Errorf("load story %s for update proposal: %w", reference, err)
	}
	if story.Team != expectedTeamID {
		return stories.CoreSingleStory{}, fmt.Errorf("%w: story team does not match reference", ErrInvalidToolArguments)
	}
	return story, nil
}

func (m *storyMutationExecutor) marshalProposal(
	ctx context.Context,
	claims storyMutationClaims,
	preview StoryMutationPreview,
	prompt string,
) (json.RawMessage, error) {
	token, err := m.signClaims(claims)
	if err != nil {
		return nil, err
	}
	if err := m.store.RegisterStoryMutationConfirmation(ctx, StoryMutationConfirmationStateInput{
		ConfirmationID: claims.ConfirmationID,
		WorkspaceID:    claims.WorkspaceID,
		UserID:         claims.UserID,
		TeamID:         claims.TeamID,
		Operation:      claims.Operation,
		TokenHash:      storyMutationTokenHash(token),
		ExpiresAt:      claims.ExpiresAt,
	}); err != nil {
		return nil, fmt.Errorf("register story mutation confirmation: %w", err)
	}
	return marshalToolResult(storyMutationConfirmationToolResult{
		Kind: storyMutationConfirmationKind,
		Confirmation: StoryMutationConfirmation{
			Operation: claims.Operation,
			Token:     token,
			ExpiresAt: claims.ExpiresAt,
			Prompt:    prompt,
			Story:     preview,
		},
	})
}

// ConfirmStoryMutation applies a signed proposal after an explicit provider
// confirmation. It re-authorizes the actor, workspace, joined-team membership,
// and the current channel audience before every write.
func (e *FortyOneToolExecutor) ConfirmStoryMutation(ctx context.Context, scope ToolScope, token string) (StoryMutationResult, error) {
	if e.mutations == nil {
		return StoryMutationResult{}, fmt.Errorf("%w: story mutations are disabled", ErrInvalidConfirmation)
	}
	if err := validateToolScope(&scope); err != nil {
		return StoryMutationResult{}, err
	}
	if !scope.AllowMutations {
		return StoryMutationResult{}, ErrMutationNotAllowed
	}
	ctx = platformauth.SetUserID(ctx, scope.UserID)
	claims, err := e.mutations.verifyClaims(token)
	if err != nil {
		return StoryMutationResult{}, err
	}
	if claims.WorkspaceID != scope.WorkspaceID || claims.UserID != scope.UserID {
		return StoryMutationResult{}, fmt.Errorf("%w: confirmation identity does not match", ErrInvalidConfirmation)
	}
	result, duplicate, err := e.mutations.store.ApplyStoryMutationConfirmation(
		ctx,
		storyMutationConfirmationBinding(claims, token),
		e.mutations.now().UTC(),
		func(applyCtx context.Context) (StoryMutationResult, error) {
			_, joinedByID, err := e.joinedTeams(applyCtx, scope)
			if err != nil {
				return StoryMutationResult{}, err
			}
			team, allowed := joinedByID[claims.TeamID]
			if !allowed {
				return StoryMutationResult{}, fmt.Errorf("%w: %s", ErrTeamNotAccessible, claims.TeamID)
			}

			switch claims.Operation {
			case StoryMutationCreate:
				return e.mutations.confirmCreate(applyCtx, scope, team, claims)
			case StoryMutationUpdate:
				return e.mutations.confirmUpdate(applyCtx, scope, team, claims)
			default:
				return StoryMutationResult{}, fmt.Errorf("%w: unsupported operation", ErrInvalidConfirmation)
			}
		},
	)
	if err != nil {
		return StoryMutationResult{}, err
	}
	if duplicate {
		result.Status = storyMutationStatusAlreadyApplied
	}
	return result, nil
}

// CancelStoryMutation atomically consumes a pending proposal without invoking
// its write callback. Only the workspace/user identity bound into the signed
// token can cancel it; a later Confirm therefore cannot mutate.
func (e *FortyOneToolExecutor) CancelStoryMutation(
	ctx context.Context,
	scope ToolScope,
	token string,
) (StoryMutationCancellationResult, error) {
	if e.mutations == nil {
		return StoryMutationCancellationResult{}, fmt.Errorf("%w: story mutations are disabled", ErrInvalidConfirmation)
	}
	if err := validateToolScope(&scope); err != nil {
		return StoryMutationCancellationResult{}, err
	}
	claims, err := e.mutations.verifyClaims(token)
	if err != nil {
		return StoryMutationCancellationResult{}, err
	}
	if claims.WorkspaceID != scope.WorkspaceID || claims.UserID != scope.UserID {
		return StoryMutationCancellationResult{}, fmt.Errorf("%w: confirmation identity does not match", ErrInvalidConfirmation)
	}
	return e.mutations.store.CancelStoryMutationConfirmation(
		ctx,
		storyMutationConfirmationBinding(claims, token),
		e.mutations.now().UTC(),
	)
}

func (m *storyMutationExecutor) confirmCreate(
	ctx context.Context,
	scope ToolScope,
	team teams.CoreTeam,
	claims storyMutationClaims,
) (StoryMutationResult, error) {
	if claims.Title == nil || claims.Priority == nil || claims.StoryID != nil || claims.ExpectedUpdatedAt != nil {
		return StoryMutationResult{}, fmt.Errorf("%w: malformed create proposal", ErrInvalidConfirmation)
	}
	statusID, err := m.stories.FindFirstStatusByCategory(ctx, team.ID, scope.WorkspaceID, "unstarted")
	if err != nil {
		return StoryMutationResult{}, fmt.Errorf("resolve default story status: %w", err)
	}
	if statusID == nil {
		return StoryMutationResult{}, errors.New("team has no unstarted story status")
	}

	var assigneeID *uuid.UUID
	if claims.AssigneeAction == assigneeActionMe {
		userID := scope.UserID
		assigneeID = &userID
	} else if claims.AssigneeAction != assigneeActionUnassigned {
		return StoryMutationResult{}, fmt.Errorf("%w: invalid create assignee action", ErrInvalidConfirmation)
	}
	creationKey := "messaging:create-story:" + claims.ConfirmationID.String()
	story, err := m.stories.CreateExternalUserAction(ctx, scope.UserID, stories.CoreNewStory{
		Title:       *claims.Title,
		Status:      statusID,
		Assignee:    assigneeID,
		Reporter:    &scope.UserID,
		Priority:    *claims.Priority,
		Team:        team.ID,
		CreationKey: &creationKey,
	}, scope.WorkspaceID)
	if err != nil {
		return StoryMutationResult{}, fmt.Errorf("create confirmed story: %w", err)
	}
	status := storyMutationStatusAlreadyApplied
	if story.CreatedNow {
		status = storyMutationStatusApplied
	}
	return storyMutationResult(status, StoryMutationCreate, story, team.Code), nil
}

func (m *storyMutationExecutor) confirmUpdate(
	ctx context.Context,
	scope ToolScope,
	team teams.CoreTeam,
	claims storyMutationClaims,
) (StoryMutationResult, error) {
	if claims.StoryID == nil || claims.ExpectedUpdatedAt == nil || claims.ConfirmationID == uuid.Nil {
		return StoryMutationResult{}, fmt.Errorf("%w: malformed update proposal", ErrInvalidConfirmation)
	}
	story, err := m.stories.Get(ctx, *claims.StoryID, scope.WorkspaceID)
	if err != nil {
		return StoryMutationResult{}, fmt.Errorf("load story for confirmed update: %w", err)
	}
	if story.Team != team.ID || story.Workspace != scope.WorkspaceID {
		return StoryMutationResult{}, fmt.Errorf("%w: story team does not match proposal", ErrInvalidConfirmation)
	}

	updates := desiredStoryUpdates(story, scope.UserID, claims.Title, claims.Priority, claims.AssigneeAction)
	if len(updates) == 0 {
		return storyMutationResult(storyMutationStatusAlreadyApplied, StoryMutationUpdate, story, team.Code), nil
	}
	if !story.UpdatedAt.Equal(claims.ExpectedUpdatedAt.UTC()) {
		return StoryMutationResult{}, fmt.Errorf("%w: %s", ErrStaleMutation, storyReference(team.Code, story.SequenceID))
	}
	if err := m.stories.UpdateExternalUserActionIfUnchanged(
		ctx,
		scope.UserID,
		story.ID,
		scope.WorkspaceID,
		claims.ExpectedUpdatedAt.UTC(),
		updates,
	); err != nil {
		if errors.Is(err, stories.ErrStoryChanged) {
			return StoryMutationResult{}, fmt.Errorf("%w: %s", ErrStaleMutation, storyReference(team.Code, story.SequenceID))
		}
		return StoryMutationResult{}, fmt.Errorf("update confirmed story: %w", err)
	}
	updated, err := m.stories.Get(ctx, story.ID, scope.WorkspaceID)
	if err != nil {
		return StoryMutationResult{}, fmt.Errorf("reload confirmed story: %w", err)
	}
	return storyMutationResult(storyMutationStatusApplied, StoryMutationUpdate, updated, team.Code), nil
}

func (m *storyMutationExecutor) signClaims(claims storyMutationClaims) (string, error) {
	var encoded bytes.Buffer
	encoder := json.NewEncoder(&encoded)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(claims); err != nil {
		return "", fmt.Errorf("encode story mutation confirmation: %w", err)
	}
	payload := bytes.TrimSuffix(encoded.Bytes(), []byte("\n"))
	signature := hmac.New(sha256.New, m.key)
	_, _ = signature.Write(payload)
	token := base64.RawURLEncoding.EncodeToString(payload) + "." + base64.RawURLEncoding.EncodeToString(signature.Sum(nil))
	if len(token) > maximumStoryMutationTokenSize {
		return "", fmt.Errorf("story mutation confirmation exceeds %d characters", maximumStoryMutationTokenSize)
	}
	return token, nil
}

func (m *storyMutationExecutor) verifyClaims(token string) (storyMutationClaims, error) {
	token = strings.TrimSpace(token)
	if token == "" || len(token) > maximumStoryMutationTokenSize {
		return storyMutationClaims{}, fmt.Errorf("%w: token is missing or too large", ErrInvalidConfirmation)
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return storyMutationClaims{}, fmt.Errorf("%w: token format", ErrInvalidConfirmation)
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || base64.RawURLEncoding.EncodeToString(payload) != parts[0] {
		return storyMutationClaims{}, fmt.Errorf("%w: token payload", ErrInvalidConfirmation)
	}
	providedSignature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || base64.RawURLEncoding.EncodeToString(providedSignature) != parts[1] {
		return storyMutationClaims{}, fmt.Errorf("%w: token signature", ErrInvalidConfirmation)
	}
	expectedSignature := hmac.New(sha256.New, m.key)
	_, _ = expectedSignature.Write(payload)
	if !hmac.Equal(providedSignature, expectedSignature.Sum(nil)) {
		return storyMutationClaims{}, fmt.Errorf("%w: token signature", ErrInvalidConfirmation)
	}

	var claims storyMutationClaims
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&claims); err != nil {
		return storyMutationClaims{}, fmt.Errorf("%w: token claims", ErrInvalidConfirmation)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return storyMutationClaims{}, fmt.Errorf("%w: trailing token claims", ErrInvalidConfirmation)
	}
	if claims.Version != storyMutationTokenVersion || claims.ConfirmationID == uuid.Nil || claims.WorkspaceID == uuid.Nil || claims.UserID == uuid.Nil || claims.TeamID == uuid.Nil || claims.ExpiresAt.IsZero() {
		return storyMutationClaims{}, fmt.Errorf("%w: incomplete token claims", ErrInvalidConfirmation)
	}
	if err := validateStoryMutationClaims(claims); err != nil {
		return storyMutationClaims{}, err
	}
	return claims, nil
}

func storyMutationConfirmationBinding(claims storyMutationClaims, token string) StoryMutationConfirmationBinding {
	return StoryMutationConfirmationBinding{
		ConfirmationID: claims.ConfirmationID,
		WorkspaceID:    claims.WorkspaceID,
		UserID:         claims.UserID,
		TokenHash:      storyMutationTokenHash(token),
	}
}

func storyMutationTokenHash(token string) []byte {
	digest := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return append([]byte(nil), digest[:]...)
}

func validateStoryMutationClaims(claims storyMutationClaims) error {
	if claims.Title != nil {
		title, err := normalizedStoryTitle(*claims.Title)
		if err != nil || title != *claims.Title {
			return fmt.Errorf("%w: invalid title claim", ErrInvalidConfirmation)
		}
	}
	if claims.Priority != nil {
		priority, err := normalizedStoryPriority(claims.Priority, "")
		if err != nil || priority != *claims.Priority {
			return fmt.Errorf("%w: invalid priority claim", ErrInvalidConfirmation)
		}
	}

	switch claims.Operation {
	case StoryMutationCreate:
		if claims.Title == nil || claims.Priority == nil || claims.StoryID != nil || claims.ExpectedUpdatedAt != nil {
			return fmt.Errorf("%w: malformed create claims", ErrInvalidConfirmation)
		}
		if claims.AssigneeAction != assigneeActionMe && claims.AssigneeAction != assigneeActionUnassigned {
			return fmt.Errorf("%w: invalid create assignee claim", ErrInvalidConfirmation)
		}
	case StoryMutationUpdate:
		if claims.StoryID == nil || *claims.StoryID == uuid.Nil || claims.ExpectedUpdatedAt == nil || claims.ExpectedUpdatedAt.IsZero() {
			return fmt.Errorf("%w: malformed update claims", ErrInvalidConfirmation)
		}
		if claims.AssigneeAction != assigneeActionUnchanged && claims.AssigneeAction != assigneeActionMe && claims.AssigneeAction != assigneeActionUnassigned {
			return fmt.Errorf("%w: invalid update assignee claim", ErrInvalidConfirmation)
		}
		if claims.Title == nil && claims.Priority == nil && claims.AssigneeAction == assigneeActionUnchanged {
			return fmt.Errorf("%w: empty update claims", ErrInvalidConfirmation)
		}
	default:
		return fmt.Errorf("%w: unsupported operation", ErrInvalidConfirmation)
	}
	return nil
}

func mutationConfirmationFromToolResult(raw json.RawMessage) (*StoryMutationConfirmation, bool, error) {
	var header struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(raw, &header); err != nil || header.Kind != storyMutationConfirmationKind {
		return nil, false, nil
	}
	var result storyMutationConfirmationToolResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, true, fmt.Errorf("%w: decode mutation confirmation: %v", ErrMalformedResponse, err)
	}
	confirmation := result.Confirmation
	if confirmation.Token == "" || confirmation.Prompt == "" || confirmation.ExpiresAt.IsZero() || confirmation.Story.TeamID == uuid.Nil {
		return nil, true, fmt.Errorf("%w: incomplete mutation confirmation", ErrMalformedResponse)
	}
	return &confirmation, true, nil
}

func validateToolScope(scope *ToolScope) error {
	if scope.WorkspaceID == uuid.Nil || scope.UserID == uuid.Nil {
		return fmt.Errorf("%w: workspace and user are required", ErrInvalidRequest)
	}
	allowedTeamIDs, err := normalizedAllowedTeamIDs(scope.AllowedTeamIDs)
	if err != nil {
		return err
	}
	scope.AllowedTeamIDs = allowedTeamIDs
	return nil
}

func parseAccessibleTeamID(raw string, joined map[uuid.UUID]teams.CoreTeam) (uuid.UUID, error) {
	teamID, err := parseRequiredUUID(raw, "team_id")
	if err != nil {
		return uuid.Nil, err
	}
	if _, ok := joined[teamID]; !ok {
		return uuid.Nil, fmt.Errorf("%w: %s", ErrTeamNotAccessible, teamID)
	}
	return teamID, nil
}

func parseRequiredUUID(raw, field string) (uuid.UUID, error) {
	value, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil || value == uuid.Nil {
		return uuid.Nil, fmt.Errorf("%w: %s must be a UUID", ErrInvalidToolArguments, field)
	}
	return value, nil
}

func normalizedStoryTitle(raw string) (string, error) {
	title := strings.TrimSpace(raw)
	if title == "" || len([]rune(title)) > maximumStoryTitleRunes {
		return "", fmt.Errorf("%w: title must contain 1-%d characters", ErrInvalidToolArguments, maximumStoryTitleRunes)
	}
	return title, nil
}

func normalizedStoryReference(raw string) (string, string, error) {
	reference := strings.ToUpper(strings.TrimSpace(raw))
	separator := strings.LastIndex(reference, "-")
	if separator < 1 || separator == len(reference)-1 {
		return "", "", fmt.Errorf("%w: story_reference must use TEAM-123 format", ErrInvalidToolArguments)
	}
	teamCode := strings.TrimSpace(reference[:separator])
	sequenceRaw := strings.TrimSpace(reference[separator+1:])
	sequenceID, err := strconv.Atoi(sequenceRaw)
	if teamCode == "" || err != nil || sequenceID < 1 {
		return "", "", fmt.Errorf("%w: story_reference must use TEAM-123 format", ErrInvalidToolArguments)
	}
	return fmt.Sprintf("%s-%d", teamCode, sequenceID), teamCode, nil
}

func normalizedStoryPriority(raw *string, fallback string) (string, error) {
	if raw == nil {
		return fallback, nil
	}
	priority := strings.TrimSpace(*raw)
	if _, ok := storyPriorities[priority]; !ok {
		return "", fmt.Errorf("%w: unsupported priority %q", ErrInvalidToolArguments, priority)
	}
	return priority, nil
}

func proposedChangedFields(story stories.CoreSingleStory, userID uuid.UUID, title, priority *string, assigneeAction string) []string {
	updates := desiredStoryUpdates(story, userID, title, priority, assigneeAction)
	fields := make([]string, 0, 3)
	for _, field := range []string{"title", "priority", "assignee_id"} {
		if _, changed := updates[field]; changed {
			fields = append(fields, field)
		}
	}
	return fields
}

func desiredStoryUpdates(story stories.CoreSingleStory, userID uuid.UUID, title, priority *string, assigneeAction string) map[string]any {
	updates := make(map[string]any, 3)
	if title != nil && story.Title != *title {
		updates["title"] = *title
	}
	if priority != nil && story.Priority != *priority {
		updates["priority"] = *priority
	}
	switch assigneeAction {
	case assigneeActionMe:
		if story.Assignee == nil || *story.Assignee != userID {
			updates["assignee_id"] = userID
		}
	case assigneeActionUnassigned:
		if story.Assignee != nil {
			updates["assignee_id"] = nil
		}
	}
	return updates
}

func storyMutationResult(status string, operation StoryMutationOperation, story stories.CoreSingleStory, teamCode string) StoryMutationResult {
	return StoryMutationResult{
		Status:     status,
		Operation:  operation,
		StoryID:    story.ID,
		Reference:  storyReference(strings.ToUpper(teamCode), story.SequenceID),
		TeamID:     story.Team,
		Title:      story.Title,
		Priority:   story.Priority,
		AssigneeID: story.Assignee,
	}
}

func storyMutationToolDefinitions() []ToolDefinition {
	nullablePriority := map[string]any{
		"type":        []string{"string", "null"},
		"description": "A FortyOne priority, or null as described by this tool.",
		"enum":        []any{"No Priority", "Low", "Medium", "High", "Urgent", nil},
	}
	return []ToolDefinition{
		{
			Type:        "function",
			Name:        toolCreateStory,
			Description: "Prepare a story creation proposal only when the user explicitly asks to create one and the team and title are unambiguous. This tool never writes; FortyOne will require explicit user confirmation.",
			Strict:      true,
			Parameters: strictObjectSchema(map[string]any{
				"team_id": map[string]any{
					"type":        "string",
					"description": "An exact team UUID returned by list_teams.",
				},
				"title": map[string]any{
					"type":        "string",
					"description": "The exact story title requested by the user.",
					"minLength":   1,
					"maxLength":   maximumStoryTitleRunes,
				},
				"priority": nullablePriority,
				"assignee": map[string]any{
					"type":        "string",
					"description": "Use me only if the user explicitly asks to assign it to themselves; otherwise use unassigned.",
					"enum":        []string{assigneeActionMe, assigneeActionUnassigned},
				},
			}, []string{"team_id", "title", "priority", "assignee"}),
		},
		{
			Type:        "function",
			Name:        toolUpdateStory,
			Description: "Prepare a story update proposal only when the target story and requested fields are unambiguous. This tool never writes; FortyOne will require explicit user confirmation.",
			Strict:      true,
			Parameters: strictObjectSchema(map[string]any{
				"story_id": map[string]any{
					"type":        []string{"string", "null"},
					"description": "An exact story UUID returned by a read tool, or null when story_reference is provided.",
				},
				"story_reference": map[string]any{
					"type":        []string{"string", "null"},
					"description": "An exact human reference such as WEB-123, or null when story_id is provided. Provide exactly one target.",
				},
				"title": map[string]any{
					"type":        []string{"string", "null"},
					"description": "The replacement title, or null to leave it unchanged.",
					"maxLength":   maximumStoryTitleRunes,
				},
				"priority": nullablePriority,
				"assignee": map[string]any{
					"type":        "string",
					"description": "Whether to leave the assignee unchanged, assign the current user, or unassign the story.",
					"enum":        []string{assigneeActionUnchanged, assigneeActionMe, assigneeActionUnassigned},
				},
			}, []string{"story_id", "story_reference", "title", "priority", "assignee"}),
		},
	}
}
