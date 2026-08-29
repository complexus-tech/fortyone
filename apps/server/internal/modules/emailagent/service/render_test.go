package emailagent

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderHTMLAllowsOnlyEscapedSafePrimitives(t *testing.T) {
	t.Parallel()

	blocks := []CopyBlock{
		{Kind: CopyBlockParagraph, Text: `Maya says <script>alert("x")</script> & continue.`},
		{Kind: CopyBlockBulletList, Text: "Changes", Items: []string{`Status: "Done"`, "Owner: Jo & Mo"}},
		{Kind: CopyBlockCallout, Text: "Reply CONFIRM\nto apply."},
	}
	copy := EmailCopy{
		Subject:   "A safe update",
		PlainText: RenderPlainText(blocks),
		Blocks:    blocks,
	}

	output, err := RenderHTML(copy)

	require.NoError(t, err)
	require.NotContains(t, output, "<script>")
	require.Contains(t, output, "&lt;script&gt;")
	require.Contains(t, output, "Jo &amp; Mo")
	require.Contains(t, output, "<ul><li>")
	require.Contains(t, output, "CONFIRM<br>to apply")
	require.NotContains(t, output, "href=")
}

func TestRenderHTMLRejectsURLsHeaderInjectionAndDrift(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		copy EmailCopy
	}{
		{
			name: "url",
			copy: EmailCopy{
				Subject: "Update",
				Blocks:  []CopyBlock{{Kind: CopyBlockParagraph, Text: "Visit https://evil.example"}},
			},
		},
		{
			name: "header injection",
			copy: EmailCopy{
				Subject: "Update\r\nBcc: victim@example.com",
				Blocks:  []CopyBlock{{Kind: CopyBlockParagraph, Text: "Safe body"}},
			},
		},
		{
			name: "plain text drift",
			copy: EmailCopy{
				Subject:   "Update",
				PlainText: "Different facts",
				Blocks:    []CopyBlock{{Kind: CopyBlockParagraph, Text: "Safe body"}},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if test.copy.PlainText == "" {
				test.copy.PlainText = RenderPlainText(test.copy.Blocks)
			}
			_, err := RenderHTML(test.copy)
			require.ErrorIs(t, err, ErrInvalidDecision)
		})
	}
}
