package notifications

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommentPreview(t *testing.T) {
	t.Parallel()
	const contactURL = "https://app.reversecontact.com/contact/find-phone-from-url"
	for _, test := range []struct {
		name, input, want string
	}{
		{"screenshot link", `<p><a target="_blank" rel="noopener noreferrer nofollow" href="` + contactURL + `">` + contactURL + `</a></p>`, contactURL},
		{"paragraphs and list", `<p>Review &amp; update</p><ul><li>First</li><li>Second<br>line</li></ul>`, "Review & update First Second line"},
		{"unsafe hidden elements", `<script>alert(1)</script><style>body{display:none}</style><iframe>hidden</iframe><p>Safe<img src=x onerror=alert(1)></p>`, "Safe"},
		{"hidden subtree with void elements", `<template><p>hidden<br><img src=x></p></template><p>Visible</p>`, "Visible"},
		{"literal entities decoded once", `<p>&lt;example&gt; &amp;amp; &#x1f44b; &#39;quote&#39;</p>`, "<example> &amp; 👋 'quote'"},
		{"plain comparison", "Priority < 5 & delivery > 2", "Priority < 5 & delivery > 2"},
		{"malformed markup", `<p title="a > b">Hello <strong>team`, "Hello team"},
	} {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, commentPreview(test.input))
		})
	}
}
