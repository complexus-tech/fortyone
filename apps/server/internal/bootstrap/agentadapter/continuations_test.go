package agentadapter

import (
	"context"
	"encoding/json"
	"testing"

	agentruns "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/service"
	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type parentRunReader struct{ run agentruns.Run }

func (r parentRunReader) Get(context.Context, uuid.UUID, uuid.UUID) (agentruns.Run, error) {
	return r.run, nil
}

type rebindingTools struct {
	runtimeTools
	previous, current messaging.ToolScope
	token             string
	count             int
}

func (r *rebindingTools) RebindStoryMutation(_ context.Context, previous, current messaging.ToolScope, token string) (messaging.StoryMutationConfirmation, error) {
	r.previous, r.current, r.token = previous, current, token
	r.count++
	return messaging.StoryMutationConfirmation{Token: "new-requester-token"}, nil
}

func TestConversationalApprovalUsesAnsweringTeammateAndPersistedProposal(t *testing.T) {
	runtime, run, _ := runtimeFixture()
	parent := run
	parent.ID, parent.RequesterID = uuid.New(), uuid.New()
	parent.Proposal = &agentruns.Proposal{RequiresApproval: true, Action: json.RawMessage(`{"token":"original-exact-proposal"}`)}
	run.ParentRunID = &parent.ID
	tools := &rebindingTools{runtimeTools: runtimeTools{direct: true}}
	runtime.Tools, runtime.Runs = tools, parentRunReader{run: parent}
	proposal, accepted, err := runtime.continueApproval(t.Context(), run, "<p>Yes, please!</p>")
	require.NoError(t, err)
	require.True(t, accepted)
	require.False(t, proposal.RequiresApproval)
	require.Equal(t, "original-exact-proposal", tools.token)
	require.Equal(t, parent.RequesterID, tools.previous.UserID)
	require.Equal(t, run.RequesterID, tools.current.UserID)
	require.Contains(t, string(proposal.Action), "new-requester-token")
	_, accepted, err = runtime.continueApproval(t.Context(), run, "<p>Yes, but change the deadline too.</p>")
	require.NoError(t, err)
	require.False(t, accepted)
	require.Equal(t, 1, tools.count, "a revised instruction must not confirm the old proposal")
}

func TestConversationalApprovalRequiresUnambiguousUnquotedAnswer(t *testing.T) {
	for _, text := range []string{"<p>yes</p>", "<p>Go ahead.</p>", "<p>@Maya apply it</p>"} {
		require.True(t, acceptsProposal(text), text)
	}
	for _, text := range []string{"<blockquote>yes</blockquote>", "<pre>apply it</pre>", "yes?", "yes but only the title", "do not apply it", "<p>\"yes\"</p>", ""} {
		require.False(t, acceptsProposal(text), text)
	}
}
