package emailreply

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderMayaReplyHTMLUsesProductShellAndDoesNotTrustTheSubject(t *testing.T) {
	t.Parallel()

	rendered, err := renderMayaReplyHTML(`Update <script>alert("x")</script>`, `<p>Safe response</p>`)

	require.NoError(t, err)
	require.Contains(t, rendered, "<!doctype html>")
	require.Contains(t, rendered, "https://fortyone.app/email-assets/v1/wordmark.png")
	require.Contains(t, rendered, "A product of Complexus")
	require.Contains(t, rendered, ">Safe response</p>")
	require.Contains(t, rendered, `font-family: Georgia, Times, serif`)
	require.NotContains(t, rendered, `<script>`)
	require.Contains(t, rendered, `&lt;script&gt;`)
}

func TestStyleMayaReplyBlocksUsesOnlyProductOwnedMarkup(t *testing.T) {
	t.Parallel()

	rendered := styleMayaReplyBlocks(`<div role="note"><p>Preview</p></div><ul><li>Health: On Track</li></ul>`)

	require.Contains(t, rendered, `role="note"`)
	require.Contains(t, rendered, `border-top: 1px solid`)
	require.Contains(t, rendered, `Health: On Track`)
	require.NotContains(t, rendered, `<p>Preview</p>`)
}
