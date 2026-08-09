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
		name     string
		body     string
		kind     slackEventKind
		threadTS string
		replyTS  string
		ok       bool
	}{
		{
			name:     "app mention",
			body:     `{"type":"event_callback","team_id":"T1","event_id":"Ev1","event":{"type":"app_mention","user":"U1","channel":"C1","ts":"10.1","text":"<@B1> status"}}`,
			kind:     slackEventKindMention,
			threadTS: "10.1",
			replyTS:  "10.1",
			ok:       true,
		},
		{
			name:     "direct message",
			body:     `{"type":"event_callback","team_id":"T1","event_id":"Ev2","event":{"type":"message","channel_type":"im","user":"U1","channel":"D1","ts":"10.2","text":"my tasks"}}`,
			kind:     slackEventKindDirect,
			threadTS: "10.2",
			ok:       true,
		},
		{
			name: "public root message",
			body: `{"type":"event_callback","team_id":"T1","event_id":"Ev3","event":{"type":"message","channel_type":"channel","user":"U1","channel":"C1","ts":"10.3","text":"hello"}}`,
			ok:   false,
		},
		{
			name:     "public thread reply",
			body:     `{"type":"event_callback","team_id":"T1","event_id":"Ev3-thread","event":{"type":"message","channel_type":"channel","user":"U1","channel":"C1","ts":"10.4","thread_ts":"10.1","text":"what about urgent work?"}}`,
			kind:     slackEventKindChannelThread,
			threadTS: "10.1",
			replyTS:  "10.1",
			ok:       true,
		},
		{
			name:     "private thread reply",
			body:     `{"type":"event_callback","team_id":"T1","event_id":"Ev3-private","event":{"type":"message","channel_type":"group","user":"U1","channel":"G1","ts":"10.4","thread_ts":"10.1","text":"what about urgent work?"}}`,
			kind:     slackEventKindChannelThread,
			threadTS: "10.1",
			replyTS:  "10.1",
			ok:       true,
		},
		{
			name: "multiparty direct message",
			body: `{"type":"event_callback","team_id":"T1","event_id":"Ev3-mpim","event":{"type":"message","channel_type":"mpim","user":"U1","channel":"G1","ts":"10.4","thread_ts":"10.1","text":"hello"}}`,
			ok:   false,
		},
		{
			name: "Slack Connect mention",
			body: `{"type":"event_callback","team_id":"T1","event_id":"Ev3-connect","is_ext_shared_channel":true,"event":{"type":"app_mention","user":"U1","channel":"C1","ts":"10.4","text":"<@B1> hello"}}`,
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
			if ok && (event.ThreadTS != test.threadTS || event.ReplyTS != test.replyTS) {
				t.Fatalf("normalizeSlackEvent() thread/reply = %q/%q, want %q/%q", event.ThreadTS, event.ReplyTS, test.threadTS, test.replyTS)
			}
		})
	}
}

func TestContainsSlackUserMention(t *testing.T) {
	t.Parallel()

	if !containsSlackUserMention("please ask <@B123> for help", "B123") {
		t.Fatal("containsSlackUserMention() = false, want true")
	}
	if containsSlackUserMention("please ask <@B1234> for help", "B123") {
		t.Fatal("containsSlackUserMention() matched a different user")
	}
}

func TestRemoveBotMention(t *testing.T) {
	t.Parallel()
	if got := removeBotMention(" <@B123>   what is due? ", "B123"); got != "what is due?" {
		t.Fatalf("removeBotMention() = %q", got)
	}
}
