// Package agentmentions validates explicit rich-text agent targets across
// product surfaces. Display names and quoted content never authorize a run.
package agentmentions

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/net/html"
)

var ErrInvalidMention = errors.New("invalid agent mention")

// ValidateMaya requires both the typed request metadata and a canonical mention
// node. The adapter supplies Maya's resolved identity, never the client.
func ValidateMaya(content string, agents []string, mayaID uuid.UUID) (bool, error) {
	if len(agents) == 0 {
		return false, nil
	}
	if len(agents) != 1 || agents[0] != "maya" || mayaID == uuid.Nil {
		return false, ErrInvalidMention
	}
	document, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return false, ErrInvalidMention
	}
	var contains func(*html.Node) bool
	contains = func(node *html.Node) bool {
		if node.Type == html.ElementNode {
			switch node.Data {
			case "blockquote", "code", "pre", "script", "style", "template":
				return false
			}
			if node.Data == "a" || node.Data == "span" {
				attributes := make(map[string]string, len(node.Attr))
				for _, attribute := range node.Attr {
					attributes[attribute.Key] = attribute.Val
				}
				id, parseErr := uuid.Parse(attributes["data-id"])
				if parseErr == nil && id == mayaID && attributes["data-type"] == "mention" &&
					attributes["data-mention-kind"] == "agent" && attributes["data-agent-key"] == "maya" {
					return true
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if contains(child) {
				return true
			}
		}
		return false
	}
	if !contains(document) {
		return false, ErrInvalidMention
	}
	return true, nil
}

// Text extracts visible text for bounded conversation context. It never returns
// markup, executable nodes, or attachment download URLs.
func Text(content string) string {
	document, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return ""
	}
	var result strings.Builder
	var visit func(*html.Node)
	visit = func(node *html.Node) {
		if node.Type == html.ElementNode {
			switch node.Data {
			case "script", "style", "template":
				return
			case "p", "div", "li", "br", "h1", "h2", "h3", "blockquote":
				result.WriteByte('\n')
			}
		}
		if node.Type == html.TextNode {
			result.WriteString(node.Data)
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(document)
	return strings.TrimSpace(result.String())
}
