package emailagent

import (
	"errors"
	"fmt"
	"html"
	"strings"
	"unicode/utf8"
)

const (
	maxCopyBlocks     = 16
	maxBlockTextRunes = 4_000
	maxBlockItems     = 20
	maxBlockItemRunes = 1_000
	maxPlainTextRunes = 16_000
)

func resolveEmailCopy(draft DraftEmailCopy, references map[string]struct{}) (EmailCopy, error) {
	referenceCatalog := make(map[string]struct{}, len(references)+1)
	for reference := range references {
		referenceCatalog[reference] = struct{}{}
	}
	referenceCatalog[proposalReference] = struct{}{}
	draft.Subject.Text = strings.TrimSpace(draft.Subject.Text)
	if err := validateSubject(draft.Subject.Text); err != nil {
		return EmailCopy{}, err
	}
	if err := validateReferences(draft.Subject.References, referenceCatalog, "subject"); err != nil {
		return EmailCopy{}, err
	}
	if len(draft.Blocks) == 0 || len(draft.Blocks) > maxCopyBlocks {
		return EmailCopy{}, fmt.Errorf("%w: copy must contain between 1 and %d blocks", ErrInvalidDecision, maxCopyBlocks)
	}

	blocks := make([]CopyBlock, len(draft.Blocks))
	for index, block := range draft.Blocks {
		block.Text = strings.TrimSpace(block.Text)
		items := make([]string, len(block.Items))
		for itemIndex, item := range block.Items {
			items[itemIndex] = strings.TrimSpace(item)
		}
		block.Items = items
		block.References = append([]string(nil), block.References...)
		if err := validateCopyBlock(block); err != nil {
			return EmailCopy{}, fmt.Errorf("%w: block %d: %v", ErrInvalidDecision, index, err)
		}
		if err := validateReferences(block.References, referenceCatalog, fmt.Sprintf("block %d", index)); err != nil {
			return EmailCopy{}, err
		}
		blocks[index] = block
	}

	plainText := RenderPlainText(blocks)
	if utf8.RuneCountInString(plainText) > maxPlainTextRunes {
		return EmailCopy{}, fmt.Errorf("%w: derived plain text exceeds %d runes", ErrInvalidDecision, maxPlainTextRunes)
	}
	return EmailCopy{
		Subject:   draft.Subject.Text,
		PlainText: plainText,
		Blocks:    blocks,
	}, nil
}

// RenderPlainText derives the text alternative from the same blocks used for
// HTML so the two representations cannot make different factual claims.
func RenderPlainText(blocks []CopyBlock) string {
	sections := make([]string, 0, len(blocks))
	for _, block := range blocks {
		var section strings.Builder
		if block.Text != "" {
			section.WriteString(block.Text)
		}
		if block.Kind == CopyBlockBulletList {
			for _, item := range block.Items {
				if section.Len() > 0 {
					section.WriteByte('\n')
				}
				section.WriteString("- ")
				section.WriteString(item)
			}
		}
		if section.Len() > 0 {
			sections = append(sections, section.String())
		}
	}
	return strings.Join(sections, "\n\n")
}

// RenderHTML renders safe inner email HTML. Every model-authored value is
// escaped and no href, style, image, or arbitrary element can be introduced.
func RenderHTML(copy EmailCopy) (string, error) {
	if err := validateSubject(copy.Subject); err != nil {
		return "", err
	}
	if len(copy.Blocks) == 0 || len(copy.Blocks) > maxCopyBlocks {
		return "", fmt.Errorf("%w: copy must contain between 1 and %d blocks", ErrInvalidDecision, maxCopyBlocks)
	}
	if copy.PlainText != RenderPlainText(copy.Blocks) {
		return "", fmt.Errorf("%w: plain text does not match copy blocks", ErrInvalidDecision)
	}

	var output strings.Builder
	for index, block := range copy.Blocks {
		if err := validateCopyBlock(block); err != nil {
			return "", fmt.Errorf("%w: block %d: %v", ErrInvalidDecision, index, err)
		}
		switch block.Kind {
		case CopyBlockParagraph:
			output.WriteString("<p>")
			output.WriteString(escapeText(block.Text))
			output.WriteString("</p>")
		case CopyBlockCallout:
			output.WriteString(`<div role="note"><p>`)
			output.WriteString(escapeText(block.Text))
			output.WriteString("</p></div>")
		case CopyBlockBulletList:
			if block.Text != "" {
				output.WriteString("<p>")
				output.WriteString(escapeText(block.Text))
				output.WriteString("</p>")
			}
			output.WriteString("<ul>")
			for _, item := range block.Items {
				output.WriteString("<li>")
				output.WriteString(escapeText(item))
				output.WriteString("</li>")
			}
			output.WriteString("</ul>")
		}
	}
	return output.String(), nil
}

func validateSubject(subject string) error {
	if subject == "" {
		return fmt.Errorf("%w: subject is required", ErrInvalidDecision)
	}
	if strings.ContainsAny(subject, "\r\n\x00") {
		return fmt.Errorf("%w: subject contains a forbidden control character", ErrInvalidDecision)
	}
	if utf8.RuneCountInString(subject) > maxSubjectRunes {
		return fmt.Errorf("%w: subject exceeds %d runes", ErrInvalidDecision, maxSubjectRunes)
	}
	if containsModelAuthoredURL(subject) {
		return fmt.Errorf("%w: subject contains a model-authored URL", ErrInvalidDecision)
	}
	return nil
}

func validateCopyBlock(block CopyBlock) error {
	if strings.ContainsRune(block.Text, '\x00') {
		return errors.New("block text contains a null byte")
	}
	if utf8.RuneCountInString(block.Text) > maxBlockTextRunes {
		return fmt.Errorf("block text exceeds %d runes", maxBlockTextRunes)
	}
	if containsModelAuthoredURL(block.Text) {
		return errors.New("block text contains a model-authored URL")
	}
	if len(block.Items) > maxBlockItems {
		return fmt.Errorf("block has more than %d items", maxBlockItems)
	}
	for _, item := range block.Items {
		if item == "" {
			return errors.New("block item cannot be empty")
		}
		if strings.ContainsRune(item, '\x00') {
			return errors.New("block item contains a null byte")
		}
		if utf8.RuneCountInString(item) > maxBlockItemRunes {
			return fmt.Errorf("block item exceeds %d runes", maxBlockItemRunes)
		}
		if containsModelAuthoredURL(item) {
			return errors.New("block item contains a model-authored URL")
		}
	}
	switch block.Kind {
	case CopyBlockParagraph, CopyBlockCallout:
		if block.Text == "" {
			return errors.New("text block cannot be empty")
		}
		if len(block.Items) != 0 {
			return errors.New("text block cannot contain list items")
		}
	case CopyBlockBulletList:
		if len(block.Items) == 0 {
			return errors.New("bullet list requires at least one item")
		}
	default:
		return fmt.Errorf("unsupported block kind %q", block.Kind)
	}
	return nil
}

func validateReferences(values []string, allowed map[string]struct{}, location string) error {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, exists := allowed[value]; !exists {
			return fmt.Errorf("%w: %s cites unknown reference %q", ErrInvalidDecision, location, value)
		}
		if _, exists := seen[value]; exists {
			return fmt.Errorf("%w: %s repeats reference %q", ErrInvalidDecision, location, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func containsModelAuthoredURL(value string) bool {
	lower := strings.ToLower(value)
	return strings.Contains(lower, "http://") ||
		strings.Contains(lower, "https://") ||
		strings.Contains(lower, "mailto:") ||
		strings.Contains(lower, "www.")
}

func escapeText(value string) string {
	escaped := html.EscapeString(value)
	escaped = strings.ReplaceAll(escaped, "\r\n", "\n")
	escaped = strings.ReplaceAll(escaped, "\r", "\n")
	return strings.ReplaceAll(escaped, "\n", "<br>")
}

func sanitizeSubject(subject string) string {
	subject = strings.ReplaceAll(subject, "\r", " ")
	subject = strings.ReplaceAll(subject, "\n", " ")
	return strings.Join(strings.Fields(subject), " ")
}

func replySubject(subject string) string {
	subject = sanitizeSubject(subject)
	if subject == "" {
		return "Your update with Maya"
	}
	if !strings.HasPrefix(strings.ToLower(subject), "re:") {
		subject = "Re: " + subject
	}
	bounded, _ := truncateRunes(subject, maxSubjectRunes)
	return bounded
}
