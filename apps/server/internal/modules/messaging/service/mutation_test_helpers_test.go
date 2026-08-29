package messaging

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

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

func newBatchMutationToolExecutorForTest(
	t *testing.T,
	teamsService TeamsService,
	storiesService *mutationStoriesStub,
	usersService UsersService,
	secret string,
) *FortyOneToolExecutor {
	t.Helper()
	executor, err := NewFortyOneToolExecutor(
		teamsService,
		storiesService,
		&searchServiceStub{},
		&objectivesServiceStub{},
		WithOperationalTools(OperationalToolServices{
			States: &statesServiceStub{},
			Users:  usersService,
		}),
		WithStoryMutations(secret),
		WithStoryMutationConfirmationStore(storiesService.confirmations),
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
	failCreateAt             int
	failedCreate             bool
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
	input     StoryMutationConfirmationStateInput
	status    StoryMutationConfirmationStatus
	result    *StoryMutationResult
	lastError string
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
	input.Proposal = append(json.RawMessage(nil), input.Proposal...)
	s.entries[input.ConfirmationID] = &mutationConfirmationStoreEntry{
		input:  input,
		status: StoryMutationConfirmationPending,
	}
	return nil
}

func (s *mutationConfirmationStoreStub) LoadStoryMutationConfirmation(
	_ context.Context,
	binding StoryMutationConfirmationBinding,
) (StoryMutationConfirmationRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, err := s.boundEntry(binding)
	if err != nil {
		return StoryMutationConfirmationRecord{}, err
	}
	return StoryMutationConfirmationRecord{
		TeamID:    entry.input.TeamID,
		Operation: entry.input.Operation,
		Status:    entry.status,
		Proposal:  append(json.RawMessage(nil), entry.input.Proposal...),
		Result:    entry.result,
		LastError: entry.lastError,
	}, nil
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
	if (entry.status == StoryMutationConfirmationPending ||
		(entry.status == StoryMutationConfirmationApplied && len(entry.input.Proposal) != 0)) &&
		!now.Before(entry.input.ExpiresAt) {
		entry.status = StoryMutationConfirmationExpired
		entry.input.Proposal = nil
	}
	switch entry.status {
	case StoryMutationConfirmationCancelled:
		return StoryMutationResult{}, false, ErrCancelledConfirmation
	case StoryMutationConfirmationExpired:
		return StoryMutationResult{}, false, ErrExpiredConfirmation
	case StoryMutationConfirmationPending:
		entry.status = StoryMutationConfirmationApplied
	case StoryMutationConfirmationApplied:
		if entry.result != nil && entry.lastError == "" {
			return *entry.result, true, nil
		}
	}
	result, err := apply(ctx)
	if err != nil {
		if result.Operation == StoryMutationCreateBatch && (entry.result == nil || len(result.Items) >= len(entry.result.Items)) {
			stored := result
			entry.result = &stored
		}
		entry.lastError = err.Error()
		if entry.result != nil {
			return *entry.result, false, err
		}
		return result, false, err
	}
	stored := result
	entry.result = &stored
	entry.input.Proposal = nil
	entry.lastError = ""
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
		entry.input.Proposal = nil
	}
	switch entry.status {
	case StoryMutationConfirmationPending:
		entry.status = StoryMutationConfirmationCancelled
		entry.input.Proposal = nil
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

func (s *mutationStoriesStub) CreateCommentExternal(_ context.Context, _ uuid.UUID, _ uuid.UUID, comment stories.CoreNewComment) (stories.CoreComment, error) {
	return stories.CoreComment{ID: uuid.New(), StoryID: comment.StoryID, UserID: comment.UserID, Comment: comment.Comment}, nil
}

func (s *mutationStoriesStub) UpdateLabels(_ context.Context, storyID, workspaceID uuid.UUID, labels []uuid.UUID) error {
	story, ok := s.persisted[storyID]
	if !ok || story.Workspace != workspaceID {
		return errors.New("story not found")
	}
	story.Labels = append([]uuid.UUID(nil), labels...)
	s.persisted[storyID] = story
	return nil
}

func (s *mutationStoriesStub) AddAssociation(_ context.Context, fromID, toID uuid.UUID, associationType string, workspaceID uuid.UUID) (stories.CoreStoryAssociation, error) {
	from, fromOK := s.persisted[fromID]
	to, toOK := s.persisted[toID]
	if !fromOK || !toOK || from.Workspace != workspaceID || to.Workspace != workspaceID {
		return stories.CoreStoryAssociation{}, errors.New("story not found")
	}
	return stories.CoreStoryAssociation{ID: uuid.New(), FromStoryID: fromID, ToStoryID: toID, Type: associationType}, nil
}

func (s *mutationStoriesStub) CreateExternalUserAction(ctx context.Context, actorID uuid.UUID, story stories.CoreNewStory, workspaceID uuid.UUID) (stories.CoreSingleStory, error) {
	contextActorID, _ := platformauth.GetUserID(ctx)
	s.createCalls = append(s.createCalls, mutationCreateCall{
		actorID:        actorID,
		contextActorID: contextActorID,
		story:          story,
		workspaceID:    workspaceID,
	})
	if s.failCreateAt > 0 && len(s.createCalls) == s.failCreateAt && !s.failedCreate {
		s.failedCreate = true
		return stories.CoreSingleStory{}, errors.New("transient create failure")
	}
	if story.CreationKey != nil {
		if existingID, ok := s.byCreationKey[*story.CreationKey]; ok {
			existing := s.persisted[existingID]
			existing.CreatedNow = false
			return existing, nil
		}
	}
	id := uuid.New()
	sequenceID := len(s.persisted) + 1
	created := stories.CoreSingleStory{
		ID:                       id,
		SequenceID:               sequenceID,
		Title:                    story.Title,
		Status:                   story.Status,
		Assignee:                 story.Assignee,
		Reporter:                 story.Reporter,
		Priority:                 story.Priority,
		Team:                     story.Team,
		Workspace:                workspaceID,
		EstimatedDurationMinutes: story.EstimatedDurationMinutes,
		MinimumFocusBlockMinutes: story.MinimumFocusBlockMinutes,
		UpdatedAt:                time.Date(2026, time.August, 9, 9, 0, 0, 0, time.UTC),
		CreatedNow:               true,
		CreationKey:              story.CreationKey,
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
	if value, exists := updates["estimated_duration_minutes"]; exists {
		switch typed := value.(type) {
		case nil:
			story.EstimatedDurationMinutes = nil
		case int:
			story.EstimatedDurationMinutes = cloneIntPointer(&typed)
		case *int:
			story.EstimatedDurationMinutes = cloneIntPointer(typed)
		}
	}
	if value, exists := updates["minimum_focus_block_minutes"]; exists {
		switch typed := value.(type) {
		case nil:
			story.MinimumFocusBlockMinutes = nil
		case int:
			story.MinimumFocusBlockMinutes = cloneIntPointer(&typed)
		case *int:
			story.MinimumFocusBlockMinutes = cloneIntPointer(typed)
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
