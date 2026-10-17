package chatadapter

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	chat "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
)

func notificationCopy(content chat.NotificationContent) (string, string) {
	author := strings.Join(strings.Fields(content.AuthorName), " ")
	if content.ReactionEmoji != "" {
		author = strings.Join(strings.Fields(content.ActorName), " ")
	}
	if author == "" {
		author = "Someone"
	}
	title := author
	if content.ConversationKind == "team" {
		title = strings.TrimSpace(content.ConversationName)
		if title == "" {
			title = "Team chat"
		}
	}

	body := messagePreview(content.Text, content.RichContent)
	if content.ReactionEmoji != "" {
		prefix := author + " reacted " + content.ReactionEmoji + " to your message"
		if body != "" {
			prefix += ": " + body
		}
		return notificationText(title, 256), notificationText(prefix, 1024)
	}
	if body == "" {
		switch len(content.Attachments) {
		case 0:
			body = "Sent a message"
		case 1:
			attachment := content.Attachments[0]
			switch {
			case attachment.Kind == "voice-note":
				body = "Sent a voice message"
			case strings.HasPrefix(attachment.MediaType, "image/"):
				body = "Sent an image"
			case strings.TrimSpace(attachment.FileName) != "":
				body = "Shared " + attachment.FileName
			default:
				body = "Sent an attachment"
			}
		default:
			body = fmt.Sprintf("Sent %d attachments", len(content.Attachments))
		}
	}
	if content.ConversationKind == "team" {
		body = author + ": " + body
	}
	return notificationText(title, 256), notificationText(body, 1024)
}

type previewNode struct {
	Type    string        `json:"type"`
	Text    string        `json:"text"`
	Content []previewNode `json:"content"`
	Attrs   struct {
		Label string `json:"label"`
	} `json:"attrs"`
}

// Match the rich message's visible text, including mention labels and block
// boundaries. Plain messages remain literal text, just as in the chat renderer.
func messagePreview(plain string, rich json.RawMessage) string {
	var document previewNode
	if json.Unmarshal(rich, &document) != nil || document.Type != "doc" {
		return strings.Join(strings.Fields(plain), " ")
	}
	var text strings.Builder
	var walk func(previewNode)
	walk = func(node previewNode) {
		switch node.Type {
		case "text":
			text.WriteString(node.Text)
		case "mention":
			label := strings.TrimSpace(node.Attrs.Label)
			if label == "" {
				label = "Member"
			}
			text.WriteString("@" + strings.TrimPrefix(label, "@"))
		case "hardBreak":
			text.WriteByte(' ')
		default:
			for _, child := range node.Content {
				walk(child)
			}
			switch node.Type {
			case "paragraph", "heading", "listItem", "blockquote", "codeBlock":
				text.WriteByte(' ')
			}
		}
	}
	walk(document)
	return strings.Join(strings.Fields(text.String()), " ")
}

// Publisher and native limits are bytes; never cut a multibyte character.
func notificationText(value string, limit int) string {
	value = strings.Join(strings.Fields(value), " ")
	if len(value) <= limit {
		return value
	}
	end := limit - len("…")
	for !utf8.RuneStart(value[end]) {
		end--
	}
	return strings.TrimSpace(value[:end]) + "…"
}
