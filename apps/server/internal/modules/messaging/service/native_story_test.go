package messaging

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func nativeFixture(t *testing.T) (*FortyOneToolExecutor, ToolScope, *mutationStoriesStub, *usersServiceStub, *statesServiceStub) {
	t.Helper()
	scope := testToolScope()
	scope.AllowMutations = true
	team := mutationTestTeam(scope.WorkspaceID)
	scope.AllowedTeamIDs = []uuid.UUID{team.ID}
	scope.SourceScope = &SourceScope{Surface: SourceSurfaceStoryComment, StoryID: uuid.New(), TeamID: team.ID}
	backend := newMutationStoriesStub()
	text, markup := "Old prose", `<p>Old prose</p><img src="https://files.example/image.png" data-attachment-id="image-id" data-width="70%" data-align="right"><video src="https://files.example/demo.mp4" data-document-media-video="true" data-attachment-id="video-id" controls></video>`
	backend.persisted[scope.SourceScope.StoryID] = stories.CoreSingleStory{ID: scope.SourceScope.StoryID, SequenceID: 42, Title: "Current story", Priority: "Low", Workspace: scope.WorkspaceID, Team: team.ID, Description: &text, DescriptionHTML: &markup, UpdatedAt: time.Now().UTC()}
	members := &usersServiceStub{items: []users.CoreUser{{ID: uuid.New(), FullName: "Alex Example", IsActive: true}}}
	statuses := &statesServiceStub{items: []states.CoreState{
		{ID: uuid.New(), Workspace: scope.WorkspaceID, Team: team.ID, Name: "Ready", Category: "unstarted", OrderIndex: 2},
		{ID: uuid.New(), Workspace: scope.WorkspaceID, Team: team.ID, Name: "Done", Category: "completed", OrderIndex: 5},
		{ID: uuid.New(), Workspace: scope.WorkspaceID, Team: team.ID, Name: "Cancelled", Category: "cancelled", IsDefault: true},
	}}
	executor := newToolExecutorForTest(t, &teamsServiceStub{joined: []teams.CoreTeam{team}}, backend, &searchServiceStub{}, &objectivesServiceStub{}, WithOperationalTools(OperationalToolServices{States: statuses, Users: members}), WithStoryMutations(testStoryMutationSecret), WithStoryMutationConfirmationStore(backend.confirmations))
	return executor, scope, backend, members, statuses
}

func nativeArguments(t *testing.T, scope ToolScope, changes map[string]any) json.RawMessage {
	t.Helper()
	args := map[string]any{"story_id": scope.SourceScope.StoryID.String(), "story_reference": nil, "title": nil, "priority": nil, "assignee": "unchanged", "estimated_duration_action": "unchanged", "estimated_duration_minutes": nil, "minimum_focus_block_action": "unchanged", "minimum_focus_block_minutes": nil, "command_intent": "direct_command", "status_action": "unchanged"}
	for key, value := range changes {
		args[key] = value
	}
	encoded, err := json.Marshal(args)
	require.NoError(t, err)
	return encoded
}

func nativeProposal(t *testing.T, executor *FortyOneToolExecutor, scope ToolScope, changes map[string]any) *StoryMutationConfirmation {
	t.Helper()
	if _, rewriting := changes["description"]; rewriting {
		story, err := executor.mutations.stories.Get(context.Background(), scope.SourceScope.StoryID, scope.WorkspaceID)
		require.NoError(t, err)
		changes["expected_updated_at"] = story.UpdatedAt.Format(time.RFC3339Nano)
	}
	output, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolUpdateStory, Arguments: nativeArguments(t, scope, changes)})
	require.NoError(t, err)
	confirmation, proposed, err := mutationConfirmationFromToolResult(output)
	require.NoError(t, err)
	require.True(t, proposed)
	return confirmation
}

func TestNativeDescriptionProposalPreservesMediaUsesCASAndDurableReceipt(t *testing.T) {
	t.Parallel()
	executor, scope, backend, _, _ := nativeFixture(t)
	replacement := strings.Repeat("A clear requirement. ", 250) + "\n\n<script>alert('text')</script>"
	proposal := nativeProposal(t, executor, scope, map[string]any{"description": replacement})
	require.Less(t, len(proposal.Token), 100, "large rich descriptions use an opaque durable proposal")
	require.Empty(t, backend.updateCalls)
	inspection, err := executor.InspectStoryMutation(context.Background(), scope, proposal.Token)
	require.NoError(t, err)
	require.True(t, inspection.DirectEligible)
	require.Equal(t, scope.SourceScope.StoryID, inspection.StoryID)
	require.Equal(t, []string{"description"}, inspection.ChangedFields)
	_, err = executor.ConfirmStoryMutation(context.Background(), scope, proposal.Token)
	require.NoError(t, err)
	require.Len(t, backend.updateCalls, 1)
	updates := backend.updateCalls[0].updates
	require.Equal(t, replacement, updates["description"])
	markup := updates["description_html"].(string)
	require.Contains(t, markup, `data-attachment-id="image-id"`)
	require.Contains(t, markup, `data-width="70%"`)
	require.Contains(t, markup, `data-align="right"`)
	require.Contains(t, markup, `data-document-media-video="true"`)
	require.Contains(t, markup, `data-attachment-id="video-id"`)
	require.NotContains(t, markup, "Old prose")
	require.NotContains(t, markup, "<script>")
	require.Contains(t, markup, "&lt;script&gt;")
	retry, err := executor.ConfirmStoryMutation(context.Background(), scope, proposal.Token)
	require.NoError(t, err)
	require.Equal(t, storyMutationStatusAlreadyApplied, retry.Status)
	require.Len(t, backend.updateCalls, 1)

	stale := nativeProposal(t, executor, scope, map[string]any{"description": "Later rewrite"})
	current := backend.persisted[scope.SourceScope.StoryID]
	current.UpdatedAt = current.UpdatedAt.Add(time.Second)
	backend.persisted[current.ID] = current
	staleReview, err := executor.InspectStoryMutation(context.Background(), scope, stale.Token)
	require.NoError(t, err)
	require.Contains(t, staleReview.Review, "Later rewrite")
	_, err = executor.ConfirmStoryMutation(context.Background(), scope, stale.Token)
	require.ErrorIs(t, err, ErrStaleMutation)
	require.Len(t, backend.updateCalls, 1)
}

func TestNativeScopeDeniesOtherStoriesToolsAndIntegrationTokens(t *testing.T) {
	t.Parallel()
	executor, scope, backend, _, _ := nativeFixture(t)
	for _, name := range []string{toolListTeams, toolListMyTasks, toolSearchWork, toolListObjectives, toolGetSprintSummary, toolCreateStory, toolAddComment, toolAddRelationship, toolAttachStoryFile} {
		_, err := executor.Execute(context.Background(), scope, ToolCall{Name: name, Arguments: json.RawMessage(`{}`)})
		require.ErrorIs(t, err, ErrMutationNotAllowed, name)
	}
	_, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolGetStory, Arguments: json.RawMessage(`{"story_reference":"WEB-999"}`)})
	require.ErrorIs(t, err, ErrTeamNotAccessible)
	_, err = executor.Execute(context.Background(), scope, ToolCall{Name: toolUpdateStory, Arguments: nativeArguments(t, scope, map[string]any{"story_id": uuid.New().String(), "title": "Outside"})})
	require.ErrorIs(t, err, ErrTeamNotAccessible)
	proposal := nativeProposal(t, executor, scope, map[string]any{"title": "Allowed"})
	outside := scope
	outside.SourceScope = &SourceScope{Surface: SourceSurfaceStoryComment, StoryID: uuid.New(), TeamID: scope.SourceScope.TeamID}
	_, err = executor.InspectStoryMutation(context.Background(), outside, proposal.Token)
	require.ErrorIs(t, err, ErrInvalidConfirmation)
	_, err = executor.ConfirmStoryMutation(context.Background(), outside, proposal.Token)
	require.ErrorIs(t, err, ErrInvalidConfirmation)
	unscoped := scope
	unscoped.SourceScope = nil
	_, err = executor.ConfirmStoryMutation(context.Background(), unscoped, proposal.Token)
	require.ErrorIs(t, err, ErrInvalidConfirmation)
	_, err = executor.CancelStoryMutation(context.Background(), outside, proposal.Token)
	require.ErrorIs(t, err, ErrInvalidConfirmation)

	args := nativeArguments(t, scope, map[string]any{"title": "Integration proposal", "command_intent": "", "status_action": ""})
	output, err := executor.Execute(context.Background(), unscoped, ToolCall{Name: toolUpdateStory, Arguments: args})
	require.NoError(t, err)
	legacy, _, err := mutationConfirmationFromToolResult(output)
	require.NoError(t, err)
	_, err = executor.ConfirmStoryMutation(context.Background(), scope, legacy.Token)
	require.ErrorIs(t, err, ErrInvalidConfirmation)
	require.Empty(t, backend.updateCalls)
}

func TestNativeNamedAssignmentRechecksMembershipAndApprovalIntent(t *testing.T) {
	t.Parallel()
	executor, scope, backend, members, _ := nativeFixture(t)
	proposal := nativeProposal(t, executor, scope, map[string]any{"assignee": "named", "assignee_id": members.items[0].ID.String(), "command_intent": "proposal"})
	inspection, err := executor.InspectStoryMutation(context.Background(), scope, proposal.Token)
	require.NoError(t, err)
	require.False(t, inspection.DirectEligible)
	require.Equal(t, []string{"assignee_id"}, inspection.ChangedFields)
	require.Equal(t, scope.SourceScope.TeamID, *members.lastListCall().filter.TeamID)
	members.items[0].IsActive = false
	_, err = executor.ConfirmStoryMutation(context.Background(), scope, proposal.Token)
	require.ErrorIs(t, err, ErrInvalidConfirmation)
	require.Empty(t, backend.updateCalls)
	members.items[0].IsActive = true
	_, err = executor.ConfirmStoryMutation(context.Background(), scope, proposal.Token)
	require.NoError(t, err)
	require.Equal(t, members.items[0].ID, backend.updateCalls[0].updates["assignee_id"])
}

func TestNativeCloseAndReopenResolveCurrentTeamWorkflow(t *testing.T) {
	t.Parallel()
	for _, action := range []string{"close", "reopen"} {
		t.Run(action, func(t *testing.T) {
			executor, scope, backend, _, statuses := nativeFixture(t)
			proposal := nativeProposal(t, executor, scope, map[string]any{"status_action": action})
			_, err := executor.ConfirmStoryMutation(context.Background(), scope, proposal.Token)
			require.NoError(t, err)
			want := statuses.items[1].ID
			if action == "reopen" {
				want = statuses.items[0].ID
			}
			require.Equal(t, want, backend.updateCalls[0].updates["status_id"])
		})
	}
}

func TestNativeCatalogIsIsolatedAndLegacyCatalogUnchanged(t *testing.T) {
	t.Parallel()
	executor, scope, _, _, _ := nativeFixture(t)
	definitions := executor.Definitions()
	native := sourceToolDefinitions(definitions, scope.SourceScope)
	require.Len(t, native, 4)
	require.NoError(t, validateToolDefinitions(native))
	for _, definition := range native {
		require.True(t, sourceAllowsTool(scope.SourceScope, definition.Name))
	}
	for _, definition := range definitions {
		if definition.Name == toolUpdateStory {
			require.NotContains(t, definition.Parameters["properties"], "description")
			require.NotContains(t, definition.Parameters["properties"], "command_intent")
		}
	}
	invalid := scope
	invalid.SourceScope = &SourceScope{Surface: "slack", StoryID: scope.SourceScope.StoryID, TeamID: scope.SourceScope.TeamID}
	require.ErrorIs(t, validateToolScope(&invalid), ErrInvalidRequest)
}

func TestNativeNeedsInputUsesExplicitResponseContract(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"completed","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"[needs_input] Please choose which Alex to assign."}]}]}`))
	}))
	defer server.Close()
	assistant, err := NewOpenAIAssistant(OpenAIConfig{APIKey: "test", BaseURL: server.URL}, &assistantToolStub{})
	require.NoError(t, err)
	request := Request{WorkspaceID: uuid.New(), UserID: uuid.New(), SourceScope: &SourceScope{Surface: SourceSurfaceStoryComment, StoryID: uuid.New(), TeamID: uuid.New()}, Prompt: "Assign Alex"}
	response, err := assistant.Respond(context.Background(), request)
	require.NoError(t, err)
	require.True(t, response.NeedsInput)
	require.Equal(t, "Please choose which Alex to assign.", response.Text)
	text, needsInput := nativeResponseText("A normal question?", request.SourceScope)
	require.False(t, needsInput)
	require.Equal(t, "A normal question?", text)
}

func TestNativeRequestUsesMarkdownAndReviewInstructionsWithoutChangingOtherSurfaces(t *testing.T) {
	t.Parallel()
	_, scope, _, _, _ := nativeFixture(t)
	base := "Use the external provider's formatting and confirmation controls."
	native, err := instructionsForRequest(base, Request{SourceScope: scope.SourceScope, AllowMutations: true})
	require.NoError(t, err)
	require.Contains(t, native, "using Markdown when formatting helps")
	require.Contains(t, native, "Do not emit raw HTML or embed images or other media")
	require.Contains(t, native, "call update_story with command_intent proposal and the exact proposed value")
	require.Contains(t, native, `"do not change it yet": preparing a reviewable proposal does not change the story`)
	require.Contains(t, native, "General informational advice without a concrete requested field change remains an ordinary answer")
	require.Contains(t, native, "resolve that person uniquely through list_team_members")
	require.Contains(t, native, "[@Display Name](mention:MEMBER_UUID)")
	require.Contains(t, native, "Do not tag people merely because a name occurs in context or quoted material")
	require.NotContains(t, native, base)
	provider, err := instructionsForRequest(base, Request{AllowMutations: true})
	require.NoError(t, err)
	require.Equal(t, base, provider)
}

func TestNativeResponsePrefixPreservesIndentedMentionExamples(t *testing.T) {
	t.Parallel()
	source := &SourceScope{Surface: SourceSurfaceStoryComment, StoryID: uuid.New(), TeamID: uuid.New()}
	example := "    [@Example](mention:" + uuid.NewString() + ")"
	for _, input := range []string{example, nativeNeedsInputPrefix + "\n" + example, nativeNeedsInputPrefix + " " + example} {
		text, _ := nativeResponseText(input, source)
		require.Equal(t, example, text)
	}
}

func TestNativeAssistantPreservesMentionExampleIndentation(t *testing.T) {
	t.Parallel()
	example := "    [@Example](mention:" + uuid.NewString() + ")"
	output, err := json.Marshal(map[string]any{
		"status": "completed",
		"output": []any{map[string]any{"type": "message", "role": "assistant", "content": []any{
			map[string]any{"type": "output_text", "text": example},
		}}},
	})
	require.NoError(t, err)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(output)
	}))
	defer server.Close()
	assistant, err := NewOpenAIAssistant(OpenAIConfig{APIKey: "test", BaseURL: server.URL}, &assistantToolStub{})
	require.NoError(t, err)
	request := Request{WorkspaceID: uuid.New(), UserID: uuid.New(), SourceScope: &SourceScope{
		Surface: SourceSurfaceStoryComment, StoryID: uuid.New(), TeamID: uuid.New(),
	}, Prompt: "Show me a code example"}
	response, err := assistant.Respond(t.Context(), request)
	require.NoError(t, err)
	require.Equal(t, example, response.Text)
	request.SourceScope = nil
	response, err = assistant.Respond(t.Context(), request)
	require.NoError(t, err)
	require.Equal(t, strings.TrimSpace(example), response.Text, "provider response normalization remains unchanged")
}

func TestDescriptionRewriteKeepsStructuredContentAndConsistentText(t *testing.T) {
	t.Parallel()
	original := `<p>Remove old prose</p><figure><img src="image.png"><figcaption>Chart caption</figcaption></figure><div data-file-id="file"><a href="https://files.example/report.pdf">Report.pdf</a></div><ul data-type="taskList"><li data-type="taskItem" data-checked="true"><p>Checked item</p></li></ul><table><tbody><tr><td>Cell</td></tr></tbody></table>`
	text, markup, err := rewriteDescription(&original, "New line\nSecond line\n\nNext paragraph")
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(markup, "<p>New line<br>Second line</p><p>Next paragraph</p>"))
	require.Contains(t, text, "Chart caption")
	require.Contains(t, text, "Report.pdf")
	require.Contains(t, text, "Checked item")
	require.Contains(t, text, "Cell")
	require.NotContains(t, text, "Remove old prose")
	require.Equal(t, 1, strings.Count(markup, "<img"))
	require.Contains(t, markup, `data-checked="true"`)
	_, _, err = rewriteDescription(nil, strings.Repeat("x", maximumNativeDescriptionRunes+1))
	require.ErrorIs(t, err, ErrInvalidToolArguments)
}

func TestNativeAlreadyClosedReturnsAnswerWithoutProposalOrWrite(t *testing.T) {
	t.Parallel()
	executor, scope, backend, _, statuses := nativeFixture(t)
	story := backend.persisted[scope.SourceScope.StoryID]
	story.Status = &statuses.items[1].ID
	backend.persisted[story.ID] = story
	arguments := nativeArguments(t, scope, map[string]any{"status_action": "close"})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "completed", "output": []map[string]any{{"type": "function_call", "call_id": "close", "name": toolUpdateStory, "arguments": string(arguments)}}})
	}))
	defer server.Close()
	assistant := mustTestAssistant(t, server.URL, executor, OpenAIConfig{})
	response, err := assistant.Respond(context.Background(), Request{WorkspaceID: scope.WorkspaceID, UserID: scope.UserID, AllowedTeamIDs: scope.AllowedTeamIDs, SourceScope: scope.SourceScope, AllowMutations: true, Prompt: "Close this story"})
	require.NoError(t, err)
	require.Equal(t, "WEB-42 already matches the requested changes.", response.Text)
	require.Nil(t, response.Confirmation)
	require.Empty(t, backend.updateCalls)
	require.Empty(t, backend.confirmations.entries)
}

func TestNativeStoryReadExcludesPrivateAndRichDocumentFields(t *testing.T) {
	t.Parallel()
	executor, scope, backend, _, _ := nativeFixture(t)
	story := backend.persisted[scope.SourceScope.StoryID]
	story.WatcherIDs = []uuid.UUID{uuid.New()}
	story.Collaborators = []uuid.UUID{uuid.New()}
	story.Reporter = &scope.UserID
	privateKey := "internal-creation-key"
	story.CreationKey = &privateKey
	backend.persisted[story.ID] = story
	output, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolGetStory, Arguments: json.RawMessage(`{"story_reference":"WEB-42"}`)})
	require.NoError(t, err)
	var fields map[string]any
	require.NoError(t, json.Unmarshal(output, &fields))
	require.Equal(t, story.ID.String(), fields["id"])
	require.Equal(t, "Old prose", fields["description"])
	for _, field := range []string{"watcher_ids", "collaborators", "reporter", "description_html", "creation_key", "memories", "email"} {
		require.NotContains(t, fields, field)
	}
	require.NotContains(t, string(output), "image-id")
	require.NotContains(t, string(output), privateKey)
	longText := strings.Repeat("x", maximumNativeDescriptionRunes+1)
	story.Description = &longText
	backend.persisted[story.ID] = story
	_, err = executor.Execute(context.Background(), scope, ToolCall{Name: toolUpdateStory, Arguments: nativeArguments(t, scope, map[string]any{"description": "Truncated-context rewrite"})})
	require.ErrorIs(t, err, ErrInvalidToolArguments)
}

func TestDescriptionRewriteRetainsSafeReferencesAndRejectsActiveMedia(t *testing.T) {
	t.Parallel()
	original := `<p>Old <a href="https://example.com/spec">Reference</a> <a href="/workspace/work/WEB-43">Relative reference</a><a data-type="mention" href="/profile/person">@Person</a><a data-mention-kind="agent" href="/maya">@Maya</a></p><a href="https://example.com/image"><img src="image.png" data-attachment-id="image"></a>`
	text, markup, err := rewriteDescription(&original, "Rewritten prose")
	require.NoError(t, err)
	require.Equal(t, "Rewritten prose\n\nReference\n\nRelative reference", text)
	require.Contains(t, markup, `<p><a href="https://example.com/spec">Reference</a></p>`)
	require.Contains(t, markup, `href="/workspace/work/WEB-43"`)
	require.Contains(t, markup, `href="https://example.com/image"`)
	require.Equal(t, 1, strings.Count(markup, "<img"))
	require.NotContains(t, markup, "@Person")
	require.NotContains(t, markup, "@Maya")
	for _, unsafe := range []string{
		`<img src="image.png" onerror="alert(1)">`,
		`<a href="java&#x09;script:alert(1)">Unsafe</a>`,
		`<iframe srcdoc="<script>alert(1)</script>"></iframe>`,
		`<img src="image.png" style="width:expression(alert(1))">`,
		`<div data-file-id="file"><script>alert(1)</script></div>`,
	} {
		_, _, err := rewriteDescription(&unsafe, "Replacement")
		require.ErrorIs(t, err, ErrInvalidToolArguments, unsafe)
	}
}

func TestNativeDescriptionRejectsEditsBetweenReadAndProposal(t *testing.T) {
	t.Parallel()
	executor, scope, backend, _, _ := nativeFixture(t)
	story := backend.persisted[scope.SourceScope.StoryID]
	readVersion := story.UpdatedAt.Format(time.RFC3339Nano)
	story.UpdatedAt = story.UpdatedAt.Add(time.Second)
	backend.persisted[story.ID] = story
	_, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolUpdateStory, Arguments: nativeArguments(t, scope, map[string]any{"description": "Rewrite based on stale context", "expected_updated_at": readVersion})})
	require.ErrorIs(t, err, ErrStaleMutation)
	require.Empty(t, backend.confirmations.entries)
	require.Empty(t, backend.updateCalls)
}
