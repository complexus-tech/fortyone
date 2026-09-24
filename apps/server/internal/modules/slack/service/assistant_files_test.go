package slack

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/stretchr/testify/require"
)

func TestAssistantAvailableFilesUsesCurrentMessageAndThreadWithoutExternalFiles(t *testing.T) {
	t.Parallel()
	event := normalizedSlackEvent{
		TeamID: "T1", ChannelID: "C1", ThreadTS: "171.100", MessageTS: "171.103",
		Files: []slackMessageFile{{ID: "F3", Name: "current.png"}},
	}
	files := assistantAvailableFiles(event, []slackMessageFile{
		{ID: "F1", Name: "earlier.pdf", MessageTS: "171.101"},
		{ID: "F2", Name: "external", IsExternal: true, MessageTS: "171.102"},
		{ID: "F3", Name: "duplicate.png", MessageTS: "171.103"},
	})
	require.Len(t, files, 2)
	require.Equal(t, "F3", files[0].FileID)
	require.Equal(t, "171.103", files[0].MessageTS)
	require.Equal(t, "F1", files[1].FileID)
	require.Equal(t, "171.101", files[1].MessageTS)
	for _, file := range files {
		require.Equal(t, "T1", file.ExternalWorkspaceID)
		require.Equal(t, "C1", file.ChannelID)
		require.Equal(t, "171.100", file.ThreadTS)
	}
	turn, err := assistantFileReferenceTurn(files)
	require.NoError(t, err)
	require.Contains(t, turn.Text, `"file_id":"F1"`)
	require.NotContains(t, turn.Text, "https://")
	require.NotContains(t, turn.Text, "external")
}

func TestAssistantAvailableFilesKeepsIDOnlyCurrentFile(t *testing.T) {
	t.Parallel()
	event := normalizedSlackEvent{
		TeamID: "T1", ChannelID: "D1", ThreadTS: "171.103", MessageTS: "171.103",
		Files: []slackMessageFile{{ID: "F1"}},
	}
	files := assistantAvailableFiles(event, nil)
	require.Len(t, files, 1)
	require.Equal(t, "F1", files[0].FileID)
	require.Equal(t, "Slack file", files[0].Name)
	require.Equal(t, "171.103", files[0].MessageTS)
	require.Equal(t, "171.103", files[0].ThreadTS)
}

func TestLoadSlackDirectMessageFilesUsesRecentOwnFileMessages(t *testing.T) {
	t.Parallel()
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != "/conversations.history" {
			t.Errorf("unexpected Slack history request: %s %s", request.Method, request.URL)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		query := request.URL.Query()
		if query.Get("channel") != "D1" || query.Get("latest") != "171.105" || query.Get("inclusive") != "true" ||
			query.Get("limit") != strconv.Itoa(slackDirectMessageFilePageLimit) {
			t.Errorf("unexpected Slack history query: %s", request.URL.RawQuery)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true,"messages":[
			{"ts":"171.105","user":"U1","text":"Can you see the attached file?","files":[{"id":"F-current","name":"current.png","mode":"hosted"}]},
			{"ts":"171.103","user":"U1","text":"","subtype":"file_share","files":[{"id":"F1","name":"draft.png","mode":"hosted"}]},
			{"ts":"171.102","user":"U1","text":"","subtype":"file_share","files":[{"id":"F-external","name":"external.png","mode":"external","is_external":true}]},
			{"ts":"171.101","user":"U2","text":"","subtype":"file_share","files":[{"id":"F-other","name":"other.png","mode":"hosted"}]},
			{"ts":"171.100","user":"U1","bot_id":"B1","text":"","subtype":"file_share","files":[{"id":"F-bot","name":"bot.png","mode":"hosted"}]}
		]}`))
	}))
	defer provider.Close()
	client := newSlackWebClient(provider.Client())
	client.baseURL = provider.URL
	event := normalizedSlackEvent{
		Kind: slackEventKindDirect, TeamID: "T1", UserID: "U1",
		ChannelID: "D1", ThreadTS: "171.105", MessageTS: "171.105",
	}
	historyFiles, err := (&EventProcessor{webClient: client}).loadSlackDirectMessageFiles(context.Background(), "xoxb-token", event)
	require.NoError(t, err)
	files := assistantAvailableFiles(event, historyFiles)
	require.Len(t, files, 2)
	require.Equal(t, "F-current", files[0].FileID)
	require.Equal(t, "171.105", files[0].MessageTS)
	require.Equal(t, "171.105", files[0].ThreadTS)
	require.Equal(t, "F1", files[1].FileID)
	require.Equal(t, "171.103", files[1].MessageTS)
	require.Equal(t, "171.103", files[1].ThreadTS)
}

func TestAssistantLoadsDirectMessageFilesForShortFollowUp(t *testing.T) {
	t.Parallel()
	event := normalizedSlackEvent{
		Kind: slackEventKindDirect, UserID: "U1", ChannelID: "D1", MessageTS: "171.105",
	}
	currentMessageTS := "171.105"
	history := []messageRecord{
		{Role: "user", Content: "Can you see the attached file?"},
		{Role: "assistant", Content: "Please attach it again."},
		{Role: "user", Content: "check", ExternalMessageID: &currentMessageTS},
	}
	require.True(t, assistantShouldLoadDirectMessageFiles(event, "check"))
	require.True(t, assistantRecentlyAskedAboutFile(event, history))
	require.True(t, assistantRecentlyAskedAboutFile(event, []messageRecord{
		{Role: "user", Content: "Can you see the attached file?"},
		{Role: "user", Content: "check"},
	}))
	require.True(t, assistantShouldLoadDirectMessageFiles(event, "check again"))
	require.False(t, assistantRecentlyAskedAboutFile(event, []messageRecord{
		{Role: "user", Content: "What stories are due?"},
	}))
	require.False(t, assistantShouldLoadDirectMessageFiles(event, "What stories are due now?"))
	require.False(t, assistantShouldLoadDirectMessageFiles(event, "Check the story status"))
	require.True(t, assistantShouldLoadDirectMessageFiles(event, "Can you see the attached file?"))
	require.True(t, assistantShouldLoadDirectMessageFiles(event, "Attach it to WEB-550"))
	require.True(t, slackPromptRequestsFileContext("Attach it to WEB-550"))
	require.False(t, assistantShouldLoadDirectMessageFiles(event, "Which files are in FortyOne?"))
	require.False(t, slackPromptRequestsFileContext("Upload this report to Google Drive"))
	require.False(t, slackPromptRequestsFileContext("Check that PDF in a story"))
}

func TestAssistantReplyCanAttachFileFromEarlierDirectMessage(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name          string
		prompt        string
		priorQuestion bool
	}{
		{name: "prior file question", prompt: "check", priorQuestion: true},
		{name: "file-only upload", prompt: "attach it to WEB-550"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				if request.Method != http.MethodGet || request.URL.Path != "/conversations.history" ||
					request.Header.Get("Authorization") != "Bearer "+testSlackBotAccessToken {
					t.Errorf("unexpected Slack history request: %s %s", request.Method, request.URL)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				_, _ = w.Write([]byte(`{"ok":true,"messages":[
					{"ts":"171.103","user":"U1","text":"check"},
					{"ts":"171.101","user":"U1","text":"","subtype":"file_share","files":[{"id":"F1","name":"draft.png","mode":"hosted"}]}
				]}`))
			}))
			defer provider.Close()

			repo := newEventRepositoryStub()
			scope := "files:read,im:history"
			repo.installation.Scope = &scope
			store := newEventStoreStub()
			previousTS := "171.100"
			currentTS := "171.103"
			if testCase.priorQuestion {
				store.history = []messageRecord{
					{Role: "user", Content: "Can you see the attached file?", ExternalMessageID: &previousTS},
					{Role: "assistant", Content: "Please check again."},
				}
			}
			store.history = append(store.history, messageRecord{
				Role: "user", Content: testCase.prompt, ExternalMessageID: &currentTS,
			})
			assistant := &assistantStub{response: AssistantResponse{Text: "I can see a Slack file is attached."}}
			processor := newTestEventProcessorWithBudgets(t, repo, store, assistant, &accessCheckerStub{allowed: true},
				&messageSenderStub{}, &callLimiterStub{decision: AssistantAdmissionDecision{Allowed: true}}, &usageBudgetStub{})
			processor.webClient = newSlackWebClient(provider.Client())
			processor.webClient.baseURL = provider.URL

			state := &assistantDeliveryState{conversationID: testConversationID}
			err := processor.generateAssistantReply(context.Background(), assistantEventInput{
				receipt:      inboundEventRecord{ID: testInboundReceiptID, AttemptCount: 1},
				workspace:    repo.workspace,
				installation: repo.installation,
				linkedUserID: testLinkedUserID,
				event: normalizedSlackEvent{
					EventID: "Ev-check", Kind: slackEventKindDirect, TeamID: "T1", UserID: "U1",
					ChannelID: "D1", ThreadTS: currentTS, MessageTS: currentTS,
				},
				botToken: testSlackBotAccessToken,
				prompt:   testCase.prompt,
			}, state)
			require.NoError(t, err)
			require.Equal(t, "I can see a Slack file is attached.", state.reply)
			require.Len(t, assistant.requests, 1)
			require.Len(t, assistant.requests[0].AvailableFiles, 1)
			require.Equal(t, "F1", assistant.requests[0].AvailableFiles[0].FileID)
			require.Equal(t, "171.101", assistant.requests[0].AvailableFiles[0].MessageTS)
			require.Equal(t, "171.101", assistant.requests[0].AvailableFiles[0].ThreadTS)
			require.Contains(t, assistant.requests[0].Conversation[len(assistant.requests[0].Conversation)-1].Text, `"file_id":"F1"`)
		})
	}
}

func TestAssistantFileRequestWithoutScopeRepliesWithReconnectInstructions(t *testing.T) {
	t.Parallel()
	store := newEventStoreStub()
	assistant := &assistantStub{}
	processor := &EventProcessor{store: store, assistant: assistant}
	state := &assistantDeliveryState{}
	err := processor.generateAssistantReply(context.Background(), assistantEventInput{
		event: normalizedSlackEvent{
			Kind: slackEventKindDirect, TeamID: "T1", UserID: "U1", ChannelID: "D1",
			MessageTS: "171.103", Files: []slackMessageFile{{ID: "F1"}},
		},
		prompt: "Can you see the attached file?",
	}, state)
	require.NoError(t, err)
	require.Contains(t, state.reply, "files:read")
	require.Contains(t, state.reply, "Update connection")
	require.Empty(t, assistant.requests)
}

func TestAssistantUnrelatedFileRequestWithoutScopeReachesAssistant(t *testing.T) {
	t.Parallel()
	repo := newEventRepositoryStub()
	store := newEventStoreStub()
	assistant := &assistantStub{response: AssistantResponse{Text: "I can help with that."}}
	processor := newTestEventProcessorWithBudgets(t, repo, store, assistant, &accessCheckerStub{allowed: true},
		&messageSenderStub{}, &callLimiterStub{decision: AssistantAdmissionDecision{Allowed: true}}, &usageBudgetStub{})
	state := &assistantDeliveryState{conversationID: testConversationID}
	err := processor.generateAssistantReply(context.Background(), assistantEventInput{
		receipt:      inboundEventRecord{ID: testInboundReceiptID, AttemptCount: 1},
		workspace:    repo.workspace,
		installation: repo.installation,
		linkedUserID: testLinkedUserID,
		event: normalizedSlackEvent{
			EventID: "Ev-drive", Kind: slackEventKindDirect, TeamID: "T1", UserID: "U1",
			ChannelID: "D1", ThreadTS: "171.103", MessageTS: "171.103",
		},
		prompt: "Upload this report to Google Drive",
	}, state)
	require.NoError(t, err)
	require.Equal(t, "I can help with that.", state.reply)
	require.Len(t, assistant.requests, 1)
}

func TestAssistantReplySeesEarlierFileOnShortThreadFollowUp(t *testing.T) {
	t.Parallel()
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != "/conversations.replies" {
			t.Errorf("unexpected Slack replies request: %s %s", request.Method, request.URL)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true,"messages":[
			{"ts":"171.100","user":"U1","text":"Start"},
			{"ts":"171.101","thread_ts":"171.100","user":"U1","text":"","subtype":"file_share","files":[{"id":"F1","name":"draft.png","mode":"hosted"}]},
			{"ts":"171.103","thread_ts":"171.100","user":"U1","text":"check"}
		]}`))
	}))
	defer provider.Close()

	repo := newEventRepositoryStub()
	scope := "files:read,groups:history"
	repo.installation.Scope = &scope
	store := newEventStoreStub()
	currentTS := "171.103"
	store.history = []messageRecord{{Role: "user", Content: "check", ExternalMessageID: &currentTS}}
	assistant := &assistantStub{response: AssistantResponse{Text: "I can see the file in this thread."}}
	processor := newTestEventProcessorWithBudgets(t, repo, store, assistant, &accessCheckerStub{allowed: true},
		&messageSenderStub{}, &callLimiterStub{decision: AssistantAdmissionDecision{Allowed: true}}, &usageBudgetStub{})
	processor.webClient = newSlackWebClient(provider.Client())
	processor.webClient.baseURL = provider.URL

	state := &assistantDeliveryState{conversationID: testConversationID}
	err := processor.generateAssistantReply(context.Background(), assistantEventInput{
		receipt:      inboundEventRecord{ID: testInboundReceiptID, AttemptCount: 1},
		workspace:    repo.workspace,
		installation: repo.installation,
		linkedUserID: testLinkedUserID,
		event: normalizedSlackEvent{
			EventID: "Ev-check", Kind: slackEventKindChannelThread, TeamID: "T1", UserID: "U1",
			ChannelID: "C1", ThreadTS: "171.100", ReplyTS: "171.100", MessageTS: currentTS,
		},
		botToken: testSlackBotAccessToken,
		prompt:   "check",
	}, state)
	require.NoError(t, err)
	require.Equal(t, "I can see the file in this thread.", state.reply)
	require.Len(t, assistant.requests, 1)
	require.Len(t, assistant.requests[0].AvailableFiles, 1)
	require.Equal(t, "F1", assistant.requests[0].AvailableFiles[0].FileID)
	require.Equal(t, "171.101", assistant.requests[0].AvailableFiles[0].MessageTS)
	require.Equal(t, "171.100", assistant.requests[0].AvailableFiles[0].ThreadTS)
}

func TestSlackDirectMessageFileFailureReply(t *testing.T) {
	t.Parallel()
	require.Contains(t, slackDirectMessageFileFailureReply(&SlackAPIError{Code: "missing_scope"}), "im:history")
	require.Contains(t, slackDirectMessageFileFailureReply(&RateLimitError{}), "try again in a minute")
	require.NotContains(t, slackDirectMessageFileFailureReply(&SlackAPIError{Code: "internal_error"}), "attach")
}

func TestSlackFileLanguageRequestsThreadHydration(t *testing.T) {
	t.Parallel()
	for _, prompt := range []string{"Attach the design file", "Upload that document to WEB-7", "Use the photo above"} {
		require.True(t, slackPromptRequestsThreadContext(prompt), prompt)
	}
	require.False(t, slackPromptRequestsThreadContext(strings.Repeat("status ", 2)))
}

func TestSlackThreadReferenceIncludesFileOnlyMessagesEvenWhenTextWasPersisted(t *testing.T) {
	t.Parallel()
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		require.Equal(t, "/conversations.replies", request.URL.Path)
		_, _ = w.Write([]byte(`{"ok":true,"messages":[
			{"ts":"171.100","user":"U1","text":"Start"},
			{"ts":"171.101","thread_ts":"171.100","user":"U2","text":"","subtype":"file_share","files":[{"id":"F1","name":"draft.pdf","mode":"hosted"}]}
		]}`))
	}))
	defer provider.Close()
	client := newSlackWebClient(provider.Client())
	client.baseURL = provider.URL
	reference, err := (&EventProcessor{webClient: client}).loadSlackThreadReference(
		context.Background(), "xoxb-token",
		slackrepository.SlackWorkspaceRecord{SlackTeamDomain: "acme"},
		normalizedSlackEvent{ChannelID: "C1", ThreadTS: "171.100", MessageTS: "171.102"},
		map[string]struct{}{"171.101": {}},
	)
	require.NoError(t, err)
	require.Len(t, reference.Files, 1)
	require.Equal(t, "F1", reference.Files[0].ID)
	require.Equal(t, "171.101", reference.Files[0].MessageTS)
	require.NotContains(t, reference.Turn.Text, `"timestamp":"171.101"`)
}
