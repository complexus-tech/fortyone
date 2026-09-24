package slack

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSourceFilesFromShortcutKeepsHostedFilesOnly(t *testing.T) {
	files := []slackPayloadFile{
		{ID: "F123", Name: "plan.pdf", Mode: "hosted"},
		{ID: "F123", Name: "duplicate.pdf", Mode: "hosted"},
		{ID: "F234", Title: "design.png"},
		{ID: "F345", Name: "external.docx", Mode: "external"},
		{ID: "F456", Name: "drive.pdf", IsExternal: true},
		{ID: "https://example.com/file", Name: "invalid.txt"},
	}

	require.Equal(t, []slackSourceFile{
		{ID: "F123", Name: "plan.pdf"},
		{ID: "F234", Name: "design.png"},
	}, sourceFilesFromShortcut(files))
}

func TestMessageShortcutIncludesFilesFromSelectedMessage(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	grantedScopes := slackBotOAuthScopeValue()
	repo := &mockRepo{
		teams: []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			teamID: {{ID: uuid.New(), Name: "Todo", Category: "unstarted"}},
		},
		teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:    workspaceID,
			SlackTeamID:    "T123",
			BotAccessToken: "xoxb-token",
			Scope:          &grantedScopes,
			IsActive:       true,
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": actorID},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	var updatedView map[string]any
	service.client = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		switch request.URL.Path {
		case "/api/views.open":
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"ok":true,"view":{"id":"V123"}}`)), Header: make(http.Header)}, nil
		case "/api/views.update":
			var payload struct {
				View map[string]any `json:"view"`
			}
			if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
				return nil, err
			}
			updatedView = payload.View
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"ok":true}`)), Header: make(http.Header)}, nil
		default:
			t.Fatalf("unexpected Slack endpoint %s", request.URL)
			return nil, nil
		}
	})}
	var payload interactionPayload
	require.NoError(t, json.Unmarshal([]byte(`{
		"type":"message_action", "trigger_id":"trigger",
		"team":{"id":"T123"}, "channel":{"id":"C123"},
		"user":{"id":"U123"},
		"message":{"text":"Create a story", "ts":"123.456", "files":[
			{"id":"F123", "name":"plan.pdf", "mode":"hosted"},
			{"id":"F234", "name":"drive.pdf", "is_external":true}
		]}
	}`), &payload))

	response, err := service.handleMessageAction(context.Background(), payload)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.NotNil(t, updatedView)
	blocks := updatedView["blocks"].([]any)
	var sourceBlock, uploadBlock map[string]any
	for _, rawBlock := range blocks {
		block := rawBlock.(map[string]any)
		switch block["block_id"] {
		case modalBlockSourceFiles:
			sourceBlock = block
		case modalBlockUploadFiles:
			uploadBlock = block
		}
	}
	require.NotNil(t, sourceBlock)
	require.NotNil(t, uploadBlock)
	options := sourceBlock["element"].(map[string]any)["options"].([]any)
	require.Len(t, options, 1)
	require.Equal(t, "F123", options[0].(map[string]any)["value"])
	metadata, err := parseSlackModalPrivateMetadata(updatedView["private_metadata"].(string))
	require.NoError(t, err)
	require.Equal(t, []slackSourceFile{{ID: "F123", Name: "plan.pdf"}}, metadata.SourceFiles)
}

func TestSlackBotHasScopeRequiresAnExactGrantedScope(t *testing.T) {
	granted := "chat:write, files:read,channels:history"
	require.True(t, slackBotHasScope(&granted, "files:read"))
	require.False(t, slackBotHasScope(nil, "files:read"))
	missing := "chat:write,files:read.limited"
	require.False(t, slackBotHasScope(&missing, "files:read"))
}

func TestCreateStoryModalOffersSourceFilesAndFileUpload(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	repo := &mockRepo{
		teams: []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			teamID: {{ID: uuid.New(), Name: "Todo", Category: "unstarted"}},
		},
		teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	sourceFiles := []slackSourceFile{{ID: "F123", Name: "plan.pdf"}, {ID: "F234", Name: "design.png"}}

	view, err := service.buildCreateTaskModalView(context.Background(), createTaskModalViewInput{
		Title:       "Ship feature",
		Source:      requestSourceContext{SlackTeamID: "T123", SlackChannelID: "C123", SlackMessageTS: "123.456"},
		SourceFiles: sourceFiles,
		EnableFiles: true,
		WorkspaceID: workspaceID,
		ActorID:     actorID,
	})
	require.NoError(t, err)
	blocks := view["blocks"].([]map[string]any)
	sourceBlock := findBlock(blocks, modalBlockSourceFiles)
	require.Equal(t, true, sourceBlock["optional"])
	sourceElement := sourceBlock["element"].(map[string]any)
	require.Equal(t, "checkboxes", sourceElement["type"])
	require.Equal(t, modalActionSourceFilesSelect, sourceElement["action_id"])
	require.Equal(t, sourceElement["options"], sourceElement["initial_options"])

	uploadBlock := findBlock(blocks, modalBlockUploadFiles)
	require.Equal(t, true, uploadBlock["optional"])
	uploadElement := uploadBlock["element"].(map[string]any)
	require.Equal(t, "file_input", uploadElement["type"])
	require.Equal(t, slackModalMaxFiles, uploadElement["max_files"])
	require.NotContains(t, uploadBlock, "dispatch_action")

	metadata, err := parseSlackModalPrivateMetadata(view["private_metadata"].(string))
	require.NoError(t, err)
	require.Equal(t, sourceFiles, metadata.SourceFiles)

	// Updating the modal after a team change must not reselect a file the user
	// unchecked. Stable file input IDs allow Slack to preserve current uploads.
	view, err = service.buildCreateTaskModalView(context.Background(), createTaskModalViewInput{
		Source:                metadata.Source,
		SourceFiles:           metadata.SourceFiles,
		SelectedSourceFileIDs: []string{"F234"},
		EnableFiles:           true,
		WorkspaceID:           workspaceID,
		ActorID:               actorID,
	})
	require.NoError(t, err)
	sourceElement = findBlockElement(view["blocks"].([]map[string]any), modalBlockSourceFiles)
	initial := sourceElement["initial_options"].([]map[string]any)
	require.Len(t, initial, 1)
	require.Equal(t, "F234", selectedOptionValue(t, initial[0]))
	uploadElement = findBlockElement(view["blocks"].([]map[string]any), modalBlockUploadFiles)
	require.Equal(t, modalActionUploadFilesInput, uploadElement["action_id"])

	view, err = service.buildCreateTaskModalView(context.Background(), createTaskModalViewInput{
		SourceFiles: sourceFiles,
		WorkspaceID: workspaceID,
		ActorID:     actorID,
	})
	require.NoError(t, err)
	blocks = view["blocks"].([]map[string]any)
	require.Empty(t, findBlock(blocks, modalBlockSourceFiles))
	require.Empty(t, findBlock(blocks, modalBlockUploadFiles))
}

func TestParseCreateStorySubmissionReadsSelectedAndUploadedFiles(t *testing.T) {
	teamID := uuid.New()
	metadata, err := json.Marshal(slackModalPrivateMetadata{
		Source:         requestSourceContext{SlackTeamID: "T123", SlackChannelID: "C123", SlackMessageTS: "123.456"},
		SelectedTeamID: teamID.String(),
		SourceFiles:    []slackSourceFile{{ID: "F123", Name: "plan.pdf"}, {ID: "F234", Name: "design.png"}},
	})
	require.NoError(t, err)
	interaction := map[string]any{
		"view": map[string]any{
			"private_metadata": string(metadata),
			"state": map[string]any{"values": map[string]any{
				modalBlockTeam: map[string]any{modalActionTeamSelect: map[string]any{
					"selected_option": map[string]any{"value": teamID.String()},
				}},
				modalBlockSourceFiles: map[string]any{modalActionSourceFilesSelect: map[string]any{
					"type": "checkboxes",
					"selected_options": []map[string]any{
						{"value": "F234"},
						{"value": "F999"},
					},
				}},
				modalBlockUploadFiles: map[string]any{modalActionUploadFilesInput: map[string]any{
					"type": "file_input",
					"files": []map[string]any{
						{"id": "F345", "name": "already-uploaded.pdf"},
						{"id": "F456", "name": "new-upload.pdf"},
					},
				}},
			}},
		},
	}
	encoded, err := json.Marshal(interaction)
	require.NoError(t, err)
	var payload interactionPayload
	require.NoError(t, json.Unmarshal(encoded, &payload))

	submission, err := parseViewSubmission(payload)
	require.NoError(t, err)
	require.Equal(t, []string{"F234"}, submission.SourceFileIDs)
	require.Equal(t, []string{"F345", "F456"}, submission.UploadedFileIDs)
}
