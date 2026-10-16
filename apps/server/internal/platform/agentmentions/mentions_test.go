package agentmentions

import (
	"testing"

	"github.com/google/uuid"
)

func TestValidateMaya(t *testing.T) {
	id := uuid.New()
	mention := `<a data-type="mention" data-mention-kind="agent" data-agent-key="maya" data-id="` + id.String() + `">@Maya</a>`
	for _, test := range []struct {
		name, content string
		agents        []string
		want          bool
		invalid       bool
	}{
		{"explicit", "<p>" + mention + " close this ticket</p>", []string{"maya"}, true, false},
		{"metadata required", mention, nil, false, false},
		{"plain text", "@Maya close it", []string{"maya"}, false, true},
		{"quoted", "<blockquote>" + mention + "</blockquote>", []string{"maya"}, false, true},
		{"code", "<pre><code>" + mention + "</code></pre>", []string{"maya"}, false, true},
		{"human name collision", `<a data-type="mention" data-id="` + uuid.NewString() + `">Maya</a>`, []string{"maya"}, false, true},
		{"unknown agent", mention, []string{"other"}, false, true},
		{"duplicate metadata", mention, []string{"maya", "maya"}, false, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := ValidateMaya(test.content, test.agents, id)
			if got != test.want || (err != nil) != test.invalid {
				t.Fatalf("got (%v, %v), want enabled=%v invalid=%v", got, err, test.want, test.invalid)
			}
		})
	}
}

func TestTextExcludesExecutableMarkup(t *testing.T) {
	got := Text(`<p>Hello <strong>Maya</strong></p><script>bad()</script><p>Next &amp; last</p>`)
	if got != "Hello Maya\nNext & last" {
		t.Fatalf("unexpected visible context: %q", got)
	}
}
