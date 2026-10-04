package notifications

import (
	"strings"

	"golang.org/x/net/html"
)

// commentPreview extracts visible text once. The plain_text variable marker
// keeps clients from decoding entities or interpreting literal tags again.
func commentPreview(content string) string {
	tokenizer := html.NewTokenizer(strings.NewReader(content))
	var text strings.Builder
	hiddenName := ""
	hiddenDepth := 0
	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			return strings.Join(strings.Fields(text.String()), " ")
		}
		token := tokenizer.Token()
		switch tokenType {
		case html.StartTagToken, html.SelfClosingTagToken:
			if hiddenDepth > 0 {
				if tokenType == html.StartTagToken && token.Data == hiddenName {
					hiddenDepth++
				}
			} else if hiddenCommentElement(token.Data) {
				if tokenType == html.StartTagToken {
					hiddenName, hiddenDepth = token.Data, 1
				}
			} else if blockCommentElement(token.Data) {
				text.WriteByte(' ')
			}
		case html.EndTagToken:
			if hiddenDepth > 0 {
				if token.Data == hiddenName {
					hiddenDepth--
				}
			} else if blockCommentElement(token.Data) {
				text.WriteByte(' ')
			}
		case html.TextToken:
			if hiddenDepth == 0 {
				text.WriteString(token.Data)
			}
		}
	}
}

func hiddenCommentElement(name string) bool {
	switch name {
	case "head", "iframe", "noscript", "object", "script", "style", "svg", "template":
		return true
	default:
		return false
	}
}

func blockCommentElement(name string) bool {
	switch name {
	case "blockquote", "br", "div", "h1", "h2", "h3", "h4", "h5", "h6", "hr", "li", "ol", "p", "pre", "table", "td", "th", "tr", "ul":
		return true
	default:
		return false
	}
}
