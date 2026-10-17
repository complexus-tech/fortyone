package agentadapter

import (
	"strings"
	"sync"
	"testing"
	"unicode/utf8"

	agentruns "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/service"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func TestMayaPublishesMarkdownAsRichCommentHTML(t *testing.T) {
	t.Parallel()
	runtime, run, _ := runtimeFixture()
	writer := &runtimeWriter{}
	runtime.Comments = writer
	_, err := runtime.Publish(t.Context(), run, agentruns.Outcome{Reply: "Suggested title: **Refine Onboarding Conversation**\n\n- Keep *recent* replies clear.\n- Use `commentId`.\n\n[Story](/first/work/PRO-98?comment=reply)"})
	require.NoError(t, err)
	content := writer.calls[0].Content
	require.Contains(t, content, "<strong>Refine Onboarding Conversation</strong>")
	require.Contains(t, content, "<ul>")
	require.Contains(t, content, "<em>recent</em>")
	require.Contains(t, content, "<code>commentId</code>")
	require.Contains(t, content, `<a href="/first/work/PRO-98?comment=reply" rel="noopener noreferrer">Story</a>`)
	require.NotContains(t, content, "**")
	require.Empty(t, writer.calls[0].MentionedUserIDs)
}

func TestReplyMarkdownPreservesCodeTablesAndOrderedListStart(t *testing.T) {
	t.Parallel()
	content := replyHTML("## Details\n\n3. Third\n4. Fourth\n\n> Keep context.\n\n~~Old~~\n\n```html\n<script>alert('code')</script>\n```\n\n| Field | Value |\n| --- | --- |\n| Title | Ready |")
	for _, fragment := range []string{"<h2>Details</h2>", `<ol start="3">`, "<blockquote>", "<del>Old</del>", "<pre><code>", "&lt;script&gt;", "<table>", "<th>Field</th>", "<td>Ready</td>"} {
		require.Contains(t, content, fragment)
	}
	require.NotContains(t, content, "<script>")
	require.NotContains(t, content, "class=")
}

func TestReplyMarkdownDoesNotCreateMediaMentionsOrExecutableMarkup(t *testing.T) {
	t.Parallel()
	content := replyHTML("**Safe**\n\n![Screenshot](https://files.example/private.png)\n\n[Unsafe](javascript:alert%281%29) [Data](data:text/html,script) [Relative](//untrusted.example/path)\n\n<script>alert(1)</script>\n\n<a href=\"https://example.com\" data-type=\"mention\" data-mention-kind=\"agent\" data-agent-key=\"maya\" onclick=\"alert(1)\">spoofed</a>\n\n<iframe src=\"https://example.com\"></iframe>")
	require.Contains(t, content, "<strong>Safe</strong>")
	require.Contains(t, content, "Screenshot")
	for _, forbidden := range []string{"<img", "<script", "<iframe", "onclick", "data-type", "data-agent", "href=\"javascript:", "href=\"data:", "href=\"//", "private.png"} {
		require.NotContains(t, content, forbidden)
	}
	// Defense in depth also strips unsafe attributes if a future renderer
	// extension emits them despite the Markdown parser's safe defaults.
	nodes, err := html.ParseFragment(strings.NewReader(`<p style="position:fixed" data-type="mention"><a href="jav&#x61;script:alert(1)" onclick="bad()">label</a><svg><script>bad()</script></svg><img src="x" onerror="bad()" alt="image label"></p>`), &html.Node{Type: html.ElementNode, Data: "div", DataAtom: atom.Div})
	require.NoError(t, err)
	var sanitized strings.Builder
	for _, node := range nodes {
		writeReplyNode(&sanitized, node, &replyMentions{})
	}
	require.Equal(t, "<p>labelimage label</p>", sanitized.String())
}

func TestReplyLinksAllowOnlySafeNavigation(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"https://example.com/path?q=1&other=2", "http://localhost:3000/first/work/PRO-98", "/first/work/PRO-98?comment=id", "../work/PRO-98", "#reply", "mailto:hello@example.com"} {
		require.True(t, safeReplyLink(value), value)
	}
	for _, value := range []string{"", "javascript:alert(1)", "JaVaScRiPt:alert(1)", "data:text/html,script", "file:///private/key", "vbscript:bad", "//example.com", `\example.com`, "java\nscript:bad", "https://user:password@example.com", "https://", " https://example.com"} {
		require.False(t, safeReplyLink(value), value)
	}
}

func TestReplyMarkdownBoundsMarkupAndSupportsConcurrentPublication(t *testing.T) {
	t.Parallel()
	input := strings.Repeat("**<&界** [link](https://example.com)\n", 1000)
	var group sync.WaitGroup
	outputs := make(chan string, 8)
	for range 8 {
		group.Go(func() { outputs <- replyHTML(input) })
	}
	group.Wait()
	close(outputs)
	var expected string
	for output := range outputs {
		require.True(t, utf8.ValidString(output))
		require.LessOrEqual(t, utf8.RuneCountInString(output), 9500)
		require.NotEmpty(t, output)
		if expected == "" {
			expected = output
		}
		require.Equal(t, expected, output)
	}
}
