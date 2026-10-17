package agentadapter

import (
	"encoding/json"
	"strings"

	"github.com/complexus-tech/projects-api/internal/platform/agentmentions"
	"github.com/google/uuid"
	"golang.org/x/net/html"
)

type chatContentNode struct {
	Type    string            `json:"type"`
	Text    string            `json:"text,omitempty"`
	Attrs   map[string]any    `json:"attrs,omitempty"`
	Marks   []chatContentNode `json:"marks,omitempty"`
	Content []chatContentNode `json:"content,omitempty"`
}

func validateChatMayaMention(content json.RawMessage, agents []string, mayaID uuid.UUID) error {
	if len(agents) != 1 || agents[0] != "maya" || mayaID == uuid.Nil {
		return agentmentions.ErrInvalidMention
	}
	var root chatContentNode
	if json.Unmarshal(content, &root) != nil || root.Type != "doc" {
		return agentmentions.ErrInvalidMention
	}
	var contains func(chatContentNode, int) bool
	contains = func(node chatContentNode, depth int) bool {
		if depth > 32 || node.Type == "blockquote" || node.Type == "codeBlock" {
			return false
		}
		for _, mark := range node.Marks {
			if mark.Type == "code" {
				return false
			}
		}
		if node.Type == "mention" && node.Attrs["id"] == mayaID.String() && node.Attrs["mentionKind"] == "agent" && node.Attrs["agentKey"] == "maya" {
			return true
		}
		for _, child := range node.Content {
			if contains(child, depth+1) {
				return true
			}
		}
		return false
	}
	if !contains(root, 0) {
		return agentmentions.ErrInvalidMention
	}
	return nil
}

// Model prose uses the same sanitized Markdown boundary as comments, then maps
// to the Chat editor's document format. Model HTML and mention attributes never
// become trusted editor nodes.
func chatReplyContent(markdown string) (string, json.RawMessage, error) {
	markup := replyHTML(markdown)
	document, err := html.Parse(strings.NewReader(markup))
	if err != nil {
		return "", nil, err
	}
	var body *html.Node
	var findBody func(*html.Node)
	findBody = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "body" {
			body = node
			return
		}
		for child := node.FirstChild; child != nil && body == nil; child = child.NextSibling {
			findBody(child)
		}
	}
	findBody(document)
	root := chatContentNode{Type: "doc"}
	if body != nil {
		root.Content = chatNodes(body, nil)
	}
	if len(root.Content) == 0 {
		root.Content = []chatContentNode{{Type: "paragraph"}}
	}
	content, err := json.Marshal(root)
	return agentmentions.Text(markup), content, err
}

func chatNodes(parent *html.Node, marks []chatContentNode) []chatContentNode {
	var result []chatContentNode
	for child := parent.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.TextNode {
			if strings.TrimSpace(child.Data) == "" && (parent.Data == "body" || parent.Data == "ul" || parent.Data == "ol") {
				continue
			}
			if child.Data != "" {
				result = append(result, chatContentNode{Type: "text", Text: child.Data, Marks: marks})
			}
			continue
		}
		if child.Type != html.ElementNode {
			continue
		}
		mark := chatContentNode{}
		switch child.Data {
		case "strong", "b":
			mark.Type = "bold"
		case "em", "i":
			mark.Type = "italic"
		case "s", "del":
			mark.Type = "strike"
		case "code":
			mark.Type = "code"
		case "a":
			mark.Type = "link"
			for _, attr := range child.Attr {
				if attr.Key == "href" {
					mark.Attrs = map[string]any{"href": attr.Val}
				}
			}
		}
		if mark.Type != "" {
			nested := append(append([]chatContentNode(nil), marks...), mark)
			result = append(result, chatNodes(child, nested)...)
			continue
		}
		node := chatContentNode{}
		switch child.Data {
		case "p":
			node.Type = "paragraph"
		case "ul":
			node.Type = "bulletList"
		case "ol":
			node.Type = "orderedList"
		case "li":
			node.Type = "listItem"
		case "blockquote":
			node.Type = "blockquote"
		case "pre":
			node.Type = "codeBlock"
		case "br":
			node.Type = "hardBreak"
		case "h1", "h2", "h3", "h4", "h5", "h6":
			node.Type = "paragraph"
		}
		if node.Type == "" {
			result = append(result, chatNodes(child, marks)...)
			continue
		}
		node.Content = chatNodes(child, marks)
		if node.Type == "listItem" {
			var blocks, inline []chatContentNode
			flush := func() {
				if len(inline) > 0 {
					blocks = append(blocks, chatContentNode{Type: "paragraph", Content: inline})
					inline = nil
				}
			}
			for _, item := range node.Content {
				if item.Type == "text" || item.Type == "hardBreak" {
					inline = append(inline, item)
				} else {
					flush()
					blocks = append(blocks, item)
				}
			}
			flush()
			if len(blocks) == 0 || blocks[0].Type != "paragraph" {
				blocks = append([]chatContentNode{{Type: "paragraph"}}, blocks...)
			}
			node.Content = blocks
		}
		if node.Type == "codeBlock" {
			for index := range node.Content {
				node.Content[index].Marks = nil
			}
		}
		result = append(result, node)
	}
	return result
}
