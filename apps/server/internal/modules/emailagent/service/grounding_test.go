package emailagent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecideRejectsInventedNumericFactsAndUncitedProtectedTokens(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		body      string
		reference string
	}{
		{name: "invented number", body: "Increase activation is at 20%.", reference: "activation_fact"},
		{name: "uncited protected health", body: "Increase activation is At Risk.", reference: "other_fact"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			request := baseRequest()
			request.Facts = []GroundedFact{
				{Reference: "activation_fact", Text: "Increase activation is at 15% and At Risk.", ProtectedTokens: []string{"Increase activation", "15%", "At Risk"}},
				{Reference: "other_fact", Text: "The review is due this week.", ProtectedTokens: []string{"this week"}},
			}
			service, err := New(generatorFunc(func(context.Context, ModelRequest) (Generation, error) {
				return Generation{Decision: answerDecision("An objective update", test.body, test.reference)}, nil
			}))
			require.NoError(t, err)

			decision, err := service.Decide(context.Background(), request)

			require.NoError(t, err)
			require.Equal(t, FallbackInvalidOutput, decision.FallbackReason)
			require.ErrorIs(t, decision.FallbackCause(), ErrInvalidDecision)
		})
	}
}

func TestDecideRejectsInventedCheckInEvenWhenOtherProposalFieldsAreValid(t *testing.T) {
	t.Parallel()

	request := requestWithTarget(TargetObjective)
	health := ObjectiveHealthOnTrack
	checkIn := "The launch risk has been fully resolved."
	model := proposeDecision(&ModelActionProposal{
		Kind:    ActionObjectiveUpdate,
		Summary: "Increase activation: set health to On Track. Check-in: " + checkIn,
		Objective: &ModelObjectiveAction{
			TargetReference: "target_one",
			Health:          &health,
			CheckIn:         &checkIn,
		},
	}, "target_one")
	service, err := New(generatorFunc(func(context.Context, ModelRequest) (Generation, error) {
		return Generation{Decision: model}, nil
	}))
	require.NoError(t, err)

	decision, err := service.Decide(context.Background(), request)

	require.NoError(t, err)
	require.Equal(t, FallbackInvalidOutput, decision.FallbackReason)
	require.ErrorContains(t, decision.FallbackCause(), "check-in is not a verbatim")
}

func TestRequestRejectsProtectedTokenNotPresentInFact(t *testing.T) {
	t.Parallel()

	request := baseRequest()
	request.Facts = []GroundedFact{{
		Reference:       "fact_one",
		Text:            "The objective is At Risk.",
		ProtectedTokens: []string{"Off Track"},
	}}
	service, err := New(nil)
	require.NoError(t, err)

	_, err = service.Decide(context.Background(), request)

	require.ErrorIs(t, err, ErrInvalidRequest)
	require.ErrorContains(t, err, "does not contain protected token")
}
