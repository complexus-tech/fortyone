package emailagent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	maxMessageRunes         = 32_000
	maxSubjectRunes         = 160
	maxFacts                = 100
	maxTargets              = 100
	maxChoices              = 300
	maxPendingProposals     = 10
	maxFactRunes            = 8_000
	maxDisplayNameRunes     = 500
	maxCurrentStateRunes    = 4_000
	maxProposalSummaryRunes = 2_400
)

var (
	ErrInvalidRequest  = errors.New("invalid email agent request")
	ErrInvalidDecision = errors.New("invalid email agent decision")

	referencePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{0,63}$`)
)

// Service turns one email reply into a safe model, control, or fallback
// decision. It has no repository or domain-service dependency.
type Service struct {
	generator     Generator
	historyLimits HistoryLimits
}

// Option configures Service.
type Option func(*Service) error

// New constructs an email-agent decision service. A nil generator is valid so
// explicit CONFIRM/CANCEL replies continue to work during provider outages;
// all other replies receive a deterministic no-mutation fallback.
func New(generator Generator, options ...Option) (*Service, error) {
	service := &Service{
		generator:     generator,
		historyLimits: DefaultHistoryLimits(),
	}
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option(service); err != nil {
			return nil, err
		}
	}
	return service, nil
}

// WithHistoryLimits overrides the model-context bounds.
func WithHistoryLimits(limits HistoryLimits) Option {
	return func(service *Service) error {
		if err := validateHistoryLimits(limits); err != nil {
			return err
		}
		service.historyLimits = limits
		return nil
	}
}

// Decide gives deterministic control commands precedence over model output,
// then validates and resolves any model proposal against caller-supplied
// authorized targets. Provider failures and malformed output degrade to a safe
// clarification that explicitly promises no mutation.
func (service *Service) Decide(ctx context.Context, request Request) (Decision, error) {
	if service == nil {
		return Decision{}, fmt.Errorf("%w: service is nil", ErrInvalidRequest)
	}
	validated, err := validateAndBuildModelRequest(request, service.historyLimits)
	if err != nil {
		return Decision{}, err
	}

	if control, ok := ParseControlCommand(request.Message); ok {
		return controlDecision(request.Subject, request.PendingProposals, control), nil
	}
	if len(request.PendingProposals) > 1 {
		return ambiguousPendingDecision(request.Subject), nil
	}
	if service.generator == nil || generatorDisabled(service.generator) {
		return fallbackDecision(request.Subject, FallbackGeneratorUnavailable, nil), nil
	}

	generation, err := service.generator.Generate(ctx, validated.modelRequest)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return Decision{}, err
		}
		if errors.Is(err, ErrGeneratorNotConfigured) {
			return fallbackDecision(request.Subject, FallbackGeneratorUnavailable, err), nil
		}
		return fallbackDecision(request.Subject, FallbackGeneratorFailed, err), nil
	}
	if err := validateUsage(generation.Usage); err != nil {
		return fallbackDecision(request.Subject, FallbackInvalidOutput, err), nil
	}

	decision, err := resolveModelDecision(request, validated, generation.Decision)
	if err != nil {
		fallback := fallbackDecision(request.Subject, FallbackInvalidOutput, err)
		fallback.Usage = generation.Usage
		return fallback, nil
	}
	decision.Usage = generation.Usage
	return decision, nil
}

type validatedRequest struct {
	modelRequest   ModelRequest
	references     map[string]struct{}
	grounding      map[string]groundingSource
	targets        map[string]AuthorizedTarget
	choices        map[string]AuthorizedChoice
	checkInSources []string
}

func validateAndBuildModelRequest(request Request, limits HistoryLimits) (validatedRequest, error) {
	if request.WorkspaceID == uuid.Nil {
		return validatedRequest{}, invalidRequest("workspace id is required")
	}
	if request.ActorID == uuid.Nil {
		return validatedRequest{}, invalidRequest("actor id is required")
	}
	message := strings.TrimSpace(request.Message)
	if message == "" {
		return validatedRequest{}, invalidRequest("message is required")
	}
	if utf8.RuneCountInString(message) > maxMessageRunes {
		return validatedRequest{}, invalidRequest("message exceeds %d runes", maxMessageRunes)
	}
	if err := validateHistoryLimits(limits); err != nil {
		return validatedRequest{}, err
	}

	allowedTeams := make(map[uuid.UUID]struct{}, len(request.AllowedTeamIDs))
	for _, teamID := range request.AllowedTeamIDs {
		if teamID == uuid.Nil {
			return validatedRequest{}, invalidRequest("allowed team id cannot be nil")
		}
		allowedTeams[teamID] = struct{}{}
	}

	references := make(map[string]struct{}, len(request.Facts)+len(request.Targets)+len(request.Choices))
	grounding := make(map[string]groundingSource, len(request.Facts)+len(request.Targets)+len(request.Choices))
	checkInSources := make([]string, 0, 2+len(request.Facts)+len(request.History))
	checkInSources = append(checkInSources, message, request.Summary)
	for _, turn := range request.History {
		checkInSources = append(checkInSources, turn.Text)
	}
	if len(request.Facts) > maxFacts {
		return validatedRequest{}, invalidRequest("facts exceed %d", maxFacts)
	}
	facts := make([]GroundedFact, 0, len(request.Facts))
	for index, fact := range request.Facts {
		fact.Reference = strings.TrimSpace(fact.Reference)
		fact.Text = strings.TrimSpace(fact.Text)
		if err := addReference(references, fact.Reference); err != nil {
			return validatedRequest{}, invalidRequest("fact %d: %v", index, err)
		}
		if fact.Text == "" {
			return validatedRequest{}, invalidRequest("fact %q has no text", fact.Reference)
		}
		if utf8.RuneCountInString(fact.Text) > maxFactRunes {
			return validatedRequest{}, invalidRequest("fact %q exceeds %d runes", fact.Reference, maxFactRunes)
		}
		protectedTokens, err := normalizeProtectedTokens(fact.Text, fact.ProtectedTokens)
		if err != nil {
			return validatedRequest{}, invalidRequest("fact %q: %v", fact.Reference, err)
		}
		fact.ProtectedTokens = protectedTokens
		facts = append(facts, fact)
		grounding[fact.Reference] = groundingSource{Text: fact.Text, ProtectedTokens: protectedTokens}
		checkInSources = append(checkInSources, fact.Text)
	}

	if len(request.Targets) > maxTargets {
		return validatedRequest{}, invalidRequest("targets exceed %d", maxTargets)
	}
	targets := make(map[string]AuthorizedTarget, len(request.Targets))
	modelTargets := make([]ModelTarget, 0, len(request.Targets))
	for index, target := range request.Targets {
		target.Reference = strings.TrimSpace(target.Reference)
		target.DisplayName = strings.TrimSpace(target.DisplayName)
		target.CurrentState = strings.TrimSpace(target.CurrentState)
		if err := addReference(references, target.Reference); err != nil {
			return validatedRequest{}, invalidRequest("target %d: %v", index, err)
		}
		if !validTargetKind(target.Kind) {
			return validatedRequest{}, invalidRequest("target %q has unsupported kind %q", target.Reference, target.Kind)
		}
		if target.ID == uuid.Nil || target.TeamID == uuid.Nil {
			return validatedRequest{}, invalidRequest("target %q requires entity and team ids", target.Reference)
		}
		if _, ok := allowedTeams[target.TeamID]; !ok {
			return validatedRequest{}, invalidRequest("target %q is outside the allowed team scope", target.Reference)
		}
		if target.ExpectedUpdatedAt.IsZero() {
			return validatedRequest{}, invalidRequest("target %q requires an expected update time", target.Reference)
		}
		if target.DisplayName == "" || utf8.RuneCountInString(target.DisplayName) > maxDisplayNameRunes {
			return validatedRequest{}, invalidRequest("target %q has an invalid display name", target.Reference)
		}
		if utf8.RuneCountInString(target.CurrentState) > maxCurrentStateRunes {
			return validatedRequest{}, invalidRequest("target %q current state exceeds %d runes", target.Reference, maxCurrentStateRunes)
		}
		target.ExpectedUpdatedAt = target.ExpectedUpdatedAt.UTC()
		targets[target.Reference] = target
		grounding[target.Reference] = groundingSource{
			Text:            strings.TrimSpace(target.DisplayName + " " + target.CurrentState),
			ProtectedTokens: targetProtectedTokens(target),
		}
		modelTargets = append(modelTargets, ModelTarget{
			Reference:    target.Reference,
			Kind:         target.Kind,
			DisplayName:  target.DisplayName,
			CurrentState: target.CurrentState,
		})
	}

	if len(request.Choices) > maxChoices {
		return validatedRequest{}, invalidRequest("choices exceed %d", maxChoices)
	}
	choices := make(map[string]AuthorizedChoice, len(request.Choices))
	modelChoices := make([]ModelChoice, 0, len(request.Choices))
	for index, choice := range request.Choices {
		choice.Reference = strings.TrimSpace(choice.Reference)
		choice.DisplayName = strings.TrimSpace(choice.DisplayName)
		if err := addReference(references, choice.Reference); err != nil {
			return validatedRequest{}, invalidRequest("choice %d: %v", index, err)
		}
		if !validChoiceKind(choice.Kind) {
			return validatedRequest{}, invalidRequest("choice %q has unsupported kind %q", choice.Reference, choice.Kind)
		}
		if choice.ID == uuid.Nil || choice.TeamID == uuid.Nil {
			return validatedRequest{}, invalidRequest("choice %q requires value and team ids", choice.Reference)
		}
		if _, ok := allowedTeams[choice.TeamID]; !ok {
			return validatedRequest{}, invalidRequest("choice %q is outside the allowed team scope", choice.Reference)
		}
		if choice.DisplayName == "" || utf8.RuneCountInString(choice.DisplayName) > maxDisplayNameRunes {
			return validatedRequest{}, invalidRequest("choice %q has an invalid display name", choice.Reference)
		}
		choices[choice.Reference] = choice
		grounding[choice.Reference] = groundingSource{
			Text:            choice.DisplayName,
			ProtectedTokens: []string{choice.DisplayName},
		}
		modelChoices = append(modelChoices, ModelChoice{
			Reference:   choice.Reference,
			Kind:        choice.Kind,
			DisplayName: choice.DisplayName,
		})
	}

	if len(request.PendingProposals) > maxPendingProposals {
		return validatedRequest{}, invalidRequest("pending proposals exceed %d", maxPendingProposals)
	}
	pendingProposals := append([]PendingProposal(nil), request.PendingProposals...)
	seenPending := make(map[uuid.UUID]struct{}, len(request.PendingProposals))
	for index := range pendingProposals {
		pending := &pendingProposals[index]
		pending.Summary = strings.TrimSpace(pending.Summary)
		if pending.ID == uuid.Nil {
			return validatedRequest{}, invalidRequest("pending proposal %d has no id", index)
		}
		if _, exists := seenPending[pending.ID]; exists {
			return validatedRequest{}, invalidRequest("pending proposal id is duplicated")
		}
		seenPending[pending.ID] = struct{}{}
		if pending.Summary == "" || utf8.RuneCountInString(pending.Summary) > maxProposalSummaryRunes {
			return validatedRequest{}, invalidRequest("pending proposal %d has an invalid summary", index)
		}
	}

	var pendingPreview *PendingProposalPreview
	if len(pendingProposals) == 1 {
		pendingPreview = &PendingProposalPreview{Summary: pendingProposals[0].Summary}
	}

	summary, summaryTruncated := BoundSummary(request.Summary, limits.MaxSummaryRunes)
	modelSubject, _ := truncateRunes(sanitizeSubject(request.Subject), maxSubjectRunes)
	return validatedRequest{
		modelRequest: ModelRequest{
			SafetyIdentifier: stableSafetyIdentifier(request),
			Subject:          modelSubject,
			Message:          message,
			Summary:          summary,
			SummaryTruncated: summaryTruncated,
			History:          BoundHistory(request.History, limits),
			Facts:            facts,
			Targets:          modelTargets,
			Choices:          modelChoices,
			PendingProposal:  pendingPreview,
		},
		references:     references,
		grounding:      grounding,
		targets:        targets,
		choices:        choices,
		checkInSources: checkInSources,
	}, nil
}

func resolveModelDecision(request Request, validated validatedRequest, modelDecision ModelDecision) (Decision, error) {
	if modelDecision.Intent == IntentConfirm || modelDecision.Intent == IntentCancel {
		return Decision{}, fmt.Errorf("%w: confirm and cancel must use exact deterministic commands", ErrInvalidDecision)
	}
	if !validModelIntent(modelDecision.Intent) {
		return Decision{}, fmt.Errorf("%w: unsupported intent %q", ErrInvalidDecision, modelDecision.Intent)
	}

	copy, err := resolveEmailCopy(modelDecision.Copy, validated.references)
	if err != nil {
		return Decision{}, err
	}
	decision := Decision{
		Intent: modelDecision.Intent,
		Source: DecisionSourceModel,
		Copy:   &copy,
	}
	if modelDecision.Intent == IntentPropose {
		proposal, err := resolveActionProposal(request, validated, modelDecision.Proposal)
		if err != nil {
			return Decision{}, err
		}
		if err := validateProposalCopy(copy, proposal); err != nil {
			return Decision{}, err
		}
		if err := validateEmailCopyGrounding(copy, modelDecision.Copy, proposalGrounding(validated.grounding, proposal)); err != nil {
			return Decision{}, err
		}
		decision.Proposal = &proposal
	} else if modelDecision.Proposal != nil {
		return Decision{}, fmt.Errorf("%w: intent %q cannot include a proposal", ErrInvalidDecision, modelDecision.Intent)
	} else if err := validateEmailCopyGrounding(copy, modelDecision.Copy, validated.grounding); err != nil {
		return Decision{}, err
	}
	if err := decision.Validate(); err != nil {
		return Decision{}, err
	}
	return decision, nil
}

// Validate checks the trusted decision shape before a caller persists or
// dispatches it.
func (decision Decision) Validate() error {
	switch decision.Intent {
	case IntentAnswer, IntentClarify, IntentRefuse:
		if decision.Copy == nil || decision.Proposal != nil || decision.Command != nil {
			return fmt.Errorf("%w: intent %q requires copy only", ErrInvalidDecision, decision.Intent)
		}
	case IntentPropose:
		if decision.Copy == nil || decision.Proposal == nil || decision.Command != nil {
			return fmt.Errorf("%w: propose requires copy and one proposal", ErrInvalidDecision)
		}
	case IntentConfirm, IntentCancel:
		if decision.Copy != nil || decision.Proposal != nil || decision.Command == nil {
			return fmt.Errorf("%w: control intent requires one command only", ErrInvalidDecision)
		}
		if decision.Command.ProposalID() == uuid.Nil {
			return fmt.Errorf("%w: control command has no pending proposal", ErrInvalidDecision)
		}
		if decision.Intent == IntentConfirm && decision.Command.Kind != ControlConfirm {
			return fmt.Errorf("%w: confirm intent has mismatched command", ErrInvalidDecision)
		}
		if decision.Intent == IntentCancel && decision.Command.Kind != ControlCancel {
			return fmt.Errorf("%w: cancel intent has mismatched command", ErrInvalidDecision)
		}
	default:
		return fmt.Errorf("%w: unsupported intent %q", ErrInvalidDecision, decision.Intent)
	}
	return nil
}

func controlDecision(subject string, pending []PendingProposal, control ControlKind) Decision {
	if len(pending) == 0 {
		return noPendingDecision(subject)
	}
	if len(pending) > 1 {
		return ambiguousPendingDecision(subject)
	}
	intent := IntentConfirm
	if control == ControlCancel {
		intent = IntentCancel
	}
	return Decision{
		Intent: intent,
		Source: DecisionSourceControl,
		Command: &ControlCommand{
			Kind:       control,
			proposalID: pending[0].ID,
		},
	}
}

func noPendingDecision(subject string) Decision {
	return deterministicCopyDecision(
		subject,
		"There isn't a pending change in this conversation, so I haven't updated anything.",
		"Tell me what you'd like to change, and I'll show you exactly what will happen before anything is applied.",
	)
}

func ambiguousPendingDecision(subject string) Decision {
	return deterministicCopyDecision(
		subject,
		"I found more than one pending change in this conversation, so I haven't applied anything.",
		"Tell me which change you want to continue with, and I'll bring back one clear proposal to confirm.",
	)
}

func fallbackDecision(subject string, reason FallbackReason, cause error) Decision {
	decision := deterministicCopyDecision(
		subject,
		"I couldn't safely interpret that update, so I haven't changed anything.",
		"Reply with the objective, key result, task, or feedback item you want to update and the new value. I'll show you the proposed change before anything is applied.",
	)
	decision.Source = DecisionSourceFallback
	decision.FallbackReason = reason
	decision.fallbackCause = cause
	return decision
}

func deterministicCopyDecision(subject string, paragraphs ...string) Decision {
	blocks := make([]CopyBlock, 0, len(paragraphs))
	for _, paragraph := range paragraphs {
		blocks = append(blocks, CopyBlock{Kind: CopyBlockParagraph, Text: paragraph})
	}
	copy := EmailCopy{
		Subject:   replySubject(subject),
		PlainText: RenderPlainText(blocks),
		Blocks:    blocks,
	}
	return Decision{Intent: IntentClarify, Source: DecisionSourceControl, Copy: &copy}
}

func addReference(references map[string]struct{}, reference string) error {
	if !referencePattern.MatchString(reference) {
		return fmt.Errorf("reference %q must match %s", reference, referencePattern.String())
	}
	if _, exists := references[reference]; exists {
		return fmt.Errorf("reference %q is duplicated", reference)
	}
	references[reference] = struct{}{}
	return nil
}

func validTargetKind(kind TargetKind) bool {
	switch kind {
	case TargetObjective, TargetKeyResult, TargetStory, TargetFeedback:
		return true
	default:
		return false
	}
}

func validChoiceKind(kind ChoiceKind) bool {
	return kind == ChoiceStoryStatus || kind == ChoiceStoryAssignee
}

func validModelIntent(intent Intent) bool {
	switch intent {
	case IntentAnswer, IntentClarify, IntentPropose, IntentRefuse:
		return true
	default:
		return false
	}
}

func invalidRequest(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidRequest, fmt.Sprintf(format, args...))
}

func stableSafetyIdentifier(request Request) string {
	identity := strings.TrimSpace(request.SafetyIdentifier)
	if identity == "" {
		identity = request.ActorID.String()
	}
	return safetyIdentifierForIdentity(identity)
}

func safetyIdentifierForIdentity(identity string) string {
	hash := sha256.New()
	_, _ = hash.Write([]byte("email-agent:"))
	_, _ = hash.Write([]byte(strings.TrimSpace(identity)))
	return "maya_email_" + hex.EncodeToString(hash.Sum(nil))[:32]
}

func validateUsage(usage Usage) error {
	if usage.InputTokens < 0 || usage.OutputTokens < 0 || usage.TotalTokens < 0 {
		return fmt.Errorf("%w: token usage cannot be negative", ErrInvalidDecision)
	}
	if usage.TotalTokens != 0 && usage.TotalTokens != usage.InputTokens+usage.OutputTokens {
		return fmt.Errorf("%w: total token usage is inconsistent", ErrInvalidDecision)
	}
	return nil
}

type enabledGenerator interface {
	Enabled() bool
}

func generatorDisabled(generator Generator) bool {
	enabled, ok := generator.(enabledGenerator)
	return ok && !enabled.Enabled()
}
