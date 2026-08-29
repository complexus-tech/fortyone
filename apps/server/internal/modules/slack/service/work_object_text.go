package slack

import (
	"errors"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/net/html"
)

func truncateSlackWorkObjectText(value string, limit int) string {
	value = strings.TrimSpace(value)
	if value == "" || limit <= 0 || utf8.RuneCountInString(value) <= limit {
		return value
	}
	runes := []rune(value)
	return strings.TrimSpace(string(runes[:limit-1])) + "…"
}

// slackWorkObjectDescription converts rich editor HTML to readable text before
// it enters a Slack Work Object. Plain-text descriptions are returned exactly
// as written so code examples such as "value <T>" are not mistaken for HTML.
func slackWorkObjectDescription(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || !strings.Contains(value, "<") {
		return value
	}

	tokenizer := html.NewTokenizer(strings.NewReader(value))
	var output strings.Builder
	sawRichTextMarkup := false
	suppressedDepth := 0
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			if err := tokenizer.Err(); err != nil && !errors.Is(err, io.EOF) {
				return value
			}
			if !sawRichTextMarkup {
				return value
			}
			return normalizeSlackWorkObjectPlainText(output.String())
		case html.TextToken:
			if suppressedDepth == 0 {
				output.Write(tokenizer.Text())
			}
		case html.StartTagToken, html.SelfClosingTagToken, html.EndTagToken:
			token := tokenizer.Token()
			tag := strings.ToLower(token.Data)
			if !isSlackRichTextHTMLTag(tag) {
				continue
			}
			sawRichTextMarkup = true
			if tag == "script" || tag == "style" {
				if tokenType == html.StartTagToken {
					suppressedDepth++
				} else if tokenType == html.EndTagToken && suppressedDepth > 0 {
					suppressedDepth--
				}
				continue
			}
			if suppressedDepth > 0 {
				continue
			}
			if isSlackRichTextBlockTag(tag) {
				appendSlackWorkObjectLineBreak(&output)
			}
			if tag == "li" && tokenType != html.EndTagToken {
				output.WriteString("• ")
			}
		}
	}
}

func isSlackRichTextHTMLTag(tag string) bool {
	switch tag {
	case "a", "b", "blockquote", "br", "code", "del", "div", "em", "h1", "h2", "h3", "h4", "h5", "h6", "hr", "i", "li", "ol", "p", "pre", "s", "script", "span", "strong", "style", "table", "tbody", "td", "th", "thead", "tr", "u", "ul":
		return true
	default:
		return false
	}
}

func isSlackRichTextBlockTag(tag string) bool {
	switch tag {
	case "blockquote", "br", "div", "h1", "h2", "h3", "h4", "h5", "h6", "hr", "li", "ol", "p", "pre", "table", "tbody", "td", "th", "thead", "tr", "ul":
		return true
	default:
		return false
	}
}

func appendSlackWorkObjectLineBreak(output *strings.Builder) {
	current := output.String()
	if current != "" && !strings.HasSuffix(current, "\n") {
		output.WriteByte('\n')
	}
}

func normalizeSlackWorkObjectPlainText(value string) string {
	lines := strings.Split(strings.ReplaceAll(value, "\u00a0", " "), "\n")
	normalized := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.Join(strings.Fields(line), " ")
		if line == "" {
			continue
		}
		normalized = append(normalized, line)
	}
	return strings.Join(normalized, "\n")
}

func unixTimestamp(value time.Time) int64 {
	if value.IsZero() {
		return 0
	}
	return value.UTC().Unix()
}
