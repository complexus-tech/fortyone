package messaging

import (
	"bytes"
	"fmt"
	stdhtml "html"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const (
	maximumNativeDescriptionRunes     = 16_000
	maximumNativeDescriptionHTMLBytes = 256 << 10
)

// rewriteDescription replaces prose, not document markup. Existing media and
// structured widgets are retained in order after the new prose. The model never
// supplies HTML or media attributes. Both stored representations come from the
// same resulting document, and the caller binds it to the source UpdatedAt.
func rewriteDescription(existing *string, replacement string) (string, string, error) {
	replacement = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(replacement, "\r\n", "\n"), "\r", "\n"))
	if !utf8.ValidString(replacement) || utf8.RuneCountInString(replacement) > maximumNativeDescriptionRunes || strings.ContainsRune(replacement, '\x00') {
		return "", "", fmt.Errorf("%w: description must be valid text of at most %d characters", ErrInvalidToolArguments, maximumNativeDescriptionRunes)
	}
	var original string
	if existing != nil {
		original = *existing
	}
	if len(original) > maximumNativeDescriptionHTMLBytes {
		return "", "", fmt.Errorf("%w: existing rich description is too large for a safe rewrite", ErrInvalidToolArguments)
	}
	nodes, err := html.ParseFragment(strings.NewReader(original), &html.Node{Type: html.ElementNode, Data: "div", DataAtom: atom.Div})
	if err != nil {
		return "", "", fmt.Errorf("%w: parse existing description: %v", ErrInvalidToolArguments, err)
	}
	var media []*html.Node
	var mediaErr error
	var collect func(*html.Node)
	collect = func(node *html.Node) {
		if mediaErr != nil {
			return
		}
		if preservedDescriptionNode(node) {
			if err := validatePreservedDescriptionNode(node); err != nil {
				mediaErr = err
				return
			}
			media = append(media, node)
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			collect(child)
		}
	}
	for _, node := range nodes {
		collect(node)
	}
	if mediaErr != nil {
		return "", "", mediaErr
	}
	var output bytes.Buffer
	if replacement != "" {
		for _, paragraph := range strings.Split(replacement, "\n\n") {
			output.WriteString("<p>")
			output.WriteString(strings.ReplaceAll(stdhtml.EscapeString(paragraph), "\n", "<br>"))
			output.WriteString("</p>")
		}
	}
	textParts := []string{replacement}
	for _, node := range media {
		if node.Data == "a" {
			output.WriteString("<p>")
		}
		if err := html.Render(&output, node); err != nil {
			return "", "", fmt.Errorf("render preserved description media: %w", err)
		}
		if node.Data == "a" {
			output.WriteString("</p>")
		}
		if text := descriptionNodeText(node); text != "" {
			textParts = append(textParts, text)
		}
	}
	if output.Len() == 0 {
		output.WriteString("<p></p>")
	}
	if output.Len() > maximumNativeDescriptionHTMLBytes {
		return "", "", fmt.Errorf("%w: rewritten rich description is too large", ErrInvalidToolArguments)
	}
	return strings.TrimSpace(strings.Join(textParts, "\n\n")), output.String(), nil
}

func preservedDescriptionNode(node *html.Node) bool {
	if node.Type != html.ElementNode {
		return false
	}
	if node.Data == "a" {
		hasReference := false
		for _, attribute := range node.Attr {
			if (attribute.Key == "data-type" && attribute.Val == "mention") || attribute.Key == "data-mention-kind" || attribute.Key == "data-agent-key" {
				return false
			}
			if attribute.Key == "href" && strings.TrimSpace(attribute.Val) != "" {
				hasReference = true
			}
		}
		return hasReference
	}
	switch node.Data {
	case "img", "picture", "video", "audio", "iframe", "object", "embed", "figure", "table", "pre", "hr":
		return true
	}
	for _, attribute := range node.Attr {
		if attribute.Key == "data-type" && attribute.Val != "mention" {
			return true
		}
		if strings.HasPrefix(attribute.Key, "data-") && (strings.Contains(attribute.Key, "attachment") || strings.Contains(attribute.Key, "file") || strings.Contains(attribute.Key, "video") || strings.Contains(attribute.Key, "drawing")) {
			return true
		}
	}
	return false
}

func descriptionNodeText(node *html.Node) string {
	var output strings.Builder
	var visit func(*html.Node)
	visit = func(current *html.Node) {
		if current.Type == html.TextNode {
			output.WriteString(current.Data)
			return
		}
		if current.Data == "script" || current.Data == "style" {
			return
		}
		if current.Data == "br" {
			output.WriteByte('\n')
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
		switch current.Data {
		case "p", "div", "li", "tr", "figcaption":
			output.WriteString("\n\n")
		}
	}
	visit(node)
	return strings.TrimSpace(output.String())
}

// Refuse a rewrite rather than copying executable attributes into the new
// document or silently deleting a user's media. Ordinary editor attributes
// (attachment IDs, dimensions, alignment, controls and styles) remain intact.
func validatePreservedDescriptionNode(node *html.Node) error {
	switch node.Data {
	case "script", "style", "base", "link", "meta":
		return fmt.Errorf("%w: existing media contains unsupported active markup", ErrInvalidToolArguments)
	}
	for _, attribute := range node.Attr {
		name := strings.ToLower(attribute.Key)
		value := strings.ToLower(strings.Map(func(r rune) rune {
			if r <= ' ' {
				return -1
			}
			return r
		}, attribute.Val))
		if strings.HasPrefix(name, "on") || name == "srcdoc" ||
			((name == "src" || name == "href" || name == "data" || name == "poster" || name == "action") && (strings.HasPrefix(value, "javascript:") || strings.HasPrefix(value, "vbscript:") || strings.HasPrefix(value, "data:") || strings.HasPrefix(value, "file:"))) ||
			(name == "style" && (strings.Contains(value, "expression(") || strings.Contains(value, "javascript:") || strings.Contains(value, "@import"))) {
			return fmt.Errorf("%w: existing media contains unsafe attributes; edit that media before rewriting the description", ErrInvalidToolArguments)
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if err := validatePreservedDescriptionNode(child); err != nil {
			return err
		}
	}
	return nil
}
