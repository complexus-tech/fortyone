package agentadapter

import (
	stdhtml "html"
	"slices"
	"strings"

	"github.com/google/uuid"
)

const maximumReplyMentions = 10

type renderedReply struct {
	HTML             string
	MentionedUserIDs []uuid.UUID
}

type replyMentions struct {
	targets      map[uuid.UUID]string
	ids          []uuid.UUID
	collect      bool
	blockedDepth int
}

func replyMentionCandidates(text string) []uuid.UUID {
	mentions := &replyMentions{collect: true}
	renderReplyMarkdownWithMentions(boundedReplyText(text), mentions)
	return mentions.ids
}

func (m *replyMentions) write(output *strings.Builder, href string) bool {
	if m.blockedDepth != 0 || !strings.HasPrefix(href, "mention:") {
		return false
	}
	value := strings.TrimPrefix(href, "mention:")
	id, err := uuid.Parse(value)
	if err != nil || id == uuid.Nil || value != id.String() {
		return false
	}
	seen := slices.Contains(m.ids, id)
	if !seen && len(m.ids) >= maximumReplyMentions {
		return false
	}
	if m.collect {
		if !seen {
			m.ids = append(m.ids, id)
		}
		return false
	}
	name, allowed := m.targets[id]
	if !allowed || strings.TrimSpace(name) == "" {
		return false
	}
	if !seen {
		m.ids = append(m.ids, id)
	}
	label := stdhtml.EscapeString(name)
	output.WriteString(`<a href="/profile/` + id.String() + `" class="mention" data-type="mention" data-mention-kind="user" data-id="` + id.String() + `" data-label="` + label + `">@` + label + `</a>`)
	return true
}
