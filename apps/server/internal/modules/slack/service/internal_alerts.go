package slack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/pkg/logger"
)

// InternalAlertConfig selects one operator-controlled destination independently
// of every source customer workspace. Disabled workers never claim alerts.
type InternalAlertConfig struct {
	Enabled   bool   `default:"true" env:"APP_INTERNAL_SLACK_ENABLED"`
	TeamID    string `env:"APP_INTERNAL_SLACK_TEAM_ID"`
	ChannelID string `env:"APP_INTERNAL_SLACK_CHANNEL_ID"`
}

var internalSlackTextEscaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
var internalSlackInvoiceIDPattern = regexp.MustCompile(`^in_[A-Za-z0-9]+$`)
var internalSlackCurrencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)
var internalSlackTeamIDPattern = regexp.MustCompile(`^T[A-Z0-9]+$`)
var internalSlackChannelIDPattern = regexp.MustCompile(`^C[A-Z0-9]+$`)

func (c InternalAlertConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if !internalSlackTeamIDPattern.MatchString(c.TeamID) || !internalSlackChannelIDPattern.MatchString(c.ChannelID) {
		return errors.New("internal Slack alerts require APP_INTERNAL_SLACK_TEAM_ID and APP_INTERNAL_SLACK_CHANNEL_ID")
	}
	return nil
}

type InternalAlertStore interface {
	ClaimInternalAlert(context.Context, string, string) (*slackdomain.InternalAlert, error)
	CompleteInternalAlert(context.Context, slackdomain.InternalAlert, string) error
	RetryInternalAlert(context.Context, slackdomain.InternalAlert, time.Time) error
}

type internalAlertInstallationStore interface {
	GetSlackWorkspaceByTeamID(context.Context, string) (slackdomain.Installation, error)
}

type InternalAlertDispatcher struct {
	cfg           InternalAlertConfig
	store         InternalAlertStore
	installations internalAlertInstallationStore
	codec         *credentialCodec
	client        *slackWebClient
	log           *logger.Logger
}

func NewInternalAlertDispatcher(cfg InternalAlertConfig, store InternalAlertStore, installations internalAlertInstallationStore, vault CredentialVault, log *logger.Logger) (*InternalAlertDispatcher, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if store == nil || installations == nil || log == nil {
		return nil, errors.New("internal Slack alert dependencies are required")
	}
	codec, err := newCredentialCodec(vault)
	if err != nil {
		return nil, err
	}
	return &InternalAlertDispatcher{cfg: cfg, store: store, installations: installations, codec: codec, client: newSlackWebClient(nil), log: log}, nil
}

// Dispatch handles one alert per scheduler tick, keeping traffic comfortably
// below Slack's per-channel message limit. Database leases fence concurrent workers.
func (d *InternalAlertDispatcher) Dispatch(ctx context.Context) error {
	if !d.cfg.Enabled {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()
	alert, err := d.store.ClaimInternalAlert(ctx, d.cfg.TeamID, d.cfg.ChannelID)
	if err != nil || alert == nil {
		return err
	}
	messageTS, sendErr := d.send(ctx, *alert)
	// Persist the outcome even if the task's request was cancelled after Slack
	// accepted it. Never store customer payloads or credentials in error logs.
	stateCtx, stateCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer stateCancel()
	if sendErr == nil {
		return d.store.CompleteInternalAlert(stateCtx, *alert, messageTS)
	}
	delay := time.Minute * time.Duration(1<<min(max(alert.Attempts-1, 0), 6))
	if retryAfter, limited := SlackRetryAfter(sendErr); limited && retryAfter > delay {
		delay = retryAfter
	}
	d.log.Error(ctx, "internal Slack alert delivery failed", "alert_id", alert.ID, "kind", alert.Kind, "attempt", alert.Attempts, "error", sendErr.Error(), "provider_code", internalAlertProviderCode(sendErr))
	retryErr := d.store.RetryInternalAlert(stateCtx, *alert, time.Now().UTC().Add(delay))
	return errors.Join(errors.New("internal Slack alert delivery failed; retry scheduled"), retryErr)
}

// The public error contains only the failing stage. The cause remains available
// for retry classification without leaking a credential or provider response.
type internalAlertSendError struct {
	stage string
	cause error
}

func (e *internalAlertSendError) Error() string {
	return "internal Slack alert failed during " + e.stage
}
func (e *internalAlertSendError) Unwrap() error { return e.cause }

func internalAlertProviderCode(err error) string {
	if _, limited := SlackRetryAfter(err); limited {
		return "ratelimited"
	}
	code, _ := SlackAPIErrorCode(err)
	switch code {
	case "not_authed", "invalid_auth", "account_inactive", "token_expired", "token_revoked", "missing_scope", "channel_not_found", "not_in_channel", "is_archived", "restricted_action":
		return code
	default:
		return "unavailable"
	}
}

func (d *InternalAlertDispatcher) send(ctx context.Context, alert slackdomain.InternalAlert) (messageTS string, err error) {
	stage := "message rendering"
	defer func() {
		if err != nil {
			err = &internalAlertSendError{stage: stage, cause: err}
		}
	}()
	text, err := renderInternalAlert(alert)
	if err != nil {
		return "", err
	}
	stage = "installation lookup"
	installation, err := d.installations.GetSlackWorkspaceByTeamID(ctx, d.cfg.TeamID)
	if err != nil {
		return "", err
	}
	if !installation.IsActive || installation.SlackTeamID != d.cfg.TeamID || installation.CredentialVersion != credentialvault.CurrentVersion {
		return "", errors.New("internal Slack installation is unavailable")
	}
	stage = "credential decryption"
	credential, version, err := d.codec.open(slackCredentialBinding{
		WorkspaceID: installation.WorkspaceID, SlackTeamID: installation.SlackTeamID, InstallGeneration: installation.InstallGeneration,
	}, installation.BotAccessToken)
	if err != nil {
		return "", err
	}
	if version != installation.CredentialVersion {
		return "", errors.New("internal Slack credential version mismatch")
	}
	stage = "team verification"
	var auth struct {
		TeamID string `json:"team_id"`
	}
	if err := d.client.callJSON(ctx, credential.AccessToken, "auth.test", nil, &auth); err != nil {
		return "", err
	}
	if auth.TeamID != d.cfg.TeamID {
		return "", errors.New("internal Slack token belongs to another team")
	}
	stage = "general channel verification"
	var info struct {
		Channel struct {
			ID          string `json:"id"`
			IsGeneral   bool   `json:"is_general"`
			IsArchived  bool   `json:"is_archived"`
			IsShared    bool   `json:"is_shared"`
			IsExtShared bool   `json:"is_ext_shared"`
		} `json:"channel"`
	}
	if err := d.client.callJSON(ctx, credential.AccessToken, "conversations.info?channel="+d.cfg.ChannelID, nil, &info); err != nil {
		return "", err
	}
	if info.Channel.ID != d.cfg.ChannelID || !info.Channel.IsGeneral || info.Channel.IsArchived || info.Channel.IsShared || info.Channel.IsExtShared {
		return "", errors.New("internal Slack destination must be the unshared general channel")
	}
	// Recheck disconnect/reinstall immediately before sending with this token.
	stage = "installation recheck"
	current, err := d.installations.GetSlackWorkspaceByTeamID(ctx, d.cfg.TeamID)
	if err != nil {
		return "", err
	}
	if !current.IsActive || current.InstallGeneration != installation.InstallGeneration || current.WorkspaceID != installation.WorkspaceID {
		return "", errSlackInstallationChanged
	}
	unfurl := false
	sender := slackAPISender{client: d.client}
	stage = "message posting"
	messageTS, err = sender.Send(ctx, credential.AccessToken, SlackOutboundMessage{
		ChannelID: d.cfg.ChannelID, Text: internalSlackTextEscaper.Replace(text), ClientMessageID: deterministicSlackMessageID("internal:" + alert.DedupeKey),
		ProviderPayload: SlackProviderPayload{
			Blocks:      []SlackBlock{{Type: "section", Text: &SlackTextObject{Type: "plain_text", Text: text}}},
			UnfurlLinks: &unfurl, UnfurlMedia: &unfurl,
		},
	})
	if err == nil && messageTS == "" {
		return "", errors.New("Slack returned no internal alert message ID")
	}
	return messageTS, err
}

type internalAlertPayload struct {
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	OccurredAt    time.Time `json:"occurred_at"`
	WorkspaceName string    `json:"workspace_name"`
	WorkspaceSlug string    `json:"workspace_slug"`
	InvoiceID     string    `json:"invoice_id"`
	AmountMinor   int64     `json:"amount_minor"`
	Currency      string    `json:"currency"`
}

func renderInternalAlert(alert slackdomain.InternalAlert) (string, error) {
	var payload internalAlertPayload
	if err := json.Unmarshal(alert.Payload, &payload); err != nil {
		return "", errors.New("invalid internal Slack alert payload")
	}
	if payload.OccurredAt.IsZero() {
		return "", errors.New("internal Slack alert timestamp is missing")
	}
	details := fmt.Sprintf("Name: %s\nEmail: %s\nTime: %s", internalAlertField(payload.Name), internalAlertField(payload.Email), payload.OccurredAt.UTC().Format("02 Jan 2006, 15:04 UTC"))
	switch alert.Kind {
	case "account_created":
		return "🎉 New FortyOne account\n" + details, nil
	case "payment_received":
		amount, err := internalAlertAmount(payload.AmountMinor, payload.Currency)
		if err != nil {
			return "", err
		}
		if !internalSlackInvoiceIDPattern.MatchString(payload.InvoiceID) {
			return "", errors.New("invalid internal payment invoice ID")
		}
		return fmt.Sprintf("💰 Payment received — %s\n%s\nWorkspace: %s (%s)\nInvoice: https://dashboard.stripe.com/invoices/%s", amount, details, internalAlertField(payload.WorkspaceName), internalAlertField(payload.WorkspaceSlug), payload.InvoiceID), nil
	default:
		return "", errors.New("unsupported internal Slack alert kind")
	}
}

func internalAlertField(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	if value == "" {
		return "Not provided"
	}
	runes := []rune(value)
	return string(runes[:min(len(runes), 250)])
}

// Stripe uses two decimal places except its documented zero- and three-decimal
// currencies. ISK and UGX retain Stripe's two-decimal API representation.
func internalAlertAmount(amount int64, currency string) (string, error) {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if amount <= 0 || !internalSlackCurrencyPattern.MatchString(currency) {
		return "", errors.New("invalid internal payment amount")
	}
	decimals, divisor := 2, int64(100)
	switch currency {
	case "BIF", "CLP", "DJF", "GNF", "JPY", "KMF", "KRW", "MGA", "PYG", "RWF", "VND", "VUV", "XAF", "XOF", "XPF":
		decimals, divisor = 0, 1
	case "BHD", "JOD", "KWD", "OMR", "TND":
		decimals, divisor = 3, 1000
	}
	if decimals == 0 {
		return fmt.Sprintf("%s %d", currency, amount), nil
	}
	return fmt.Sprintf("%s %d.%0*d", currency, amount/divisor, decimals, amount%divisor), nil
}
