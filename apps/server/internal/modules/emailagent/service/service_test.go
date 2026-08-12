package emailagent

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type generatorFunc func(context.Context, ModelRequest) (Generation, error)

func (fn generatorFunc) Generate(ctx context.Context, request ModelRequest) (Generation, error) {
	return fn(ctx, request)
}

func TestParseControlCommandIsExactAndReferenceFree(t *testing.T) {
	t.Parallel()

	tests := []struct {
		message string
		kind    ControlKind
		ok      bool
	}{
		{message: "CONFIRM", kind: ControlConfirm, ok: true},
		{message: "  confirm\r\n", kind: ControlConfirm, ok: true},
		{message: "Cancel", kind: ControlCancel, ok: true},
		{message: "yes", ok: false},
		{message: "CONFIRM please", ok: false},
		{message: "CONFIRM opaque-ref", ok: false},
		{message: "CONFIRM\n\nOn Tuesday, Maya wrote:\nCANCEL", ok: false},
		{message: "-- \nCONFIRM", ok: false},
	}
	for _, test := range tests {
		test := test
		t.Run(strings.ReplaceAll(test.message, "\n", "_"), func(t *testing.T) {
			t.Parallel()
			kind, ok := ParseControlCommand(test.message)
			require.Equal(t, test.ok, ok)
			require.Equal(t, test.kind, kind)
		})
	}
}

func TestDecideConfirmBypassesGeneratorAndHidesInternalProposalID(t *testing.T) {
	t.Parallel()

	called := false
	service, err := New(generatorFunc(func(context.Context, ModelRequest) (Generation, error) {
		called = true
		return Generation{}, errors.New("must not be called")
	}))
	require.NoError(t, err)
	request := baseRequest()
	request.Message = "  confirm \n"
	proposalID := uuid.New()
	request.PendingProposals = []PendingProposal{{ID: proposalID, Summary: "Move WEB-42 to Done"}}

	decision, err := service.Decide(context.Background(), request)

	require.NoError(t, err)
	require.False(t, called)
	require.Equal(t, IntentConfirm, decision.Intent)
	require.Equal(t, DecisionSourceControl, decision.Source)
	require.NotNil(t, decision.Command)
	require.Equal(t, ControlConfirm, decision.Command.Kind)
	require.Equal(t, proposalID, decision.Command.ProposalID())
	require.NoError(t, decision.Validate())
	encoded, err := json.Marshal(decision)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), proposalID.String())
	require.NotContains(t, string(encoded), "opaque")
}

func TestDecideControlRequiresExactlyOnePendingProposal(t *testing.T) {
	t.Parallel()

	service, err := New(nil)
	require.NoError(t, err)

	noPending := baseRequest()
	noPending.Message = "CANCEL"
	decision, err := service.Decide(context.Background(), noPending)
	require.NoError(t, err)
	require.Equal(t, IntentClarify, decision.Intent)
	require.Nil(t, decision.Command)
	require.Contains(t, decision.Copy.PlainText, "isn't a pending change")

	ambiguous := baseRequest()
	ambiguous.Message = "CONFIRM"
	ambiguous.PendingProposals = []PendingProposal{
		{ID: uuid.New(), Summary: "First change"},
		{ID: uuid.New(), Summary: "Second change"},
	}
	decision, err = service.Decide(context.Background(), ambiguous)
	require.NoError(t, err)
	require.Equal(t, IntentClarify, decision.Intent)
	require.Nil(t, decision.Command)
	require.Contains(t, decision.Copy.PlainText, "more than one pending change")
}

func TestDecideAllowsAConfirmedPreviewToBeReplacedByACorrectedProposal(t *testing.T) {
	t.Parallel()

	request := baseRequest()
	targetID := uuid.New()
	request.Message = "Actually, set it to On Track instead."
	request.Targets = []AuthorizedTarget{{
		Reference:         "objective_one",
		Kind:              TargetObjective,
		DisplayName:       "Increase activation",
		CurrentState:      "Health: At Risk",
		ID:                targetID,
		TeamID:            request.AllowedTeamIDs[0],
		ExpectedUpdatedAt: time.Now(),
	}}
	request.PendingProposals = []PendingProposal{{
		ID:      uuid.New(),
		Summary: `Set "Increase activation" health to Off Track`,
	}}
	health := ObjectiveHealthOnTrack
	service, err := New(generatorFunc(func(_ context.Context, modelRequest ModelRequest) (Generation, error) {
		require.NotNil(t, modelRequest.PendingProposal)
		return Generation{Decision: proposeDecision(&ModelActionProposal{
			Kind:    ActionObjectiveUpdate,
			Summary: `Set "Increase activation" health to On Track`,
			Objective: &ModelObjectiveAction{
				TargetReference: "objective_one",
				Health:          &health,
			},
		}, "objective_one")}, nil
	}))
	require.NoError(t, err)

	decision, err := service.Decide(context.Background(), request)

	require.NoError(t, err)
	require.Equal(t, IntentPropose, decision.Intent)
	require.NotNil(t, decision.Proposal)
	require.Equal(t, ObjectiveHealthOnTrack, *decision.Proposal.Objective.Health)
}

func TestDecideBoundsContextHashesSafetyIdentityAndWithholdsIDs(t *testing.T) {
	t.Parallel()

	request := baseRequest()
	request.SafetyIdentifier = "joseph@example.com"
	request.Summary = "a durable summary that is deliberately long"
	request.History = []HistoryTurn{
		{Role: RoleUser, Text: "oldest message"},
		{Role: RoleAssistant, Text: "middle message"},
		{Role: RoleUser, Text: "newest message"},
	}
	targetID := uuid.New()
	request.Targets = []AuthorizedTarget{{
		Reference:         "objective_one",
		Kind:              TargetObjective,
		DisplayName:       "Increase activation",
		CurrentState:      "At Risk",
		ID:                targetID,
		TeamID:            request.AllowedTeamIDs[0],
		ExpectedUpdatedAt: time.Now(),
	}}

	var captured ModelRequest
	service, err := New(generatorFunc(func(_ context.Context, modelRequest ModelRequest) (Generation, error) {
		captured = modelRequest
		return Generation{Decision: answerDecision("An objective update", "I can help with that.", "objective_one")}, nil
	}), WithHistoryLimits(HistoryLimits{MaxTurns: 2, MaxTotalRunes: 40, MaxTurnRunes: 20, MaxSummaryRunes: 10}))
	require.NoError(t, err)

	decision, err := service.Decide(context.Background(), request)

	require.NoError(t, err)
	require.Equal(t, IntentAnswer, decision.Intent)
	require.Equal(t, DecisionSourceModel, decision.Source)
	require.Equal(t, "a durable…", captured.Summary)
	require.True(t, captured.SummaryTruncated)
	require.Len(t, captured.History.Turns, 2)
	require.Equal(t, "middle message", captured.History.Turns[0].Text)
	require.Equal(t, "newest message", captured.History.Turns[1].Text)
	require.Equal(t, 1, captured.History.OmittedTurns)
	require.True(t, captured.History.Truncated)
	require.NotEmpty(t, captured.SafetyIdentifier)
	require.NotContains(t, captured.SafetyIdentifier, request.SafetyIdentifier)

	modelJSON, err := json.Marshal(captured)
	require.NoError(t, err)
	require.NotContains(t, string(modelJSON), targetID.String())
	require.NotContains(t, string(modelJSON), request.WorkspaceID.String())
	require.NotContains(t, string(modelJSON), request.ActorID.String())
	require.NotContains(t, string(modelJSON), request.SafetyIdentifier)

	second := request
	second.ActorID = uuid.New()
	var secondSafetyID string
	secondService, err := New(generatorFunc(func(_ context.Context, modelRequest ModelRequest) (Generation, error) {
		secondSafetyID = modelRequest.SafetyIdentifier
		return Generation{Decision: answerDecision("An objective update", "I can help with that.", "objective_one")}, nil
	}), WithHistoryLimits(HistoryLimits{MaxTurns: 2, MaxTotalRunes: 40, MaxTurnRunes: 20, MaxSummaryRunes: 10}))
	require.NoError(t, err)
	_, err = secondService.Decide(context.Background(), second)
	require.NoError(t, err)
	require.Equal(t, captured.SafetyIdentifier, secondSafetyID, "explicit safety identity must be stable across requests")
}

func TestDecideResolvesEverySupportedProposalFromAuthorizedReferences(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		kind     TargetKind
		choices  func(teamID uuid.UUID) []AuthorizedChoice
		proposal func() *ModelActionProposal
		assert   func(*testing.T, ActionProposal, uuid.UUID, uuid.UUID)
	}{
		{
			name: "objective health and check-in",
			kind: TargetObjective,
			proposal: func() *ModelActionProposal {
				health := ObjectiveHealthOnTrack
				checkIn := "The launch risk is resolved."
				return &ModelActionProposal{
					Kind:    ActionObjectiveUpdate,
					Summary: "Increase activation: set health to On Track. Check-in: The launch risk is resolved.",
					Objective: &ModelObjectiveAction{
						TargetReference: "target_one",
						Health:          &health,
						CheckIn:         &checkIn,
					},
				}
			},
			assert: func(t *testing.T, proposal ActionProposal, targetID, _ uuid.UUID) {
				require.Equal(t, targetID, proposal.Objective.Target.ID)
				require.Equal(t, ObjectiveHealthOnTrack, *proposal.Objective.Health)
				require.Equal(t, "The launch risk is resolved.", *proposal.Objective.CheckIn)
				require.Equal(t, `Set "Increase activation" health to On Track`, proposal.Summary)
			},
		},
		{
			name: "key result current value and check-in",
			kind: TargetKeyResult,
			proposal: func() *ModelActionProposal {
				value := 72.5
				checkIn := "Instrumentation is now complete."
				return &ModelActionProposal{
					Kind:    ActionKeyResultUpdate,
					Summary: "Increase activation: set current value to 72.5. Check-in: Instrumentation is now complete.",
					KeyResult: &ModelKeyResultAction{
						TargetReference: "target_one",
						CurrentValue:    &value,
						CheckIn:         &checkIn,
					},
				}
			},
			assert: func(t *testing.T, proposal ActionProposal, targetID, _ uuid.UUID) {
				require.Equal(t, targetID, proposal.KeyResult.Target.ID)
				require.Equal(t, 72.5, *proposal.KeyResult.CurrentValue)
			},
		},
		{
			name: "story date status and assignee",
			kind: TargetStory,
			choices: func(teamID uuid.UUID) []AuthorizedChoice {
				return []AuthorizedChoice{
					{Reference: "status_done", Kind: ChoiceStoryStatus, DisplayName: "Done", ID: uuid.New(), TeamID: teamID},
					{Reference: "person_maya", Kind: ChoiceStoryAssignee, DisplayName: "Maya Chen", ID: uuid.New(), TeamID: teamID},
				}
			},
			proposal: func() *ModelActionProposal {
				return &ModelActionProposal{
					Kind:    ActionStoryUpdate,
					Summary: "Increase activation: set due date to 2026-08-20, status to Done, and assignee to Maya Chen",
					Story: &ModelStoryAction{
						TargetReference: "target_one",
						DueDate:         &ModelDateChange{Operation: DateSet, Date: "2026-08-20"},
						Status:          &ModelStatusChange{ChoiceReference: "status_done"},
						Assignee:        &ModelAssigneeChange{Operation: AssigneeAssign, ChoiceReference: "person_maya"},
					},
				}
			},
			assert: func(t *testing.T, proposal ActionProposal, targetID, _ uuid.UUID) {
				require.Equal(t, targetID, proposal.Story.Target.ID)
				require.Equal(t, "2026-08-20", proposal.Story.DueDate.Date)
				require.Equal(t, "Done", proposal.Story.Status.StatusName)
				require.Equal(t, "Maya Chen", proposal.Story.Assignee.AssigneeName)
			},
		},
		{
			name: "feedback status single target",
			kind: TargetFeedback,
			proposal: func() *ModelActionProposal {
				return &ModelActionProposal{
					Kind:    ActionFeedbackStatus,
					Summary: "Increase activation: set feedback status to planned",
					Feedback: &ModelFeedbackStatusAction{
						TargetReference: "target_one",
						Status:          FeedbackStatusPlanned,
					},
				}
			},
			assert: func(t *testing.T, proposal ActionProposal, targetID, _ uuid.UUID) {
				require.Equal(t, targetID, proposal.Feedback.Target.ID)
				require.Equal(t, FeedbackStatusPlanned, proposal.Feedback.Status)
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			request := baseRequest()
			targetID := uuid.New()
			expectedVersion := time.Date(2026, 8, 12, 8, 30, 0, 0, time.FixedZone("CAT", 2*60*60))
			request.Targets = []AuthorizedTarget{{
				Reference:         "target_one",
				Kind:              test.kind,
				DisplayName:       "Increase activation",
				CurrentState:      "Needs an update",
				ID:                targetID,
				TeamID:            request.AllowedTeamIDs[0],
				ExpectedUpdatedAt: expectedVersion,
			}}
			if test.choices != nil {
				request.Choices = test.choices(request.AllowedTeamIDs[0])
			}
			proposal := test.proposal()
			if proposal.Objective != nil && proposal.Objective.CheckIn != nil {
				request.Message = *proposal.Objective.CheckIn
			}
			if proposal.KeyResult != nil && proposal.KeyResult.CheckIn != nil {
				request.Message = *proposal.KeyResult.CheckIn
			}
			service, err := New(generatorFunc(func(context.Context, ModelRequest) (Generation, error) {
				return Generation{
					Decision: proposeDecision(proposal, "target_one"),
					Usage:    Usage{InputTokens: 100, OutputTokens: 50, TotalTokens: 150},
				}, nil
			}))
			require.NoError(t, err)

			decision, err := service.Decide(context.Background(), request)

			require.NoError(t, err)
			if decision.Intent != IntentPropose {
				t.Fatalf("unexpected fallback: %v", decision.FallbackCause())
			}
			require.Equal(t, IntentPropose, decision.Intent)
			require.Equal(t, DecisionSourceModel, decision.Source)
			require.NotNil(t, decision.Proposal)
			require.Equal(t, request.WorkspaceID, decision.Proposal.WorkspaceID)
			require.Equal(t, request.ActorID, decision.Proposal.ActorID)
			require.Equal(t, expectedVersion.UTC(), proposalTarget(*decision.Proposal).ExpectedUpdatedAt)
			require.Equal(t, 150, decision.Usage.TotalTokens)
			test.assert(t, *decision.Proposal, targetID, request.AllowedTeamIDs[0])
		})
	}
}

func TestDecideRejectsUnsafeOrSemanticallyInvalidModelOutputWithFallback(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		mutate   func(*Request)
		decision func() ModelDecision
	}{
		{
			name: "model control intent",
			decision: func() ModelDecision {
				decision := answerDecision("Confirmed", "Done.", "target_one")
				decision.Intent = IntentConfirm
				return decision
			},
		},
		{
			name: "objective without health",
			decision: func() ModelDecision {
				return proposeDecision(&ModelActionProposal{
					Kind:      ActionObjectiveUpdate,
					Summary:   "Add a check-in",
					Objective: &ModelObjectiveAction{TargetReference: "target_one", CheckIn: stringPointer("Status unchanged")},
				}, "target_one")
			},
		},
		{
			name: "key result without current value",
			mutate: func(request *Request) {
				request.Targets[0].Kind = TargetKeyResult
			},
			decision: func() ModelDecision {
				return proposeDecision(&ModelActionProposal{
					Kind:      ActionKeyResultUpdate,
					Summary:   "Add a check-in",
					KeyResult: &ModelKeyResultAction{TargetReference: "target_one", CheckIn: stringPointer("Status unchanged")},
				}, "target_one")
			},
		},
		{
			name: "unknown copy reference",
			decision: func() ModelDecision {
				return answerDecision("An update", "The value is ready.", "invented_reference")
			},
		},
		{
			name: "model-authored URL",
			decision: func() ModelDecision {
				return answerDecision("An update", "Open https://evil.example to continue.", "target_one")
			},
		},
		{
			name: "two proposal payloads",
			decision: func() ModelDecision {
				health := ObjectiveHealthAtRisk
				value := 10.0
				return proposeDecision(&ModelActionProposal{
					Kind:      ActionObjectiveUpdate,
					Summary:   "Two changes",
					Objective: &ModelObjectiveAction{TargetReference: "target_one", Health: &health},
					KeyResult: &ModelKeyResultAction{TargetReference: "target_one", CurrentValue: &value},
				}, "target_one")
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			request := requestWithTarget(TargetObjective)
			if test.mutate != nil {
				test.mutate(&request)
			}
			service, err := New(generatorFunc(func(context.Context, ModelRequest) (Generation, error) {
				return Generation{Decision: test.decision()}, nil
			}))
			require.NoError(t, err)

			decision, err := service.Decide(context.Background(), request)

			require.NoError(t, err)
			require.Equal(t, IntentClarify, decision.Intent)
			require.Equal(t, DecisionSourceFallback, decision.Source)
			require.Equal(t, FallbackInvalidOutput, decision.FallbackReason)
			require.ErrorIs(t, decision.FallbackCause(), ErrInvalidDecision)
			require.Nil(t, decision.Proposal)
			require.Contains(t, decision.Copy.PlainText, "haven't changed anything")
		})
	}
}

func TestDecideFallbackAndCancellationBehaviorIsDeterministic(t *testing.T) {
	t.Parallel()

	request := baseRequest()
	service, err := New(nil)
	require.NoError(t, err)
	decision, err := service.Decide(context.Background(), request)
	require.NoError(t, err)
	require.Equal(t, FallbackGeneratorUnavailable, decision.FallbackReason)
	require.Equal(t, DecisionSourceFallback, decision.Source)

	providerError := errors.New("provider unavailable")
	service, err = New(generatorFunc(func(context.Context, ModelRequest) (Generation, error) {
		return Generation{}, providerError
	}))
	require.NoError(t, err)
	decision, err = service.Decide(context.Background(), request)
	require.NoError(t, err)
	require.Equal(t, FallbackGeneratorFailed, decision.FallbackReason)
	require.ErrorIs(t, decision.FallbackCause(), providerError)

	service, err = New(generatorFunc(func(context.Context, ModelRequest) (Generation, error) {
		return Generation{}, context.Canceled
	}))
	require.NoError(t, err)
	_, err = service.Decide(context.Background(), request)
	require.ErrorIs(t, err, context.Canceled)
}

func TestDecideAccountsForUsageWhenModelOutputIsRejected(t *testing.T) {
	t.Parallel()

	request := requestWithTarget(TargetObjective)
	service, err := New(generatorFunc(func(context.Context, ModelRequest) (Generation, error) {
		decision := answerDecision("Unsafe answer", "Open https://evil.example", "target_one")
		return Generation{
			Decision: decision,
			Usage:    Usage{InputTokens: 70, OutputTokens: 30, TotalTokens: 100},
		}, nil
	}))
	require.NoError(t, err)

	decision, err := service.Decide(context.Background(), request)

	require.NoError(t, err)
	require.Equal(t, FallbackInvalidOutput, decision.FallbackReason)
	require.Equal(t, 100, decision.Usage.TotalTokens)
}

func TestRequestValidationRejectsOutOfScopeTargetsBeforeGenerator(t *testing.T) {
	t.Parallel()

	called := false
	service, err := New(generatorFunc(func(context.Context, ModelRequest) (Generation, error) {
		called = true
		return Generation{}, nil
	}))
	require.NoError(t, err)
	request := requestWithTarget(TargetStory)
	request.Targets[0].TeamID = uuid.New()

	_, err = service.Decide(context.Background(), request)

	require.ErrorIs(t, err, ErrInvalidRequest)
	require.False(t, called)
}

func TestStoryProposalRejectsChoiceFromAnotherAllowedTeam(t *testing.T) {
	t.Parallel()

	request := requestWithTarget(TargetStory)
	otherTeamID := uuid.New()
	request.AllowedTeamIDs = append(request.AllowedTeamIDs, otherTeamID)
	request.Choices = []AuthorizedChoice{{
		Reference:   "other_status",
		Kind:        ChoiceStoryStatus,
		DisplayName: "Done",
		ID:          uuid.New(),
		TeamID:      otherTeamID,
	}}
	model := proposeDecision(&ModelActionProposal{
		Kind:    ActionStoryUpdate,
		Summary: "Move the story to Done",
		Story: &ModelStoryAction{
			TargetReference: "target_one",
			Status:          &ModelStatusChange{ChoiceReference: "other_status"},
		},
	}, "target_one")
	service, err := New(generatorFunc(func(context.Context, ModelRequest) (Generation, error) {
		return Generation{Decision: model}, nil
	}))
	require.NoError(t, err)

	decision, err := service.Decide(context.Background(), request)

	require.NoError(t, err)
	require.Equal(t, FallbackInvalidOutput, decision.FallbackReason)
	require.ErrorContains(t, decision.FallbackCause(), "does not belong to the target team")
}

func baseRequest() Request {
	return Request{
		WorkspaceID:    uuid.New(),
		ActorID:        uuid.New(),
		AllowedTeamIDs: []uuid.UUID{uuid.New()},
		Subject:        "Objective health needs attention",
		Message:        "Can you help me update this?",
		Summary:        "The conversation is about a product update.",
	}
}

func requestWithTarget(kind TargetKind) Request {
	request := baseRequest()
	request.Targets = []AuthorizedTarget{{
		Reference:         "target_one",
		Kind:              kind,
		DisplayName:       "Increase activation",
		CurrentState:      "At Risk",
		ID:                uuid.New(),
		TeamID:            request.AllowedTeamIDs[0],
		ExpectedUpdatedAt: time.Now().UTC(),
	}}
	return request
}

func answerDecision(subject, body, reference string) ModelDecision {
	references := []string{}
	if reference != "" {
		references = []string{reference}
	}
	return ModelDecision{
		Intent: IntentAnswer,
		Copy: DraftEmailCopy{
			Subject: GroundedSubject{Text: subject, References: references},
			Blocks: []CopyBlock{{
				Kind:       CopyBlockParagraph,
				Text:       body,
				References: references,
			}},
		},
	}
}

func proposeDecision(proposal *ModelActionProposal, reference string) ModelDecision {
	proposalReferences := []string{reference, proposalReference}
	return ModelDecision{
		Intent: IntentPropose,
		Copy: DraftEmailCopy{
			Subject: GroundedSubject{Text: "A proposed update", References: proposalReferences},
			Blocks: []CopyBlock{
				{Kind: CopyBlockParagraph, Text: proposal.Summary, References: proposalReferences},
				{Kind: CopyBlockCallout, Text: "Reply CONFIRM to apply this change, or CANCEL to leave it unchanged.", References: []string{}},
			},
		},
		Proposal: proposal,
	}
}

func proposalTarget(proposal ActionProposal) TargetSnapshot {
	switch proposal.Kind {
	case ActionObjectiveUpdate:
		return proposal.Objective.Target
	case ActionKeyResultUpdate:
		return proposal.KeyResult.Target
	case ActionStoryUpdate:
		return proposal.Story.Target
	case ActionFeedbackStatus:
		return proposal.Feedback.Target
	default:
		return TargetSnapshot{}
	}
}

func stringPointer(value string) *string {
	return &value
}
