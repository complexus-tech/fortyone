package emailcopy

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

var numericTokenPattern = regexp.MustCompile(`(?i)(?:[$€£]?[-+]?\d+(?:[.,]\d+)*(?:%|st|nd|rd|th)?|\bq[1-4]\b)`)

func validateRequest(request Request) error {
	if strings.TrimSpace(request.Purpose) == "" {
		return fmt.Errorf("email copy purpose is required")
	}
	if len(request.Facts) == 0 {
		return fmt.Errorf("at least one email copy fact is required")
	}
	factIDs := make(map[string]struct{}, len(request.Facts))
	for _, fact := range request.Facts {
		if err := validateReferenceID(fact.ReferenceID); err != nil {
			return fmt.Errorf("invalid fact reference: %w", err)
		}
		if _, exists := factIDs[fact.ReferenceID]; exists {
			return fmt.Errorf("duplicate fact reference %q", fact.ReferenceID)
		}
		factIDs[fact.ReferenceID] = struct{}{}
		if strings.TrimSpace(fact.Text) == "" {
			return fmt.Errorf("fact %q has no text", fact.ReferenceID)
		}
		for _, token := range factProtectedTokens(fact) {
			if trimmed := strings.TrimSpace(token); trimmed != "" && !strings.Contains(fact.Text, trimmed) {
				return fmt.Errorf("fact %q does not contain protected token %q", fact.ReferenceID, token)
			}
		}
	}
	actionIDs := make(map[string]struct{}, len(request.Actions))
	for _, action := range request.Actions {
		if err := validateReferenceID(action.ReferenceID); err != nil {
			return fmt.Errorf("invalid action reference: %w", err)
		}
		if _, exists := actionIDs[action.ReferenceID]; exists {
			return fmt.Errorf("duplicate action reference %q", action.ReferenceID)
		}
		actionIDs[action.ReferenceID] = struct{}{}
		if strings.TrimSpace(action.Description) == "" {
			return fmt.Errorf("action %q has no description", action.ReferenceID)
		}
	}
	return nil
}

func validateReferenceID(referenceID string) error {
	trimmed := strings.TrimSpace(referenceID)
	if trimmed == "" || len(trimmed) > 100 {
		return fmt.Errorf("reference ID must contain 1 to 100 characters")
	}
	for _, character := range trimmed {
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || character == '_' || character == '-' || character == ':' {
			continue
		}
		return fmt.Errorf("reference ID %q contains unsupported characters", referenceID)
	}
	return nil
}

func validateOutput(request Request, output Output) error {
	facts := make(map[string]Fact, len(request.Facts))
	for _, fact := range request.Facts {
		facts[fact.ReferenceID] = fact
	}
	actions := make(map[string]Action, len(request.Actions))
	for _, action := range request.Actions {
		actions[action.ReferenceID] = action
	}

	if err := validateGroundedText("subject", output.Subject, 90, facts); err != nil {
		return err
	}
	if err := validateGroundedText("h1", output.H1, 110, facts); err != nil {
		return err
	}
	if err := validateGroundedText("intro", output.Intro, 420, facts); err != nil {
		return err
	}
	if err := validateOptionalGroundedText("sender prose", output.SenderProse, request.IncludeSenderProse, 320, facts); err != nil {
		return err
	}
	if err := validateOptionalGroundedText("feedback theme summary", output.FeedbackThemeSummary, request.IncludeFeedbackThemeSummary, 520, facts); err != nil {
		return err
	}
	if err := validateOptionalGroundedText("reply prompt", output.ReplyPrompt, request.IncludeReplyPrompt, 320, facts); err != nil {
		return err
	}
	if request.IncludeReplyPrompt {
		prompt := strings.ToLower(output.ReplyPrompt.Text)
		for _, requiredPhrase := range []string{"maya", "ai agent", "reply", "email"} {
			if !strings.Contains(prompt, requiredPhrase) {
				return fmt.Errorf("reply prompt must include %q", requiredPhrase)
			}
		}
	}

	seenRows := make(map[string]struct{}, len(output.Rows))
	for _, row := range output.Rows {
		fact, exists := facts[row.ReferenceID]
		if !exists {
			return fmt.Errorf("row cites unknown fact %q", row.ReferenceID)
		}
		if _, duplicate := seenRows[row.ReferenceID]; duplicate {
			return fmt.Errorf("duplicate row for fact %q", row.ReferenceID)
		}
		if !fact.Required {
			return fmt.Errorf("row cites context-only fact %q", row.ReferenceID)
		}
		seenRows[row.ReferenceID] = struct{}{}
		if err := validatePlainText("row", row.Text, 360); err != nil {
			return err
		}
		for _, token := range factProtectedTokens(fact) {
			if token != "" && !strings.Contains(row.Text, token) {
				return fmt.Errorf("row %q omitted protected token %q", row.ReferenceID, token)
			}
		}
		if err := requireNumericTokens(fact.Text, row.Text); err != nil {
			return fmt.Errorf("row %q: %w", row.ReferenceID, err)
		}
		if err := rejectContradictoryProtectedTokens(row.Text, []string{row.ReferenceID}, facts); err != nil {
			return fmt.Errorf("row %q: %w", row.ReferenceID, err)
		}
		if row.CTAReferenceID != "" {
			if _, exists := actions[row.CTAReferenceID]; !exists {
				return fmt.Errorf("row %q cites unknown action %q", row.ReferenceID, row.CTAReferenceID)
			}
		}
	}
	for _, fact := range request.Facts {
		if fact.Required {
			if _, exists := seenRows[fact.ReferenceID]; !exists {
				return fmt.Errorf("required fact %q has no row", fact.ReferenceID)
			}
		}
	}

	seenCTAs := make(map[string]struct{}, len(output.CTAs))
	allGroundingText := strings.Builder{}
	for _, fact := range request.Facts {
		allGroundingText.WriteString(fact.Text)
		allGroundingText.WriteByte(' ')
	}
	for _, cta := range output.CTAs {
		action, exists := actions[cta.ReferenceID]
		if !exists {
			return fmt.Errorf("CTA cites unknown action %q", cta.ReferenceID)
		}
		if _, duplicate := seenCTAs[cta.ReferenceID]; duplicate {
			return fmt.Errorf("duplicate CTA for action %q", cta.ReferenceID)
		}
		seenCTAs[cta.ReferenceID] = struct{}{}
		if err := validatePlainText("CTA label", cta.Label, 48); err != nil {
			return err
		}
		if err := rejectInventedNumericTokens(allGroundingText.String()+" "+action.Description, cta.Label); err != nil {
			return fmt.Errorf("CTA %q: %w", cta.ReferenceID, err)
		}
	}
	for referenceID := range actions {
		if _, exists := seenCTAs[referenceID]; !exists {
			return fmt.Errorf("action %q has no CTA", referenceID)
		}
	}
	return nil
}

func factProtectedTokens(fact Fact) []string {
	tokens := make([]string, 0, len(fact.EntityTokens)+len(fact.ProtectedTokens))
	tokens = append(tokens, fact.EntityTokens...)
	tokens = append(tokens, fact.ProtectedTokens...)
	return tokens
}

func validateOptionalGroundedText(name string, value *GroundedText, enabled bool, maxRunes int, facts map[string]Fact) error {
	if !enabled {
		if value != nil {
			return fmt.Errorf("%s was not requested", name)
		}
		return nil
	}
	if value == nil {
		return fmt.Errorf("%s was requested but missing", name)
	}
	return validateGroundedText(name, *value, maxRunes, facts)
}

func validateGroundedText(name string, value GroundedText, maxRunes int, facts map[string]Fact) error {
	if err := validatePlainText(name, value.Text, maxRunes); err != nil {
		return err
	}
	if len(value.ReferenceIDs) == 0 {
		return fmt.Errorf("%s has no fact references", name)
	}
	seen := make(map[string]struct{}, len(value.ReferenceIDs))
	sourceText := strings.Builder{}
	for _, referenceID := range value.ReferenceIDs {
		fact, exists := facts[referenceID]
		if !exists {
			return fmt.Errorf("%s cites unknown fact %q", name, referenceID)
		}
		if _, duplicate := seen[referenceID]; duplicate {
			return fmt.Errorf("%s repeats fact reference %q", name, referenceID)
		}
		seen[referenceID] = struct{}{}
		sourceText.WriteString(fact.Text)
		sourceText.WriteByte(' ')
	}
	if err := rejectInventedNumericTokens(sourceText.String(), value.Text); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	if err := rejectContradictoryProtectedTokens(value.Text, value.ReferenceIDs, facts); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	if err := requireProtectedNumericAssociations(value.Text, value.ReferenceIDs, facts); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

// requireProtectedNumericAssociations prevents a generated surface from
// retaining the right numbers while assigning them to the wrong roles. A
// caller expresses a role as a protected phrase (for example, "current value
// is 15%" or "ends on August 31, 2026"). The model may write freely around
// that phrase, but it cannot detach a number or date from its label.
func requireProtectedNumericAssociations(generated string, citedReferenceIDs []string, facts map[string]Fact) error {
	generatedTokens := numericTokens(generated)
	orderedTokens := make([]string, 0, len(generatedTokens))
	for token := range generatedTokens {
		orderedTokens = append(orderedTokens, token)
	}
	sort.Strings(orderedTokens)

	for _, numericToken := range orderedTokens {
		associations := make([]string, 0)
		seenAssociations := make(map[string]struct{})
		for _, referenceID := range citedReferenceIDs {
			for _, protectedToken := range factProtectedTokens(facts[referenceID]) {
				if _, containsNumericToken := numericTokens(protectedToken)[numericToken]; !containsNumericToken {
					continue
				}
				if _, duplicate := seenAssociations[protectedToken]; duplicate {
					continue
				}
				seenAssociations[protectedToken] = struct{}{}
				associations = append(associations, protectedToken)
			}
		}
		if len(associations) == 0 {
			continue
		}
		sort.Strings(associations)
		for _, association := range associations {
			if strings.Contains(generated, association) {
				associations = nil
				break
			}
		}
		if associations != nil {
			return fmt.Errorf("used numeric/date token %q without one of its protected factual phrases", numericToken)
		}
	}
	return nil
}

func rejectContradictoryProtectedTokens(generated string, citedReferenceIDs []string, facts map[string]Fact) error {
	allowed := make(map[string]struct{})
	for _, referenceID := range citedReferenceIDs {
		for _, token := range factProtectedTokens(facts[referenceID]) {
			if normalized := normalizeProtectedToken(token); normalized != "" {
				allowed[normalized] = struct{}{}
			}
		}
	}
	all := make(map[string]string)
	for _, fact := range facts {
		for _, token := range factProtectedTokens(fact) {
			if normalized := normalizeProtectedToken(token); normalized != "" {
				all[normalized] = token
			}
		}
	}
	generatedLower := strings.ToLower(generated)
	for normalized, original := range all {
		if _, ok := allowed[normalized]; ok {
			continue
		}
		if strings.Contains(generatedLower, normalized) {
			return fmt.Errorf("used protected token %q without supporting fact", original)
		}
	}
	return nil
}

func normalizeProtectedToken(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func validatePlainText(name, value string, maxRunes int) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fmt.Errorf("%s is empty", name)
	}
	if utf8.RuneCountInString(trimmed) > maxRunes {
		return fmt.Errorf("%s exceeds %d characters", name, maxRunes)
	}
	lower := strings.ToLower(trimmed)
	if strings.ContainsAny(trimmed, "<>\r\n") || strings.Contains(lower, "http://") || strings.Contains(lower, "https://") || strings.Contains(lower, "www.") {
		return fmt.Errorf("%s contains markup, a URL, or a line break", name)
	}
	return nil
}

func requireNumericTokens(source, generated string) error {
	sourceTokens := numericTokens(source)
	generatedTokens := numericTokens(generated)
	for token := range sourceTokens {
		if _, exists := generatedTokens[token]; !exists {
			return fmt.Errorf("omitted protected numeric/date token %q", token)
		}
	}
	return rejectInventedNumericTokens(source, generated)
}

func rejectInventedNumericTokens(source, generated string) error {
	sourceTokens := numericTokens(source)
	for token := range numericTokens(generated) {
		if _, exists := sourceTokens[token]; !exists {
			return fmt.Errorf("invented numeric/date token %q", token)
		}
	}
	return nil
}

func numericTokens(value string) map[string]struct{} {
	tokens := make(map[string]struct{})
	for _, token := range numericTokenPattern.FindAllString(strings.ToLower(value), -1) {
		tokens[token] = struct{}{}
	}
	return tokens
}
