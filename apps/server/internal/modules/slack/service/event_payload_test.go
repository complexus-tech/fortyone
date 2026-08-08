package slack

import "testing"

func TestDecodeSlackEventRejectsTrailingContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{
			name: "multiple JSON values",
			body: `{"type":"event_callback"} {"type":"event_callback"}`,
		},
		{
			name: "malformed trailing bytes",
			body: `{"type":"event_callback"} not-json`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := decodeSlackEvent([]byte(test.body)); err == nil {
				t.Fatal("decodeSlackEvent() error = nil, want trailing-content error")
			}
		})
	}
}

func TestNormalizeSlackEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
		kind slackEventKind
		ok   bool
	}{
		{
			name: "app mention",
			body: `{"type":"event_callback","team_id":"T1","event_id":"Ev1","event":{"type":"app_mention","user":"U1","channel":"C1","ts":"10.1","text":"<@B1> status"}}`,
			kind: slackEventKindMention,
			ok:   true,
		},
		{
			name: "direct message",
			body: `{"type":"event_callback","team_id":"T1","event_id":"Ev2","event":{"type":"message","channel_type":"im","user":"U1","channel":"D1","ts":"10.2","text":"my tasks"}}`,
			kind: slackEventKindDirect,
			ok:   true,
		},
		{
			name: "public unmentioned message",
			body: `{"type":"event_callback","team_id":"T1","event_id":"Ev3","event":{"type":"message","channel_type":"channel","user":"U1","channel":"C1","ts":"10.3","text":"hello"}}`,
			ok:   false,
		},
		{
			name: "bot loop",
			body: `{"type":"event_callback","team_id":"T1","event_id":"Ev4","event":{"type":"message","channel_type":"im","bot_id":"B1","channel":"D1","ts":"10.4","text":"hello"}}`,
			ok:   false,
		},
		{
			name: "edited message",
			body: `{"type":"event_callback","team_id":"T1","event_id":"Ev5","event":{"type":"message","subtype":"message_changed","channel_type":"im","user":"U1","channel":"D1","ts":"10.5","text":"hello"}}`,
			ok:   false,
		},
		{
			name: "uninstall",
			body: `{"type":"event_callback","team_id":"T1","event_id":"Ev6","event":{"type":"app_uninstalled"}}`,
			kind: slackEventKindUninstalled,
			ok:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			envelope, err := decodeSlackEvent([]byte(test.body))
			if err != nil {
				t.Fatalf("decodeSlackEvent() error = %v", err)
			}
			event, ok := normalizeSlackEvent(envelope)
			if ok != test.ok {
				t.Fatalf("normalizeSlackEvent() ok = %v, want %v", ok, test.ok)
			}
			if ok && event.Kind != test.kind {
				t.Fatalf("normalizeSlackEvent() kind = %q, want %q", event.Kind, test.kind)
			}
		})
	}
}

func TestRemoveBotMention(t *testing.T) {
	t.Parallel()
	if got := removeBotMention(" <@B123>   what is due? ", "B123"); got != "what is due?" {
		t.Fatalf("removeBotMention() = %q", got)
	}
}
