package slack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type internalAlertStoreStub struct {
	alert          *slackdomain.InternalAlert
	claimedTeam    string
	claimedChannel string
	completed      string
	retryAt        time.Time
	completeErr    error
}

func (s *internalAlertStoreStub) ClaimInternalAlert(_ context.Context, team, channel string) (*slackdomain.InternalAlert, error) {
	s.claimedTeam, s.claimedChannel = team, channel
	return s.alert, nil
}
func (s *internalAlertStoreStub) CompleteInternalAlert(_ context.Context, _ slackdomain.InternalAlert, ts string) error {
	s.completed = ts
	if s.completeErr == nil {
		s.alert = nil
	}
	return s.completeErr
}
func (s *internalAlertStoreStub) RetryInternalAlert(_ context.Context, _ slackdomain.InternalAlert, at time.Time) error {
	s.retryAt = at
	return nil
}

type internalAlertInstallationStub struct {
	installation        slackdomain.Installation
	calls               int
	disconnectOnRecheck bool
}

func (s *internalAlertInstallationStub) GetSlackWorkspaceByTeamID(_ context.Context, team string) (slackdomain.Installation, error) {
	s.calls++
	if team != s.installation.SlackTeamID {
		return slackdomain.Installation{}, errors.New("unexpected team lookup")
	}
	result := s.installation
	if s.disconnectOnRecheck && s.calls > 1 {
		result.IsActive = false
	}
	return result, nil
}

func TestInternalAlertsUseOnlyConfiguredGeneralAndExistingInstallation(t *testing.T) {
	for _, scenario := range []string{"success", "wrong_team", "wrong_channel", "not_general", "shared", "archived", "disconnected", "rate_limited", "receipt_failure", "disabled"} {
		t.Run(scenario, func(t *testing.T) {
			cfg := InternalAlertConfig{Enabled: scenario != "disabled", TeamID: "T015B85FC6R", ChannelID: "C014XSVSRF1"}
			alert := &slackdomain.InternalAlert{ID: uuid.New(), LeaseToken: uuid.New(), Kind: "account_created", DedupeKey: "account_created:" + uuid.NewString(), Attempts: 1,
				Payload: json.RawMessage(`{"name":"A <!channel>","email":"person@example.test","occurred_at":"2026-09-13T10:00:00Z"}`)}
			store := &internalAlertStoreStub{alert: alert}
			if scenario == "receipt_failure" {
				store.completeErr = slackdomain.ErrInternalAlertLeaseLost
			}
			vault := newTestCredentialVault(t)
			codec, err := newCredentialCodec(vault)
			require.NoError(t, err)
			installation := slackdomain.Installation{WorkspaceID: uuid.New(), SlackTeamID: cfg.TeamID, InstallGeneration: uuid.New(), IsActive: true}
			installation.BotAccessToken, installation.CredentialVersion, err = codec.seal(slackCredentialBinding{WorkspaceID: installation.WorkspaceID, SlackTeamID: cfg.TeamID, InstallGeneration: installation.InstallGeneration}, slackCredential{AccessToken: "test-internal-token"})
			require.NoError(t, err)
			installations := &internalAlertInstallationStub{installation: installation, disconnectOnRecheck: scenario == "disconnected"}
			posts := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "Bearer test-internal-token", r.Header.Get("Authorization"))
				w.Header().Set("Content-Type", "application/json")
				switch r.URL.Path {
				case "/auth.test":
					team := cfg.TeamID
					if scenario == "wrong_team" {
						team = "TOTHER"
					}
					_, _ = fmt.Fprintf(w, `{"ok":true,"team_id":%q}`, team)
				case "/conversations.info":
					require.Equal(t, http.MethodGet, r.Method)
					require.Equal(t, cfg.ChannelID, r.URL.Query().Get("channel"))
					channel := cfg.ChannelID
					if scenario == "wrong_channel" {
						channel = "COTHER"
					}
					_, _ = fmt.Fprintf(w, `{"ok":true,"channel":{"id":%q,"is_general":%t,"is_shared":%t,"is_archived":%t}}`, channel, scenario != "not_general", scenario == "shared", scenario == "archived")
				case "/chat.postMessage":
					posts++
					var body map[string]any
					require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
					require.Equal(t, cfg.ChannelID, body["channel"])
					require.Equal(t, deterministicSlackMessageID("internal:"+alert.DedupeKey), body["client_msg_id"])
					require.NotContains(t, body["text"], "<!channel>")
					require.Equal(t, false, body["unfurl_links"])
					blocks := body["blocks"].([]any)
					require.Equal(t, "plain_text", blocks[0].(map[string]any)["text"].(map[string]any)["type"])
					if scenario == "rate_limited" {
						w.Header().Set("Retry-After", "180")
						w.WriteHeader(http.StatusTooManyRequests)
						return
					}
					_, _ = io.WriteString(w, `{"ok":true,"ts":"171.234"}`)
				default:
					t.Errorf("unexpected Slack endpoint %s", r.URL.Path)
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			dispatcher, err := NewInternalAlertDispatcher(cfg, store, installations, vault, logger.NewWithText(io.Discard, slog.LevelError, "internal-alert-test"))
			require.NoError(t, err)
			dispatcher.client = newSlackWebClient(server.Client())
			dispatcher.client.baseURL = server.URL
			before := time.Now()
			err = dispatcher.Dispatch(t.Context())
			switch scenario {
			case "success":
				require.NoError(t, err)
				require.Equal(t, "171.234", store.completed)
				require.Equal(t, 1, posts)
				require.NoError(t, dispatcher.Dispatch(t.Context()))
				require.Equal(t, 1, posts)
			case "disabled":
				require.NoError(t, err)
				require.Empty(t, store.claimedTeam)
				require.Zero(t, posts)
			case "receipt_failure":
				require.ErrorIs(t, err, slackdomain.ErrInternalAlertLeaseLost)
				require.Equal(t, 1, posts)
			default:
				require.Error(t, err)
				require.Empty(t, store.completed)
				require.True(t, store.retryAt.After(before.Add(time.Minute)))
				if scenario == "rate_limited" {
					require.Equal(t, 1, posts)
					require.True(t, store.retryAt.After(before.Add(180*time.Second)))
				} else {
					require.Zero(t, posts)
				}
			}
			if cfg.Enabled {
				require.Equal(t, cfg.TeamID, store.claimedTeam)
				require.Equal(t, cfg.ChannelID, store.claimedChannel)
			}
		})
	}
}

func TestInternalAlertContentAndAmounts(t *testing.T) {
	for _, tc := range []struct {
		currency string
		amount   int64
		want     string
	}{
		{"usd", 12345, "USD 123.45"}, {"jpy", 12345, "JPY 12345"}, {"kwd", 12345, "KWD 12.345"}, {"ugx", 500, "UGX 5.00"}, {"isk", 500, "ISK 5.00"},
	} {
		got, err := internalAlertAmount(tc.amount, tc.currency)
		require.NoError(t, err)
		require.Equal(t, tc.want, got)
	}
	_, err := internalAlertAmount(0, "usd")
	require.Error(t, err)
	_, err = internalAlertAmount(100, "<!channel>")
	require.Error(t, err)
	text, err := renderInternalAlert(slackdomain.InternalAlert{Kind: "payment_received", Payload: json.RawMessage(`{"name":"Buyer","email":"buyer@example.test","amount_minor":4900,"currency":"usd","workspace_name":"Customer","workspace_slug":"customer","invoice_id":"in_123","occurred_at":"2026-09-13T10:00:00Z"}`)})
	require.NoError(t, err)
	require.Contains(t, text, "USD 49.00")
	require.Contains(t, text, "Workspace: Customer (customer)")
	require.Contains(t, text, "https://dashboard.stripe.com/invoices/in_123")
	_, err = renderInternalAlert(slackdomain.InternalAlert{Kind: "workspace_created", Payload: json.RawMessage(`{"occurred_at":"2026-09-13T10:00:00Z"}`)})
	require.Error(t, err)
}
