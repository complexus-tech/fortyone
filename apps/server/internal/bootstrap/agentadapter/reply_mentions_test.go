package agentadapter

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/complexus-tech/projects-api/internal/platform/agentmentions"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func TestReplyMentionsUseTrustedNamesAndCanonicalHumanMarkup(t *testing.T) {
	t.Parallel()
	id := uuid.MustParse("abcde123-4567-4890-abcd-123456789abc")
	name := `Ada "Lovelace" & <script>alert('name')</script>`
	input := "Please ask [@Untrusted label](mention:" + id.String() + ") to review **this**."

	require.Equal(t, []uuid.UUID{id}, replyMentionCandidates(input))
	reply := renderMentionReply(input, map[uuid.UUID]string{id: name})
	require.Equal(t, []uuid.UUID{id}, reply.MentionedUserIDs)
	anchors := replyMentionAnchors(t, reply.HTML)
	require.Len(t, anchors, 1)
	anchor := anchors[0]
	require.Equal(t, "/profile/"+id.String(), replyAttribute(anchor, "href"))
	require.Equal(t, "mention", replyAttribute(anchor, "class"))
	require.Equal(t, "mention", replyAttribute(anchor, "data-type"))
	require.Equal(t, "user", replyAttribute(anchor, "data-mention-kind"))
	require.Equal(t, id.String(), replyAttribute(anchor, "data-id"))
	require.Equal(t, name, replyAttribute(anchor, "data-label"))
	require.Empty(t, replyAttribute(anchor, "data-agent-key"))
	require.NotNil(t, anchor.FirstChild)
	require.Equal(t, html.TextNode, anchor.FirstChild.Type)
	require.Equal(t, "@"+name, anchor.FirstChild.Data)
	require.Nil(t, anchor.FirstChild.NextSibling)
	require.Contains(t, reply.HTML, "<strong>this</strong>")
	require.NotContains(t, reply.HTML, "Untrusted label")
	require.NotContains(t, reply.HTML, "<script>")

	// Even when an ID matches, human mention markup cannot invoke the agent parser.
	invokesMaya, err := agentmentions.ValidateMaya(reply.HTML, []string{"maya"}, id)
	require.ErrorIs(t, err, agentmentions.ErrInvalidMention)
	require.False(t, invokesMaya)
}

func TestReplyMentionsIgnoreQuotedCodeAndRawHTML(t *testing.T) {
	t.Parallel()
	ignoredID := uuid.MustParse("abcde123-4567-4890-abcd-123456789abc")
	visibleID := uuid.MustParse("abcde123-4567-4890-abcd-123456789abd")
	ignored := "[@Quoted](mention:" + ignoredID.String() + ")"
	visible := "[@Visible](mention:" + visibleID.String() + ")"
	for _, test := range []struct {
		name string
		text string
	}{
		{"inline code", "`" + ignored + "`"},
		{"fenced code", "```markdown\n" + ignored + "\n```"},
		{"indented code", "    " + ignored},
		{"tab-indented code", "\t" + ignored},
		{"indented code after paragraph", "Example:\n\n    " + ignored},
		{"blockquote", "> " + ignored},
		{"nested blockquote", "> > " + ignored},
		{"raw anchor", `<a href="mention:` + ignoredID.String() + `" data-type="mention" data-id="` + ignoredID.String() + `">@Quoted</a>`},
		{"raw script", `<script>` + ignored + `</script>`},
		{"image destination", "!" + ignored},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := test.text + "\n\nOutside: " + visible
			require.Equal(t, []uuid.UUID{visibleID}, replyMentionCandidates(input))
			reply := renderMentionReply(input, map[uuid.UUID]string{
				ignoredID: "Quoted", visibleID: "Visible",
			})
			require.Equal(t, []uuid.UUID{visibleID}, reply.MentionedUserIDs)
			anchors := replyMentionAnchors(t, reply.HTML)
			require.Len(t, anchors, 1)
			require.Equal(t, visibleID.String(), replyAttribute(anchors[0], "data-id"))
			require.NotContains(t, reply.HTML, "<script>")
		})
	}
}

func TestReplyMentionsRejectNoncanonicalTargetsWithoutLosingLabels(t *testing.T) {
	t.Parallel()
	id := uuid.MustParse("abcde123-4567-4890-abcd-123456789abc")
	for _, target := range []string{
		"mention:not-a-uuid",
		"mention:" + uuid.Nil.String(),
		"mention:" + strings.ToUpper(id.String()),
		"mention:" + strings.ReplaceAll(id.String(), "-", ""),
		"mention:{" + id.String() + "}",
		"mention://" + id.String(),
		"mention:" + id.String() + "?notify=1",
		"mention:" + id.String() + "#person",
		"mention:" + id.String() + "/extra",
		"mention:%61" + id.String()[1:],
		"MENTION:" + id.String(),
	} {
		t.Run(target, func(t *testing.T) {
			input := "Please ask [@Readable label](" + target + ")."
			require.Empty(t, replyMentionCandidates(input))
			reply := renderMentionReply(input, map[uuid.UUID]string{id: "Trusted"})
			require.Empty(t, reply.MentionedUserIDs)
			require.Empty(t, replyMentionAnchors(t, reply.HTML))
			require.Equal(t, "Please ask @Readable label.", agentmentions.Text(reply.HTML))
		})
	}
}

func TestReplyMentionsLeaveUnknownOrUnavailableNamesAsPlainText(t *testing.T) {
	t.Parallel()
	unknownID := uuid.MustParse("abcde123-4567-4890-abcd-123456789abc")
	blankID := uuid.MustParse("abcde123-4567-4890-abcd-123456789abd")
	allowedID := uuid.MustParse("abcde123-4567-4890-abcd-123456789abe")
	input := fmt.Sprintf("[@Unknown](mention:%s), [@Unavailable](mention:%s), [@Allowed](mention:%s)", unknownID, blankID, allowedID)
	reply := renderMentionReply(input, map[uuid.UUID]string{blankID: " \t ", allowedID: "Grace"})
	require.Equal(t, []uuid.UUID{allowedID}, reply.MentionedUserIDs)
	require.Len(t, replyMentionAnchors(t, reply.HTML), 1)
	require.Equal(t, "@Unknown, @Unavailable, @Grace", agentmentions.Text(reply.HTML))
	require.NotContains(t, reply.HTML, "mention:")
}

func TestReplyMentionsBoundUniqueRecipientsWhileKeepingRepeatedTags(t *testing.T) {
	t.Parallel()
	names := make(map[uuid.UUID]string)
	ids := make([]uuid.UUID, 12)
	links := make([]string, 0, 13)
	for index := range ids {
		ids[index] = uuid.MustParse(fmt.Sprintf("10000000-0000-4000-8000-%012d", index+1))
		names[ids[index]] = fmt.Sprintf("Person %02d", index)
		links = append(links, fmt.Sprintf("[@Person %02d](mention:%s)", index, ids[index]))
	}
	// Repeating an accepted tag after reaching the limit must not consume a slot.
	links = append(links, links[0])
	input := strings.Join(links, ", ")
	require.Equal(t, ids[:10], replyMentionCandidates(input))
	reply := renderMentionReply(input, names)
	require.Equal(t, ids[:10], reply.MentionedUserIDs)
	require.Len(t, replyMentionAnchors(t, reply.HTML), 11)
	require.Contains(t, agentmentions.Text(reply.HTML), "@Person 10, @Person 11")
	require.NotContains(t, reply.HTML, `/profile/`+ids[10].String())
	require.NotContains(t, reply.HTML, `/profile/`+ids[11].String())
}

func TestReplyMentionsSourceTruncationExcludesDiscardedTargets(t *testing.T) {
	t.Parallel()
	firstID := uuid.MustParse("abcde123-4567-4890-abcd-123456789abc")
	lastID := uuid.MustParse("abcde123-4567-4890-abcd-123456789abd")
	input := "[@First](mention:" + firstID.String() + ") " + strings.Repeat("界", 6001) +
		" [@Discarded](mention:" + lastID.String() + ")"
	require.Equal(t, []uuid.UUID{firstID}, replyMentionCandidates(input))
	reply := renderMentionReply(input, map[uuid.UUID]string{firstID: "First", lastID: "Discarded"})
	require.Equal(t, []uuid.UUID{firstID}, reply.MentionedUserIDs)
	require.Len(t, replyMentionAnchors(t, reply.HTML), 1)
	require.NotContains(t, reply.HTML, "Discarded")
	require.True(t, utf8.ValidString(reply.HTML))
	require.LessOrEqual(t, utf8.RuneCountInString(reply.HTML), 9500)
}

func TestReplyMentionsOutputTruncationRecomputesRecipients(t *testing.T) {
	t.Parallel()
	firstID := uuid.MustParse("abcde123-4567-4890-abcd-123456789abc")
	lastID := uuid.MustParse("abcde123-4567-4890-abcd-123456789abd")
	// Escaping this source expands it past the rendered-HTML budget, while both
	// links fit comfortably inside the source budget and candidate resolution.
	input := "[@First](mention:" + firstID.String() + ") " + strings.Repeat("&", 2000) +
		" [@Discarded](mention:" + lastID.String() + ")"
	require.Less(t, utf8.RuneCountInString(input), 6000)
	require.Equal(t, []uuid.UUID{firstID, lastID}, replyMentionCandidates(input))
	reply := renderMentionReply(input, map[uuid.UUID]string{firstID: "First", lastID: "Discarded"})
	require.Equal(t, []uuid.UUID{firstID}, reply.MentionedUserIDs)
	require.Len(t, replyMentionAnchors(t, reply.HTML), 1)
	require.NotContains(t, reply.HTML, "Discarded")
	require.LessOrEqual(t, utf8.RuneCountInString(reply.HTML), 9500)
}

func replyMentionAnchors(t *testing.T, content string) []*html.Node {
	t.Helper()
	document, err := html.Parse(strings.NewReader(content))
	require.NoError(t, err)
	var anchors []*html.Node
	var visit func(*html.Node)
	visit = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "a" {
			anchors = append(anchors, node)
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(document)
	return anchors
}
