package agentadapter

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	agentruns "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/service"
	chatdomain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestChatFailureFeedbackRequiresAnAddressedRequest(t *testing.T) {
	runtime := &ChatRuntime{}
	failure := errors.New("temporary failure")
	run := agentruns.Run{}
	require.True(t, runtime.CanPublishFailure(run, failure), "explicit asks retain useful failure feedback")
	run.Payload = json.RawMessage(`{"followup":true}`)
	require.False(t, runtime.CanPublishFailure(run, failure), "an unclassified message never requested assistant feedback")
	require.True(t, runtime.CanPublishFailure(run, &addressedChatFailure{cause: failure}))
	run.Proposal = &agentruns.Proposal{Reply: "A classified answer"}
	require.True(t, runtime.CanPublishFailure(run, failure), "a lost proposal-save acknowledgement retains the positive decision")
	run.Proposal = &agentruns.Proposal{Ignore: true}
	require.False(t, runtime.CanPublishFailure(run, failure))
}

func TestChatFollowupRequiresMayaInTheSameExistingThread(t *testing.T) {
	maya, root, conversation := uuid.New(), uuid.New(), uuid.New()
	source := chatdomain.Message{ID: uuid.New(), ConversationID: conversation, ThreadRootID: &root, Sequence: 5, Text: "Can you explain the next step?"}
	reply := chatdomain.Message{ID: uuid.New(), ConversationID: conversation, ThreadRootID: &root, Sequence: 2, Author: chatdomain.Person{ID: maya}, Text: "I can help."}
	require.True(t, chatFollowupCandidate([]chatdomain.Message{reply}, source, maya))
	rootMessage := reply
	rootMessage.ID, rootMessage.ThreadRootID = root, nil
	require.True(t, chatFollowupCandidate([]chatdomain.Message{rootMessage}, source, maya), "Maya can start the thread")
	require.False(t, chatFollowupCandidate(nil, source, maya))
	for _, scenario := range []string{"another thread", "another conversation", "another author", "deleted", "future"} {
		t.Run(scenario, func(t *testing.T) {
			unrelated := reply
			switch scenario {
			case "another thread":
				other := uuid.New()
				unrelated.ThreadRootID = &other
			case "another conversation":
				unrelated.ConversationID = uuid.New()
			case "another author":
				unrelated.Author.ID = uuid.New()
			case "deleted":
				now := time.Now()
				unrelated.DeletedAt = &now
			case "future":
				unrelated.Sequence = source.Sequence + 1
			}
			require.False(t, chatFollowupCandidate([]chatdomain.Message{unrelated}, source, maya))
		})
	}
	source.ThreadRootID = nil
	require.False(t, chatFollowupCandidate([]chatdomain.Message{reply}, source, maya), "Maya participation must not subscribe her to the whole channel")
}

func TestChatFollowupSeparatesDirectHumanMentionsFromAssignmentReferences(t *testing.T) {
	other, maya := uuid.New(), uuid.New()
	mention := chatContentNode{Type: "mention", Attrs: map[string]any{"id": other.String(), "label": "Alex"}}
	for _, scenario := range []struct {
		name     string
		content  []chatContentNode
		directed bool
	}{
		{"direct request", []chatContentNode{mention, {Type: "text", Text: " can you take this?"}}, true},
		{"assignment reference", []chatContentNode{{Type: "text", Text: "Can you assign this to "}, mention}, false},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			rich, err := json.Marshal(chatContentNode{Type: "doc", Content: []chatContentNode{{Type: "paragraph", Content: scenario.content}}})
			require.NoError(t, err)
			source := chatdomain.Message{RichContent: rich, MentionedUserIDs: []uuid.UUID{other}}
			require.Equal(t, scenario.directed, chatAddressesAnotherPerson(source, maya))
		})
	}
}
