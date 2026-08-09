package messaging

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const testStoryMutationSecret = "a stable worker secret used only by messaging tests"

func TestFortyOneToolExecutorMutationCatalogIsStrictAndOptIn(t *testing.T) {
	t.Parallel()

	readOnly := newToolExecutorForTest(t, &teamsServiceStub{}, &storiesServiceStub{}, &searchServiceStub{}, &objectivesServiceStub{})
	require.Len(t, readOnly.Definitions(), 4)

	mutationStories := newMutationStoriesStub()
	executor := newMutationToolExecutorForTest(t, &teamsServiceStub{}, mutationStories, testStoryMutationSecret)
	definitions := executor.Definitions()
	require.Len(t, definitions, 6)
	require.NoError(t, validateToolDefinitions(definitions))
	require.Equal(t, toolCreateStory, definitions[4].Name)
	require.Equal(t, toolUpdateStory, definitions[5].Name)
}

func TestStoryMutationsRequireDurableConfirmationStore(t *testing.T) {
	t.Parallel()

	_, err := NewFortyOneToolExecutor(
		&teamsServiceStub{},
		newMutationStoriesStub(),
		&searchServiceStub{},
		&objectivesServiceStub{},
		WithStoryMutations(testStoryMutationSecret),
	)

	require.ErrorContains(t, err, "confirmation store is required")
}

func TestStoryCreateRequiresExplicitConfirmationAndIsRetrySafe(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	scope.AllowMutations = true
	team := mutationTestTeam(scope.WorkspaceID)
	scope.AllowedTeamIDs = []uuid.UUID{team.ID}
	storyService := newMutationStoriesStub()
	teamsService := &teamsServiceStub{joined: []teams.CoreTeam{team}}
	executor := newMutationToolExecutorForTest(t, teamsService, storyService, testStoryMutationSecret)

	arguments, err := json.Marshal(map[string]any{
		"team_id":  team.ID.String(),
		"title":    "  Fix checkout retries  ",
		"priority": "High",
		"assignee": assigneeActionMe,
	})
	require.NoError(t, err)
	output, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolCreateStory, Arguments: arguments})
	require.NoError(t, err)
	require.Empty(t, storyService.createCalls, "a model tool call must never write")

	confirmation, proposed, err := mutationConfirmationFromToolResult(output)
	require.NoError(t, err)
	require.True(t, proposed)
	require.Equal(t, StoryMutationCreate, confirmation.Operation)
	require.Equal(t, "Fix checkout retries", confirmation.Story.Title)
	require.Equal(t, "WEB", confirmation.Story.TeamCode)
	require.NotEmpty(t, confirmation.Token)

	deniedScope := scope
	deniedScope.AllowedTeamIDs = []uuid.UUID{}
	_, err = executor.ConfirmStoryMutation(context.Background(), deniedScope, confirmation.Token)
	require.ErrorIs(t, err, ErrTeamNotAccessible)
	require.Empty(t, storyService.createCalls)
	disabledScope := scope
	disabledScope.AllowMutations = false
	_, err = executor.ConfirmStoryMutation(context.Background(), disabledScope, confirmation.Token)
	require.ErrorIs(t, err, ErrMutationNotAllowed)
	require.Empty(t, storyService.createCalls)

	// Confirm on another executor to prove the signed token requires no process-
	// local state and works across worker replicas.
	secondExecutor := newMutationToolExecutorForTest(t, teamsService, storyService, testStoryMutationSecret)
	result, err := secondExecutor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
	require.NoError(t, err)
	require.Equal(t, storyMutationStatusApplied, result.Status)
	require.Equal(t, StoryMutationCreate, result.Operation)
	require.Equal(t, "WEB-1", result.Reference)
	require.Equal(t, scope.UserID, *result.AssigneeID)
	require.Len(t, storyService.persisted, 1)
	require.Len(t, storyService.createCalls, 1)
	created := storyService.createCalls[0]
	require.Equal(t, "Fix checkout retries", created.story.Title)
	require.Equal(t, storyService.defaultStatusID, created.story.Status)
	require.NotNil(t, created.story.CreationKey)
	require.True(t, strings.HasPrefix(*created.story.CreationKey, "messaging:create-story:"))
	require.Equal(t, scope.UserID, created.actorID)
	require.Equal(t, scope.UserID, created.contextActorID)

	retry, err := executor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
	require.NoError(t, err)
	require.Equal(t, storyMutationStatusAlreadyApplied, retry.Status)
	require.Equal(t, result.StoryID, retry.StoryID)
	require.Len(t, storyService.persisted, 1, "a retried confirmation must not duplicate the story")

	tampered := confirmation.Token[:len(confirmation.Token)-1] + "x"
	_, err = executor.ConfirmStoryMutation(context.Background(), scope, tampered)
	require.ErrorIs(t, err, ErrInvalidConfirmation)

	wrongUser := scope
	wrongUser.UserID = uuid.New()
	_, err = executor.ConfirmStoryMutation(context.Background(), wrongUser, confirmation.Token)
	require.ErrorIs(t, err, ErrInvalidConfirmation)
}

func TestStoryMutationCancellationIsOwnerBoundTerminalAndIdempotent(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	scope.AllowMutations = true
	team := mutationTestTeam(scope.WorkspaceID)
	scope.AllowedTeamIDs = []uuid.UUID{team.ID}
	storyService := newMutationStoriesStub()
	executor := newMutationToolExecutorForTest(
		t,
		&teamsServiceStub{joined: []teams.CoreTeam{team}},
		storyService,
		testStoryMutationSecret,
	)
	confirmation := proposeCreateMutationForTest(t, executor, scope, team.ID, "Cancel me")

	wrongActor := scope
	wrongActor.UserID = uuid.New()
	_, err := executor.CancelStoryMutation(context.Background(), wrongActor, confirmation.Token)
	require.ErrorIs(t, err, ErrInvalidConfirmation)

	cancelled, err := executor.CancelStoryMutation(context.Background(), scope, confirmation.Token)
	require.NoError(t, err)
	require.Equal(t, "cancelled", cancelled.Status)
	retry, err := executor.CancelStoryMutation(context.Background(), scope, confirmation.Token)
	require.NoError(t, err)
	require.Equal(t, "already_cancelled", retry.Status)

	_, err = executor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
	require.ErrorIs(t, err, ErrCancelledConfirmation)
	require.Empty(t, storyService.createCalls, "Cancel followed by Confirm must never write")
}

func TestStoryMutationConfirmationCannotLaterClaimCancellation(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	scope.AllowMutations = true
	team := mutationTestTeam(scope.WorkspaceID)
	scope.AllowedTeamIDs = []uuid.UUID{team.ID}
	storyService := newMutationStoriesStub()
	executor := newMutationToolExecutorForTest(
		t,
		&teamsServiceStub{joined: []teams.CoreTeam{team}},
		storyService,
		testStoryMutationSecret,
	)
	confirmation := proposeCreateMutationForTest(t, executor, scope, team.ID, "Confirm me")

	result, err := executor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
	require.NoError(t, err)
	require.Equal(t, storyMutationStatusApplied, result.Status)

	_, err = executor.CancelStoryMutation(context.Background(), scope, confirmation.Token)
	require.ErrorIs(t, err, ErrAppliedConfirmation)
	retry, err := executor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
	require.NoError(t, err)
	require.Equal(t, storyMutationStatusAlreadyApplied, retry.Status)
	require.Len(t, storyService.createCalls, 1)
}

func TestStoryMutationConfirmCancelRaceHasExactlyOneTerminalOutcome(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	scope.AllowMutations = true
	team := mutationTestTeam(scope.WorkspaceID)
	scope.AllowedTeamIDs = []uuid.UUID{team.ID}
	storyService := newMutationStoriesStub()
	executor := newMutationToolExecutorForTest(
		t,
		&teamsServiceStub{joined: []teams.CoreTeam{team}},
		storyService,
		testStoryMutationSecret,
	)
	confirmation := proposeCreateMutationForTest(t, executor, scope, team.ID, "Race me")

	start := make(chan struct{})
	confirmErr := make(chan error, 1)
	cancelErr := make(chan error, 1)
	go func() {
		<-start
		_, err := executor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
		confirmErr <- err
	}()
	go func() {
		<-start
		_, err := executor.CancelStoryMutation(context.Background(), scope, confirmation.Token)
		cancelErr <- err
	}()
	close(start)

	confirmed := <-confirmErr
	cancelled := <-cancelErr
	switch {
	case confirmed == nil:
		require.ErrorIs(t, cancelled, ErrAppliedConfirmation)
		require.Len(t, storyService.createCalls, 1)
	case cancelled == nil:
		require.ErrorIs(t, confirmed, ErrCancelledConfirmation)
		require.Empty(t, storyService.createCalls)
	default:
		t.Fatalf("neither terminal outcome won: confirm=%v cancel=%v", confirmed, cancelled)
	}
}

func TestStoryUpdateRequiresConfirmationRejectsStaleWritesAndIsRetrySafe(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	scope.AllowMutations = true
	team := mutationTestTeam(scope.WorkspaceID)
	scope.AllowedTeamIDs = []uuid.UUID{team.ID}
	storyService := newMutationStoriesStub()
	storyID := uuid.MustParse("baaaaaaa-0000-0000-0000-000000000001")
	storyService.persisted[storyID] = stories.CoreSingleStory{
		ID:         storyID,
		SequenceID: 42,
		Title:      "Old title",
		Priority:   "Low",
		Team:       team.ID,
		Workspace:  scope.WorkspaceID,
		UpdatedAt:  time.Date(2026, time.August, 9, 7, 0, 0, 123000000, time.UTC),
	}
	executor := newMutationToolExecutorForTest(t, &teamsServiceStub{joined: []teams.CoreTeam{team}}, storyService, testStoryMutationSecret)

	arguments, err := json.Marshal(map[string]any{
		"story_id":        nil,
		"story_reference": "web-42",
		"title":           "New title",
		"priority":        "High",
		"assignee":        assigneeActionMe,
	})
	require.NoError(t, err)
	output, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolUpdateStory, Arguments: arguments})
	require.NoError(t, err)
	require.Empty(t, storyService.updateCalls, "a model tool call must never write")
	confirmation, proposed, err := mutationConfirmationFromToolResult(output)
	require.NoError(t, err)
	require.True(t, proposed)
	require.Equal(t, []string{"title", "priority", "assignee_id"}, confirmation.Story.ChangedFields)

	result, err := executor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
	require.NoError(t, err)
	require.Equal(t, storyMutationStatusApplied, result.Status)
	require.Equal(t, "WEB-42", result.Reference)
	require.Equal(t, "New title", result.Title)
	require.Equal(t, "High", result.Priority)
	require.Equal(t, scope.UserID, *result.AssigneeID)
	require.Len(t, storyService.updateCalls, 1)

	retry, err := executor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
	require.NoError(t, err)
	require.Equal(t, storyMutationStatusAlreadyApplied, retry.Status)
	require.Len(t, storyService.updateCalls, 1, "a retried update must be a no-op")

	staleArguments, err := json.Marshal(map[string]any{
		"story_id":        storyID.String(),
		"story_reference": nil,
		"title":           nil,
		"priority":        "Urgent",
		"assignee":        assigneeActionUnchanged,
	})
	require.NoError(t, err)
	staleOutput, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolUpdateStory, Arguments: staleArguments})
	require.NoError(t, err)
	staleConfirmation, proposed, err := mutationConfirmationFromToolResult(staleOutput)
	require.NoError(t, err)
	require.True(t, proposed)
	external := storyService.persisted[storyID]
	external.Title = "Changed elsewhere"
	external.UpdatedAt = external.UpdatedAt.Add(time.Second)
	storyService.persisted[storyID] = external

	_, err = executor.ConfirmStoryMutation(context.Background(), scope, staleConfirmation.Token)
	require.ErrorIs(t, err, ErrStaleMutation)
	require.Len(t, storyService.updateCalls, 1, "stale confirmation must not write")

	raceArguments, err := json.Marshal(map[string]any{
		"story_id":        storyID.String(),
		"story_reference": nil,
		"title":           nil,
		"priority":        "Medium",
		"assignee":        assigneeActionUnchanged,
	})
	require.NoError(t, err)
	raceOutput, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolUpdateStory, Arguments: raceArguments})
	require.NoError(t, err)
	raceConfirmation, proposed, err := mutationConfirmationFromToolResult(raceOutput)
	require.NoError(t, err)
	require.True(t, proposed)
	storyService.forceConditionalConflict = true
	_, err = executor.ConfirmStoryMutation(context.Background(), scope, raceConfirmation.Token)
	require.ErrorIs(t, err, ErrStaleMutation)
	require.Len(t, storyService.updateCalls, 1, "a write-time version conflict must not update")
}

func TestStoryMutationProposalHonorsChannelScopeAndExpires(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	scope.AllowMutations = true
	team := mutationTestTeam(scope.WorkspaceID)
	storyService := newMutationStoriesStub()
	executor := newMutationToolExecutorForTest(t, &teamsServiceStub{joined: []teams.CoreTeam{team}}, storyService, testStoryMutationSecret)
	fixedNow := time.Date(2026, time.August, 9, 8, 0, 0, 0, time.UTC)
	executor.mutations.now = func() time.Time { return fixedNow }

	arguments, err := json.Marshal(map[string]any{
		"team_id":  team.ID.String(),
		"title":    "Scoped story",
		"priority": nil,
		"assignee": assigneeActionUnassigned,
	})
	require.NoError(t, err)
	deniedScope := scope
	deniedScope.AllowedTeamIDs = []uuid.UUID{}
	_, err = executor.Execute(context.Background(), deniedScope, ToolCall{Name: toolCreateStory, Arguments: arguments})
	require.ErrorIs(t, err, ErrTeamNotAccessible)

	scope.AllowedTeamIDs = []uuid.UUID{team.ID}
	output, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolCreateStory, Arguments: arguments})
	require.NoError(t, err)
	confirmation, proposed, err := mutationConfirmationFromToolResult(output)
	require.NoError(t, err)
	require.True(t, proposed)
	require.Equal(t, fixedNow.Add(storyMutationConfirmationTTL), confirmation.ExpiresAt)

	executor.mutations.now = func() time.Time { return fixedNow.Add(storyMutationConfirmationTTL) }
	_, err = executor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
	require.ErrorIs(t, err, ErrExpiredConfirmation)
	require.Empty(t, storyService.createCalls)

	// Expiration is durable. Even if a skewed replica reports an earlier clock,
	// the proposal can no longer transition to either terminal user outcome.
	executor.mutations.now = func() time.Time { return fixedNow }
	_, err = executor.CancelStoryMutation(context.Background(), scope, confirmation.Token)
	require.ErrorIs(t, err, ErrExpiredConfirmation)
	_, err = executor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
	require.ErrorIs(t, err, ErrExpiredConfirmation)
	require.Empty(t, storyService.createCalls)
}

func TestStoryMutationTokenFitsSlackActionValueAtMaximumTitleLength(t *testing.T) {
	t.Parallel()

	executor := newMutationStoriesStub()
	mutations := newStoryMutationExecutor(executor, testStoryMutationSecret, executor.confirmations)
	title := strings.Repeat("😀", maximumStoryTitleRunes)
	priority := "Urgent"
	storyID := uuid.New()
	expectedUpdatedAt := time.Now().UTC()
	token, err := mutations.signClaims(storyMutationClaims{
		Version:           storyMutationTokenVersion,
		ConfirmationID:    uuid.New(),
		Operation:         StoryMutationUpdate,
		WorkspaceID:       uuid.New(),
		UserID:            uuid.New(),
		TeamID:            uuid.New(),
		StoryID:           &storyID,
		ExpectedUpdatedAt: &expectedUpdatedAt,
		Title:             &title,
		Priority:          &priority,
		AssigneeAction:    assigneeActionUnassigned,
		ExpiresAt:         time.Now().UTC().Add(storyMutationConfirmationTTL),
	})

	require.NoError(t, err)
	require.LessOrEqual(t, len(token), maximumStoryMutationTokenSize)
}

func newMutationToolExecutorForTest(
	t *testing.T,
	teamsService TeamsService,
	storiesService StoryMutationService,
	secret string,
) *FortyOneToolExecutor {
	t.Helper()
	mutationStories, ok := storiesService.(*mutationStoriesStub)
	require.True(t, ok, "mutation tests require the durable confirmation store stub")
	executor, err := NewFortyOneToolExecutor(
		teamsService,
		storiesService,
		&searchServiceStub{},
		&objectivesServiceStub{},
		WithStoryMutations(secret),
		WithStoryMutationConfirmationStore(mutationStories.confirmations),
	)
	require.NoError(t, err)
	return executor
}

func mutationTestTeam(workspaceID uuid.UUID) teams.CoreTeam {
	return teams.CoreTeam{
		ID:        uuid.MustParse("caaaaaaa-0000-0000-0000-000000000001"),
		Name:      "Web",
		Code:      "web",
		Workspace: workspaceID,
	}
}

func proposeCreateMutationForTest(
	t *testing.T,
	executor *FortyOneToolExecutor,
	scope ToolScope,
	teamID uuid.UUID,
	title string,
) *StoryMutationConfirmation {
	t.Helper()
	arguments, err := json.Marshal(map[string]any{
		"team_id":  teamID.String(),
		"title":    title,
		"priority": nil,
		"assignee": assigneeActionUnassigned,
	})
	require.NoError(t, err)
	output, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolCreateStory, Arguments: arguments})
	require.NoError(t, err)
	confirmation, proposed, err := mutationConfirmationFromToolResult(output)
	require.NoError(t, err)
	require.True(t, proposed)
	return confirmation
}

type mutationCreateCall struct {
	actorID        uuid.UUID
	contextActorID uuid.UUID
	story          stories.CoreNewStory
	workspaceID    uuid.UUID
}

type mutationUpdateCall struct {
	actorID     uuid.UUID
	storyID     uuid.UUID
	workspaceID uuid.UUID
	updates     map[string]any
}

type mutationStoriesStub struct {
	persisted                map[uuid.UUID]stories.CoreSingleStory
	byCreationKey            map[string]uuid.UUID
	defaultStatusID          *uuid.UUID
	forceConditionalConflict bool
	createCalls              []mutationCreateCall
	updateCalls              []mutationUpdateCall
	confirmations            *mutationConfirmationStoreStub
}

func newMutationStoriesStub() *mutationStoriesStub {
	statusID := uuid.MustParse("daaaaaaa-0000-0000-0000-000000000001")
	return &mutationStoriesStub{
		persisted:       make(map[uuid.UUID]stories.CoreSingleStory),
		byCreationKey:   make(map[string]uuid.UUID),
		defaultStatusID: &statusID,
		confirmations:   newMutationConfirmationStoreStub(),
	}
}

type mutationConfirmationStoreEntry struct {
	input  StoryMutationConfirmationStateInput
	status StoryMutationConfirmationStatus
	result *StoryMutationResult
}

type mutationConfirmationStoreStub struct {
	mu      sync.Mutex
	entries map[uuid.UUID]*mutationConfirmationStoreEntry
}

func newMutationConfirmationStoreStub() *mutationConfirmationStoreStub {
	return &mutationConfirmationStoreStub{entries: make(map[uuid.UUID]*mutationConfirmationStoreEntry)}
}

func (s *mutationConfirmationStoreStub) RegisterStoryMutationConfirmation(
	_ context.Context,
	input StoryMutationConfirmationStateInput,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.entries[input.ConfirmationID]; exists {
		return errors.New("duplicate confirmation")
	}
	input.TokenHash = append([]byte(nil), input.TokenHash...)
	s.entries[input.ConfirmationID] = &mutationConfirmationStoreEntry{
		input:  input,
		status: StoryMutationConfirmationPending,
	}
	return nil
}

func (s *mutationConfirmationStoreStub) ApplyStoryMutationConfirmation(
	ctx context.Context,
	binding StoryMutationConfirmationBinding,
	now time.Time,
	apply func(context.Context) (StoryMutationResult, error),
) (StoryMutationResult, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, err := s.boundEntry(binding)
	if err != nil {
		return StoryMutationResult{}, false, err
	}
	if entry.status == StoryMutationConfirmationPending && !now.Before(entry.input.ExpiresAt) {
		entry.status = StoryMutationConfirmationExpired
	}
	switch entry.status {
	case StoryMutationConfirmationCancelled:
		return StoryMutationResult{}, false, ErrCancelledConfirmation
	case StoryMutationConfirmationExpired:
		return StoryMutationResult{}, false, ErrExpiredConfirmation
	case StoryMutationConfirmationPending:
		entry.status = StoryMutationConfirmationApplied
	case StoryMutationConfirmationApplied:
		if entry.result != nil {
			return *entry.result, true, nil
		}
	}
	result, err := apply(ctx)
	if err != nil {
		return StoryMutationResult{}, false, err
	}
	entry.result = &result
	return result, false, nil
}

func (s *mutationConfirmationStoreStub) CancelStoryMutationConfirmation(
	_ context.Context,
	binding StoryMutationConfirmationBinding,
	now time.Time,
) (StoryMutationCancellationResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, err := s.boundEntry(binding)
	if err != nil {
		return StoryMutationCancellationResult{}, err
	}
	if entry.status == StoryMutationConfirmationPending && !now.Before(entry.input.ExpiresAt) {
		entry.status = StoryMutationConfirmationExpired
	}
	switch entry.status {
	case StoryMutationConfirmationPending:
		entry.status = StoryMutationConfirmationCancelled
		return StoryMutationCancellationResult{Status: "cancelled"}, nil
	case StoryMutationConfirmationCancelled:
		return StoryMutationCancellationResult{Status: "already_cancelled"}, nil
	case StoryMutationConfirmationApplied:
		return StoryMutationCancellationResult{}, ErrAppliedConfirmation
	case StoryMutationConfirmationExpired:
		return StoryMutationCancellationResult{}, ErrExpiredConfirmation
	default:
		return StoryMutationCancellationResult{}, ErrInvalidConfirmation
	}
}

func (s *mutationConfirmationStoreStub) boundEntry(
	binding StoryMutationConfirmationBinding,
) (*mutationConfirmationStoreEntry, error) {
	entry := s.entries[binding.ConfirmationID]
	if entry == nil || entry.input.WorkspaceID != binding.WorkspaceID || entry.input.UserID != binding.UserID || !bytes.Equal(entry.input.TokenHash, binding.TokenHash) {
		return nil, ErrInvalidConfirmation
	}
	return entry, nil
}

func (s *mutationStoriesStub) MyStories(context.Context, uuid.UUID) ([]stories.CoreStoryList, error) {
	return nil, nil
}

func (s *mutationStoriesStub) Get(_ context.Context, id, workspaceID uuid.UUID) (stories.CoreSingleStory, error) {
	story, ok := s.persisted[id]
	if !ok || story.Workspace != workspaceID {
		return stories.CoreSingleStory{}, errors.New("story not found")
	}
	return story, nil
}

func (s *mutationStoriesStub) QueryByRef(_ context.Context, workspaceID uuid.UUID, storyRef string) (stories.CoreSingleStory, error) {
	for _, story := range s.persisted {
		if story.Workspace == workspaceID && storyRef == "WEB-"+strconv.Itoa(story.SequenceID) {
			return story, nil
		}
	}
	return stories.CoreSingleStory{}, errors.New("story not found")
}

func (s *mutationStoriesStub) CreateExternalUserAction(ctx context.Context, actorID uuid.UUID, story stories.CoreNewStory, workspaceID uuid.UUID) (stories.CoreSingleStory, error) {
	contextActorID, _ := platformauth.GetUserID(ctx)
	s.createCalls = append(s.createCalls, mutationCreateCall{
		actorID:        actorID,
		contextActorID: contextActorID,
		story:          story,
		workspaceID:    workspaceID,
	})
	if story.CreationKey != nil {
		if existingID, ok := s.byCreationKey[*story.CreationKey]; ok {
			existing := s.persisted[existingID]
			existing.CreatedNow = false
			return existing, nil
		}
	}
	id := uuid.MustParse("eaaaaaaa-0000-0000-0000-000000000001")
	created := stories.CoreSingleStory{
		ID:          id,
		SequenceID:  1,
		Title:       story.Title,
		Status:      story.Status,
		Assignee:    story.Assignee,
		Reporter:    story.Reporter,
		Priority:    story.Priority,
		Team:        story.Team,
		Workspace:   workspaceID,
		UpdatedAt:   time.Date(2026, time.August, 9, 9, 0, 0, 0, time.UTC),
		CreatedNow:  true,
		CreationKey: story.CreationKey,
	}
	s.persisted[id] = created
	if story.CreationKey != nil {
		s.byCreationKey[*story.CreationKey] = id
	}
	return created, nil
}

func (s *mutationStoriesStub) UpdateExternalUserActionIfUnchanged(
	_ context.Context,
	actorID, storyID, workspaceID uuid.UUID,
	expectedUpdatedAt time.Time,
	updates map[string]any,
) error {
	story, ok := s.persisted[storyID]
	if !ok || story.Workspace != workspaceID {
		return errors.New("story not found")
	}
	if !story.UpdatedAt.Equal(expectedUpdatedAt) {
		return stories.ErrStoryChanged
	}
	if s.forceConditionalConflict {
		return stories.ErrStoryChanged
	}
	copyUpdates := make(map[string]any, len(updates))
	for key, value := range updates {
		copyUpdates[key] = value
	}
	s.updateCalls = append(s.updateCalls, mutationUpdateCall{
		actorID:     actorID,
		storyID:     storyID,
		workspaceID: workspaceID,
		updates:     copyUpdates,
	})
	if value, ok := updates["title"].(string); ok {
		story.Title = value
	}
	if value, ok := updates["priority"].(string); ok {
		story.Priority = value
	}
	if value, exists := updates["assignee_id"]; exists {
		switch typed := value.(type) {
		case nil:
			story.Assignee = nil
		case uuid.UUID:
			valueCopy := typed
			story.Assignee = &valueCopy
		case *uuid.UUID:
			story.Assignee = typed
		}
	}
	story.UpdatedAt = story.UpdatedAt.Add(time.Second)
	s.persisted[storyID] = story
	return nil
}

func (s *mutationStoriesStub) FindFirstStatusByCategory(_ context.Context, _, _ uuid.UUID, category string) (*uuid.UUID, error) {
	if category != "unstarted" {
		return nil, errors.New("unexpected category")
	}
	return s.defaultStatusID, nil
}
