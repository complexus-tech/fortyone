package chatrepository

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPlainLinkResourcesMatchRenderedLinks(t *testing.T) {
	for _, test := range []struct {
		name, text string
		want       []string
	}{
		{"protocols and inference", "https://fortyone.app/docs http://localhost:4177 mailto:maya@example.com www.fortyone.app product.fortyone.app hello@example.com", []string{"https://fortyone.app/docs", "http://localhost:4177/", "mailto:maya@example.com", "https://www.fortyone.app/", "https://product.fortyone.app/", "mailto:hello@example.com"}},
		{"punctuation and balancing", "(https://fortyone.app/docs_(chat)). [www.fortyone.app/guide]! https://fortyone.app/foo(bar)))", []string{"https://fortyone.app/docs_(chat)", "https://www.fortyone.app/guide", "https://fortyone.app/foo(bar)"}},
		{"code span", "`https://fortyone.app/code` and ``www.fortyone.app/also-code`` then https://fortyone.app/visible", []string{"https://fortyone.app/visible"}},
		{"fenced code", "```text\nhttps://fortyone.app/code\n```\nhttps://fortyone.app/visible", []string{"https://fortyone.app/visible"}},
		{"unfinished fence", "Visible https://fortyone.app/visible\n~~~\nwww.fortyone.app/code\nhello@example.com", []string{"https://fortyone.app/visible"}},
		{"unsafe and credentials", "ftp://fortyone.app/file javascript:https://fortyone.app/private data:text/plain,https://fortyone.app/hidden https://user:secret@fortyone.app/ https://fortyone.app/safe", []string{"https://fortyone.app/safe"}},
		{"unknown suffix and dedupe", "artifact.notarealtld https://FORTYONE.APP https://fortyone.app/ FORTYONE.APP https://fortyone.app:443/", []string{"https://fortyone.app/"}},
		{"inferred ports and private suffix", "www.fortyone.app:4177/docs project.github.io hello+chat@example.com", []string{"https://www.fortyone.app:4177/docs", "https://project.github.io/", "mailto:hello+chat@example.com"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, extractLinks(test.text, nil))
		})
	}
}

func TestRichLinkResourcesExcludeCodeAndTreatExplicitLinksAsAuthoritative(t *testing.T) {
	rich := json.RawMessage(`{"type":"doc","content":[
	 {"type":"codeBlock","content":[{"type":"text","text":"https://fortyone.app/block"}]},
	 {"type":"paragraph","content":[
	  {"type":"text","text":"https://fortyone.app/code","marks":[{"type":"code"}]},
	  {"type":"text","text":"https://fortyone.app/label","marks":[{"type":"link","attrs":{"href":"https://fortyone.app/destination"}}]},
	  {"type":"text","text":"https://fortyone.app/unsafe-label","marks":[{"type":"link","attrs":{"href":"javascript:alert(1)"}}]},
	  {"type":"text","text":"www.fortyone.app/plain"},
	  {"type":"text","text":"Contact Maya","marks":[{"type":"link","attrs":{"href":"mailto:maya@example.com?subject=Chat"}}]}
	 ]}
	]}`)
	require.Equal(t, []string{"https://fortyone.app/destination", "https://www.fortyone.app/plain", "mailto:maya@example.com?subject=Chat"}, extractLinks("https://fortyone.app/mirror", rich))
}

func TestLinkResourceExtractionIsBoundedAndFallsBackForLegacyPlainContent(t *testing.T) {
	var input strings.Builder
	for i := 0; i < 75; i++ {
		fmt.Fprintf(&input, "https://fortyone.app/page/%d ", i)
	}
	require.Len(t, extractLinks(input.String(), nil), maximumMessageLinks)
	require.Equal(t, []string{"https://fortyone.app/"}, extractLinks("fortyone.app", json.RawMessage(`{"legacy":true}`)))
	require.Empty(t, extractLinks("https://fortyone.app/"+strings.Repeat("x", 2048), nil))
}
