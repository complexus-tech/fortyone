package slack

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type workObjectStoryServiceStub struct {
	*mockStoryService
	mu            sync.Mutex
	story         stories.CoreSingleStory
	updates       map[string]any
	updateCalls   int
	queryCalls    int
	updateErr     error
	expectedActor uuid.UUID
	expectedAt    time.Time
}

func (s *workObjectStoryServiceStub) QueryByRef(_ context.Context, _ uuid.UUID, _ string) (stories.CoreSingleStory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queryCalls++
	return s.story, nil
}

func (s *workObjectStoryServiceStub) UpdateExternalUserActionIfUnchanged(
	_ context.Context,
	actorID, storyID, workspaceID uuid.UUID,
	expectedUpdatedAt time.Time,
	updates map[string]any,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updateCalls++
	s.expectedActor = actorID
	s.expectedAt = expectedUpdatedAt
	s.updates = make(map[string]any, len(updates))
	for field, value := range updates {
		s.updates[field] = value
	}
	if s.updateErr != nil {
		return s.updateErr
	}
	if storyID != s.story.ID || workspaceID != s.story.Workspace || !expectedUpdatedAt.Equal(s.story.UpdatedAt) {
		return stories.ErrStoryChanged
	}
	applyWorkObjectStoryUpdates(&s.story, updates)
	s.story.UpdatedAt = s.story.UpdatedAt.Add(time.Second)
	return nil
}

func applyWorkObjectStoryUpdates(story *stories.CoreSingleStory, updates map[string]any) {
	if value, ok := updates["title"].(string); ok {
		story.Title = value
	}
	if value, present := updates["description"]; present {
		if value == nil {
			story.Description = nil
		} else if description, ok := value.(string); ok {
			story.Description = &description
		}
	}
	if _, present := updates["description_html"]; present {
		story.DescriptionHTML = nil
	}
	if value, ok := updates["status_id"].(uuid.UUID); ok {
		story.Status = &value
	}
	if value, ok := updates["priority"].(string); ok {
		story.Priority = value
	}
	if value, present := updates["assignee_id"]; present {
		if value == nil {
			story.Assignee = nil
		} else if assigneeID, ok := value.(uuid.UUID); ok {
			story.Assignee = &assigneeID
		}
	}
	if value, present := updates["end_date"]; present {
		if value == nil {
			story.EndDate = nil
		} else if dueDate, ok := value.(time.Time); ok {
			story.EndDate = &dueDate
		}
	}
}

func TestProcessSlackWorkObjectEditReauthorizesUpdatesAndRefreshesFlexpane(t *testing.T) {
	t.Parallel()

	fixture := newWorkObjectEditFixture(t)
	newStatusID := fixture.repo.statuses[1].ID
	newAssigneeID := fixture.repo.teamMembers[1].UserID
	fixture.payload.View.State.Values = interactionViewStateValues{
		"title":       workObjectTextState("title", "A sharper title"),
		"description": workObjectTextState("description", "### Context\nFrom Slack"),
		"status":      workObjectSelectState("status", newStatusID.String()),
		"priority":    workObjectSelectState("priority", "Urgent"),
		"assignee":    workObjectSelectState("assignee", newAssigneeID.String()),
		"due_date":    workObjectDateState("due_date", "2026-08-31"),
	}

	var presented SlackEntityDetailsRequest
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		require.Equal(t, "/entity.presentDetails", request.URL.Path)
		require.Equal(t, "Bearer xoxb-test", request.Header.Get("Authorization"))
		require.NoError(t, json.NewDecoder(request.Body).Decode(&presented))
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer provider.Close()
	fixture.service.webClient.baseURL = provider.URL

	err := fixture.service.processSlackWorkObjectEdit(context.Background(), fixture.payload)
	require.NoError(t, err)
	require.Equal(t, 1, fixture.stories.updateCalls)
	require.Equal(t, fixture.actorID, fixture.stories.expectedActor)
	require.Equal(t, "A sharper title", fixture.stories.updates["title"])
	require.Equal(t, "### Context\nFrom Slack", fixture.stories.updates["description"])
	require.Contains(t, fixture.stories.updates, "description_html")
	require.Equal(t, newStatusID, fixture.stories.updates["status_id"])
	require.Equal(t, "Urgent", fixture.stories.updates["priority"])
	require.Equal(t, newAssigneeID, fixture.stories.updates["assignee_id"])
	require.Equal(t, "2026-08-31", fixture.stories.updates["end_date"].(time.Time).Format(time.DateOnly))

	require.Equal(t, fixture.payload.TriggerID, presented.TriggerID)
	require.NotNil(t, presented.Metadata)
	require.Equal(t, "A sharper title", presented.Metadata.EntityPayload.Attributes.Title.Text)
	require.True(t, presented.Metadata.EntityPayload.Attributes.Title.Edit.Enabled)
	require.Equal(t, newStatusID.String(), presented.Metadata.EntityPayload.Fields["status"].Edit.Select.CurrentValue)
	require.Len(t, presented.Metadata.EntityPayload.Fields["status"].Edit.Select.StaticOptions, 2)
	require.Equal(t, newAssigneeID.String(), presented.Metadata.EntityPayload.Fields["assignee"].Edit.Select.CurrentValue)
	require.Len(t, presented.Metadata.EntityPayload.Fields["assignee"].Edit.Select.StaticOptions, 2)
	require.Equal(t, "2026-08-31", presented.Metadata.EntityPayload.Fields["due_date"].Value)
}

func TestProcessSlackWorkObjectEditRejectsForgedTeamScopedIDs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state interactionViewStateValues
	}{
		{
			name:  "status from another team",
			state: interactionViewStateValues{"status": workObjectSelectState("status", uuid.NewString())},
		},
		{
			name:  "assignee from another team",
			state: interactionViewStateValues{"assignee": workObjectSelectState("assignee", uuid.NewString())},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			fixture := newWorkObjectEditFixture(t)
			fixture.payload.View.State.Values = test.state

			err := fixture.service.processSlackWorkObjectEdit(context.Background(), fixture.payload)
			require.ErrorIs(t, err, errSlackWorkObjectEditDenied)
			require.Zero(t, fixture.stories.updateCalls)
		})
	}
}

func TestProcessSlackWorkObjectEditRejectsExternalReferenceForgery(t *testing.T) {
	t.Parallel()

	fixture := newWorkObjectEditFixture(t)
	fixture.payload.View.ExternalRef.ID = "acme:" + uuid.NewString()
	fixture.payload.View.State.Values = interactionViewStateValues{"title": workObjectTextState("title", "Do not apply")}

	err := fixture.service.processSlackWorkObjectEdit(context.Background(), fixture.payload)
	require.ErrorIs(t, err, errSlackWorkObjectEditMalformed)
	require.Zero(t, fixture.stories.updateCalls)
}

func TestProcessSlackWorkObjectEditRejectsRevokedChannelAudience(t *testing.T) {
	t.Parallel()

	fixture := newWorkObjectEditFixture(t)
	fixture.repo.authorizedTeamIDs = []uuid.UUID{}
	fixture.payload.View.State.Values = interactionViewStateValues{"title": workObjectTextState("title", "Do not apply")}

	err := fixture.service.processSlackWorkObjectEdit(context.Background(), fixture.payload)
	require.ErrorIs(t, err, errSlackWorkObjectEditDenied)
	require.Zero(t, fixture.stories.updateCalls)
}

func TestProcessSlackWorkObjectEditReturnsStaleConflictWithoutRefresh(t *testing.T) {
	t.Parallel()

	fixture := newWorkObjectEditFixture(t)
	fixture.stories.updateErr = stories.ErrStoryChanged
	fixture.payload.View.State.Values = interactionViewStateValues{"title": workObjectTextState("title", "Conflicting title")}

	err := fixture.service.processSlackWorkObjectEdit(context.Background(), fixture.payload)
	require.ErrorIs(t, err, stories.ErrStoryChanged)
	require.Equal(t, 1, fixture.stories.updateCalls)
	require.Equal(t, "This task changed while you were editing it. Refresh the task and try again.", slackWorkObjectEditErrorMessage(err))
	require.Equal(t, 1, fixture.stories.queryCalls)
}

func TestProcessSlackWorkObjectEditRejectsMalformedSubmission(t *testing.T) {
	t.Parallel()

	fixture := newWorkObjectEditFixture(t)
	fixture.payload.View.State.Values = interactionViewStateValues{
		"due_date": workObjectDateState("due_date", "31/08/2026"),
	}

	err := fixture.service.processSlackWorkObjectEdit(context.Background(), fixture.payload)
	require.ErrorIs(t, err, errSlackWorkObjectEditMalformed)
	require.Zero(t, fixture.stories.updateCalls)
}

func TestAuthorizedSlackWorkObjectUpdatesPreservesExplicitClears(t *testing.T) {
	t.Parallel()

	fixture := newWorkObjectEditFixture(t)
	updates, err := fixture.service.authorizedSlackWorkObjectUpdates(
		context.Background(),
		fixture.stories.story,
		interactionViewStateValues{
			"description": workObjectTextState("description", ""),
			"assignee":    workObjectSelectState("assignee", ""),
			"due_date":    workObjectDateState("due_date", ""),
		},
	)
	require.NoError(t, err)
	require.Contains(t, updates, "description")
	require.Nil(t, updates["description"])
	require.Contains(t, updates, "description_html")
	require.Nil(t, updates["description_html"])
	require.Contains(t, updates, "assignee_id")
	require.Nil(t, updates["assignee_id"])
	require.Contains(t, updates, "end_date")
	require.Nil(t, updates["end_date"])
}

func TestSlackWorkObjectEditDetectionAcceptsSlackEmptyCallback(t *testing.T) {
	t.Parallel()

	payload := interactionPayload{Type: "view_submission"}
	payload.View.Type = slackWorkObjectEntityViewType
	require.True(t, isSlackWorkObjectEditSubmission(payload))
	payload.View.CallbackID = slackWorkObjectEditCallbackID
	require.True(t, isSlackWorkObjectEditSubmission(payload))
	payload.View.CallbackID = "fortyone_create_task"
	require.False(t, isSlackWorkObjectEditSubmission(payload))
}

type workObjectEditFixture struct {
	service *Service
	repo    *mockRepo
	stories *workObjectStoryServiceStub
	payload interactionPayload
	actorID uuid.UUID
}

func newWorkObjectEditFixture(t *testing.T) workObjectEditFixture {
	t.Helper()
	workspaceID := uuid.New()
	installationID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	assigneeID := uuid.New()
	storyID := uuid.New()
	statusID := uuid.New()
	updatedAt := time.Date(2026, time.August, 9, 8, 0, 0, 0, time.UTC)
	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
		team:      slackrepository.TeamRecord{ID: teamID, Code: "WEB", Name: "Web"},
		statuses: []slackrepository.StatusRecord{
			{ID: statusID, Name: "Backlog", Category: "unstarted"},
			{ID: uuid.New(), Name: "In progress", Category: "started"},
		},
		teamMembers: []slackrepository.TeamMemberRecord{
			{UserID: actorID, FullName: "Joseph Mukorivo", Email: "joseph@example.com"},
			{UserID: assigneeID, FullName: "Ari Engineer", Email: "ari@example.com"},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID:                installationID,
			WorkspaceID:       workspaceID,
			SlackTeamID:       "T123",
			BotAccessToken:    "xoxb-test",
			InstallGeneration: uuid.New(),
			IsActive:          true,
		},
		slackUserLinks:    map[string]uuid.UUID{"T123:U123": actorID},
		authorizedTeamIDs: []uuid.UUID{teamID},
	}
	storyService := &workObjectStoryServiceStub{
		mockStoryService: &mockStoryService{},
		story: stories.CoreSingleStory{
			ID:         storyID,
			SequenceID: 123,
			Title:      "Original title",
			Status:     &statusID,
			Assignee:   &actorID,
			Reporter:   &actorID,
			Priority:   "High",
			Team:       teamID,
			TeamCode:   "WEB",
			Workspace:  workspaceID,
			CreatedAt:  updatedAt.Add(-time.Hour),
			UpdatedAt:  updatedAt,
		},
	}
	testLogger := logger.NewWithJSON(io.Discard, slog.LevelError, "test")
	service := New(testLogger, repo, &mockRequestStore{}, storyService, Config{})

	payload := interactionPayload{Type: "view_submission", TriggerID: "trigger-123"}
	payload.Team.ID = "T123"
	payload.User.ID = "U123"
	payload.User.Username = "joseph"
	payload.View.Type = slackWorkObjectEntityViewType
	payload.View.CallbackID = ""
	payload.View.EntityURL = "https://acme.fortyone.app/work/WEB-123"
	payload.View.ExternalRef = SlackWorkObjectExternalRef{ID: "acme:" + storyID.String(), Type: slackStoryExternalRefType}
	payload.View.Channel = "C123"
	payload.View.MessageTS = "1754700000.123"

	return workObjectEditFixture{
		service: service,
		repo:    repo,
		stories: storyService,
		payload: payload,
		actorID: actorID,
	}
}

func workObjectTextState(field, value string) map[string]interactionViewStateValue {
	return map[string]interactionViewStateValue{
		field + ".input": {Type: "plain_text_input", Value: value},
	}
}

func workObjectSelectState(field, value string) map[string]interactionViewStateValue {
	state := interactionViewStateValue{Type: "static_select"}
	state.SelectedOption.Value = value
	return map[string]interactionViewStateValue{field + ".input": state}
}

func workObjectDateState(field, value string) map[string]interactionViewStateValue {
	return map[string]interactionViewStateValue{
		field + ".input": {Type: "datepicker", SelectedDate: value},
	}
}

func TestBuildSlackStoryEntityDetailsErrorRequest(t *testing.T) {
	t.Parallel()

	request, err := BuildSlackStoryEntityDetailsErrorRequest("trigger-123", "Try again")
	require.NoError(t, err)
	require.Nil(t, request.Metadata)
	require.Equal(t, "edit_error", request.Error.Status)
	require.Equal(t, "Try again", request.Error.CustomMessage)
}

func TestSlackWorkObjectPublisherPresentsEditError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		var payload SlackEntityDetailsRequest
		require.NoError(t, json.NewDecoder(request.Body).Decode(&payload))
		require.Equal(t, "trigger-123", payload.TriggerID)
		require.Equal(t, "edit_error", payload.Error.Status)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	client := newSlackWebClient(server.Client())
	client.baseURL = server.URL
	publisher := newSlackWorkObjectPublisher(client)
	request, err := BuildSlackStoryEntityDetailsErrorRequest("trigger-123", "Refresh and retry")
	require.NoError(t, err)
	require.NoError(t, publisher.PresentDetails(context.Background(), "xoxb-test", request))
}

func TestProcessSlackWorkObjectEditRejectsInactiveInstallation(t *testing.T) {
	t.Parallel()

	fixture := newWorkObjectEditFixture(t)
	fixture.repo.slackWorkspace.IsActive = false
	fixture.payload.View.State.Values = interactionViewStateValues{"title": workObjectTextState("title", "Do not apply")}

	err := fixture.service.processSlackWorkObjectEdit(context.Background(), fixture.payload)
	require.ErrorIs(t, err, errSlackWorkObjectEditDenied)
	require.Zero(t, fixture.stories.updateCalls)
}

func TestSlackWorkObjectEditErrorMessageDoesNotExposeProviderDetails(t *testing.T) {
	t.Parallel()

	message := slackWorkObjectEditErrorMessage(errors.New("database password leaked here"))
	require.NotContains(t, message, "password")
	require.Equal(t, "FortyOne could not save these changes. Refresh the task and try again.", message)
}
