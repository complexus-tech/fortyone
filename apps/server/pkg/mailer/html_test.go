package mailer

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSafeEmailHTMLPreservesGeneratedMarkup(t *testing.T) {
	t.Parallel()

	style := EmailStyleString("notificationLink")
	input := `<div><p>Review <a href="https://app.fortyone.test/acme/work/PRD-42?tab=overview&amp;key=1" style="` + style + `">Launch &amp; learn</a>.</p></div>`

	sanitized := string(safeEmailHTML(input))

	require.Contains(t, sanitized, `<a href="https://app.fortyone.test/acme/work/PRD-42?tab=overview&amp;key=1" style="`+style+`">Launch &amp; learn</a>`)
}

func TestSafeEmailHTMLRemovesExecutableMarkupAndUnapprovedCSS(t *testing.T) {
	t.Parallel()

	input := `<script>alert("secret")</script><svg><script>alert(2)</script></svg>` +
		`<p style="background-image:url(https://attacker.test/track)">Hello <img src=x onerror=alert(1)></p>` +
		`<a href="javascript:alert(1)" onclick="alert(2)">Open</a>`

	sanitized := string(safeEmailHTML(input))

	require.Equal(t, "<p>Hello </p><a>Open</a>", sanitized)
	for _, forbidden := range []string{"script", "svg", "img", "javascript:", "onclick", "background-image", "attacker.test", "secret"} {
		require.NotContains(t, strings.ToLower(sanitized), forbidden)
	}
}

func TestSafeEmailHTMLEscapesTextAndBalancesAllowedElements(t *testing.T) {
	t.Parallel()

	sanitized := string(safeEmailHTML(`<div><strong>one</div><p>two & three`))

	require.Equal(t, `<div><strong>one</strong></div><p>two &amp; three</p>`, sanitized)
}
