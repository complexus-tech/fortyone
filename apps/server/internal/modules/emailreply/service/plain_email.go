package emailreply

import (
	"regexp"
	"strings"
	"unicode"
)

const maximumInboundReplyRunes = 32_000

var (
	markdownImagePattern  = regexp.MustCompile(`!\[([^]]*)\]\([^)]+\)`)
	markdownLinkPattern   = regexp.MustCompile(`\[([^]]+)\]\([^)]+\)`)
	markdownListPattern   = regexp.MustCompile(`^\s*(?:[-+*]|\d+[.)])\s+`)
	markdownHeaderPattern = regexp.MustCompile(`^\s{0,3}#{1,6}\s+`)
	markdownQuotePattern  = regexp.MustCompile(`^\s*>`)
	markdownFencePattern  = regexp.MustCompile("^\\s*```.*$")
)

// plainInboundReply keeps only the current visible reply. Brevo's extracted
// field is preferred because it already excludes quoted history and signature
// blocks; lightweight formatting markers are then removed before persistence.
func plainInboundReply(email InboundEmail) string {
	value := strings.TrimSpace(email.ExtractedMarkdownMessage)
	if value == "" && email.RawTextBody != nil {
		value = strings.TrimSpace(*email.RawTextBody)
	}
	if value == "" {
		return ""
	}
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	lines := strings.Split(value, "\n")
	kept := make([]string, 0, len(lines))
	inFence := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		if markdownFencePattern.MatchString(line) {
			inFence = !inFence
			continue
		}
		if !inFence && (markdownQuotePattern.MatchString(line) ||
			strings.HasPrefix(lower, "on ") && strings.HasSuffix(lower, " wrote:") ||
			lower == "-----original message-----" || lower == "---------- forwarded message ---------") {
			break
		}
		if !inFence && (trimmed == "--" || trimmed == "-- ") {
			break
		}
		line = markdownHeaderPattern.ReplaceAllString(line, "")
		line = markdownListPattern.ReplaceAllString(line, "")
		line = markdownImagePattern.ReplaceAllString(line, "$1")
		line = markdownLinkPattern.ReplaceAllString(line, "$1")
		line = stripInlineMarkdown(line)
		kept = append(kept, strings.TrimSpace(line))
	}
	return truncateRunes(collapseBlankLines(strings.Join(kept, "\n")), maximumInboundReplyRunes)
}

func stripInlineMarkdown(value string) string {
	value = strings.ReplaceAll(value, "**", "")
	value = strings.ReplaceAll(value, "__", "")
	value = strings.ReplaceAll(value, "~~", "")
	value = strings.ReplaceAll(value, "`", "")
	return strings.TrimFunc(value, func(r rune) bool {
		return unicode.IsSpace(r) || r == '*' || r == '_' || r == '~'
	})
}

func collapseBlankLines(value string) string {
	lines := strings.Split(strings.TrimSpace(value), "\n")
	result := make([]string, 0, len(lines))
	blank := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if blank || len(result) == 0 {
				continue
			}
			blank = true
			result = append(result, "")
			continue
		}
		blank = false
		result = append(result, line)
	}
	return strings.TrimSpace(strings.Join(result, "\n"))
}

func sanitizeEmailSubject(subject string) string {
	subject = strings.ReplaceAll(subject, "\r", " ")
	subject = strings.ReplaceAll(subject, "\n", " ")
	return truncateRunes(strings.Join(strings.Fields(subject), " "), 160)
}

func replyEmailSubject(subject string) string {
	subject = sanitizeEmailSubject(subject)
	if subject == "" {
		return "Your update with Maya"
	}
	if !strings.HasPrefix(strings.ToLower(subject), "re:") {
		subject = "Re: " + subject
	}
	return truncateRunes(subject, 160)
}

func safeInternetMessageID(value string) string {
	value = strings.TrimSpace(value)
	if len(value) < 3 || len(value) > 998 || value[0] != '<' || value[len(value)-1] != '>' ||
		strings.ContainsAny(value, "\r\n\x00") {
		return ""
	}
	return value
}

func safeOptionalInternetMessageID(value *string) string {
	if value == nil {
		return ""
	}
	return safeInternetMessageID(*value)
}
