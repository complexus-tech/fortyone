package emailagent

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	maxProtectedTokens       = 30
	maxProtectedTokenRunes   = 300
	minCheckInGroundingRunes = 8
)

var groundedNumericTokenPattern = regexp.MustCompile(`(?i)(?:[$€£]?[-+]?\d+(?:[.,]\d+)*(?:%|st|nd|rd|th)?|\bq[1-4]\b)`)

type groundingSource struct {
	Text            string
	ProtectedTokens []string
}

func normalizeProtectedTokens(source string, values []string) ([]string, error) {
	if len(values) > maxProtectedTokens {
		return nil, fmt.Errorf("protected tokens exceed %d", maxProtectedTokens)
	}
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, fmt.Errorf("protected token cannot be empty")
		}
		if utf8.RuneCountInString(value) > maxProtectedTokenRunes {
			return nil, fmt.Errorf("protected token exceeds %d runes", maxProtectedTokenRunes)
		}
		if !strings.Contains(source, value) {
			return nil, fmt.Errorf("fact text does not contain protected token %q", value)
		}
		if _, duplicate := seen[value]; duplicate {
			return nil, fmt.Errorf("protected token %q is duplicated", value)
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result, nil
}

func targetProtectedTokens(target AuthorizedTarget) []string {
	tokens := []string{target.DisplayName}
	if target.CurrentState != "" {
		tokens = append(tokens, target.CurrentState)
	}
	return tokens
}

func proposalGrounding(all map[string]groundingSource, proposal ActionProposal) map[string]groundingSource {
	grounding := cloneGrounding(all)
	proposalText := trustedProposalSummary(proposal)
	protected := make([]string, 0, 8)
	switch proposal.Kind {
	case ActionObjectiveUpdate:
		protected = append(protected, proposal.Objective.Target.DisplayName, string(*proposal.Objective.Health))
		if proposal.Objective.CheckIn != nil {
			protected = append(protected, *proposal.Objective.CheckIn)
		}
	case ActionKeyResultUpdate:
		protected = append(protected, proposal.KeyResult.Target.DisplayName)
		if proposal.KeyResult.CheckIn != nil {
			protected = append(protected, *proposal.KeyResult.CheckIn)
		}
	case ActionStoryUpdate:
		protected = append(protected, proposal.Story.Target.DisplayName)
		if proposal.Story.DueDate != nil && proposal.Story.DueDate.Operation == DateSet {
			protected = append(protected, proposal.Story.DueDate.Date)
		}
		if proposal.Story.Status != nil {
			protected = append(protected, proposal.Story.Status.StatusName)
		}
		if proposal.Story.Assignee != nil && proposal.Story.Assignee.Operation == AssigneeAssign {
			protected = append(protected, proposal.Story.Assignee.AssigneeName)
		}
	case ActionFeedbackStatus:
		protected = append(protected, proposal.Feedback.Target.DisplayName, string(proposal.Feedback.Status))
	}
	grounding[proposalReference] = groundingSource{Text: proposalText, ProtectedTokens: protected}
	return grounding
}

const proposalReference = "proposal"

func cloneGrounding(source map[string]groundingSource) map[string]groundingSource {
	cloned := make(map[string]groundingSource, len(source)+1)
	for reference, value := range source {
		value.ProtectedTokens = append([]string(nil), value.ProtectedTokens...)
		cloned[reference] = value
	}
	return cloned
}

func validateEmailCopyGrounding(copy EmailCopy, draft DraftEmailCopy, grounding map[string]groundingSource) error {
	if err := validateGroundedSurface("subject", copy.Subject, draft.Subject.References, grounding); err != nil {
		return err
	}
	for index, block := range copy.Blocks {
		if err := validateGroundedSurface(fmt.Sprintf("block %d", index), blockSurfaceText(block), block.References, grounding); err != nil {
			return err
		}
	}
	return nil
}

func validateGroundedSurface(location, text string, references []string, grounding map[string]groundingSource) error {
	sourceText := strings.Builder{}
	allowedProtected := make(map[string]struct{})
	for _, reference := range references {
		source, ok := grounding[reference]
		if !ok {
			return fmt.Errorf("%w: %s cites unknown grounding reference %q", ErrInvalidDecision, location, reference)
		}
		sourceText.WriteString(source.Text)
		sourceText.WriteByte(' ')
		for _, token := range source.ProtectedTokens {
			allowedProtected[strings.ToLower(token)] = struct{}{}
		}
	}
	for token := range numericTokens(text) {
		if _, ok := numericTokens(sourceText.String())[token]; !ok {
			return fmt.Errorf("%w: %s invents numeric or date token %q", ErrInvalidDecision, location, token)
		}
	}
	textLower := strings.ToLower(text)
	for _, source := range grounding {
		for _, protected := range source.ProtectedTokens {
			normalized := strings.ToLower(protected)
			if normalized == "" {
				continue
			}
			if _, allowed := allowedProtected[normalized]; allowed {
				continue
			}
			if strings.Contains(textLower, normalized) {
				return fmt.Errorf("%w: %s uses protected token %q without citing it", ErrInvalidDecision, location, protected)
			}
		}
	}
	return nil
}

func validateCheckInGrounding(checkIn *string, sources []string) error {
	if checkIn == nil {
		return nil
	}
	normalized := strings.TrimSpace(*checkIn)
	if utf8.RuneCountInString(normalized) < minCheckInGroundingRunes {
		return fmt.Errorf("%w: check-in is too short to ground safely", ErrInvalidDecision)
	}
	for _, source := range sources {
		if strings.Contains(source, normalized) {
			return nil
		}
	}
	return fmt.Errorf("%w: check-in is not a verbatim user or grounded fact statement", ErrInvalidDecision)
}

func blockSurfaceText(block CopyBlock) string {
	parts := make([]string, 0, 1+len(block.Items))
	if block.Text != "" {
		parts = append(parts, block.Text)
	}
	parts = append(parts, block.Items...)
	return strings.Join(parts, " ")
}

func numericTokens(value string) map[string]struct{} {
	tokens := make(map[string]struct{})
	for _, token := range groundedNumericTokenPattern.FindAllString(strings.ToLower(value), -1) {
		tokens[token] = struct{}{}
	}
	return tokens
}
