package agentreadinesshttp

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"

	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/golang-jwt/jwt/v5"
)

const (
	accessTokenTTL       = 15 * time.Minute
	authorizationCodeTTL = 5 * time.Minute
	refreshTokenTTL      = 30 * 24 * time.Hour
)

type oauthClient struct {
	ClientID     string   `json:"client_id"`
	ClientName   string   `json:"client_name"`
	RedirectURIs []string `json:"redirect_uris"`
}
type authorizationRequest struct {
	ClientID      string
	RedirectURI   string
	State         string
	Scope         string
	CodeChallenge string
	UserID        string
}
type authorizationCode struct {
	ClientID      string
	RedirectURI   string
	Scope         string
	CodeChallenge string
	UserID        string
}
type refreshGrant struct {
	ClientID string
	Scope    string
	UserID   string
}

type approvalPageData struct {
	ClientName  string
	Approval    string
	RedirectURI string
}

var approvalPage = template.Must(template.New("mcp-approval").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="color-scheme" content="light dark">
  <title>Connect to FortyOne</title>
  <style>
    :root {
      color-scheme: light;
      --radius-scale: 0.75;
      --radius-base-lg: 0.6rem;
      --radius-base-xl: 0.8rem;
      --radius-lg: calc(var(--radius-base-lg) * var(--radius-scale));
      --radius-xl: calc(var(--radius-base-xl) * var(--radius-scale));
      --font-body: -apple-system, BlinkMacSystemFont, "Inter", sans-serif;
      --background: oklch(0.997 0.002 95);
      --surface: oklch(0.997 0.002 95);
      --surface-muted: oklch(0.975 0.003 95);
      --foreground: oklch(0.145 0.006 95);
      --text-secondary: oklch(0.35 0.005 95);
      --color-border: oklch(0.92 0.004 95);
      --color-border-strong: oklch(0.89 0.005 95);
      --primary: oklch(0.6522 0.2135 38);
      --primary-foreground: oklch(0.1821 0.0139 94);
      --focus: oklch(0.6522 0.2135 38);
      --shadow: oklch(0.145 0.006 95 / 0.06);
    }

    @media (prefers-color-scheme: dark) {
      :root {
        color-scheme: dark;
        --background: oklch(0.1821 0.0139 94);
        --surface: oklch(0.2135 0.0118 91.8);
        --surface-muted: oklch(0.2351 0.0115 91.7);
        --foreground: oklch(0.9439 0.0011 17.2);
        --text-secondary: oklch(0.9439 0.0011 17.2 / 60%);
        --color-border: oklch(0.9439 0.0011 17.2 / 10%);
        --color-border-strong: oklch(0.9439 0.0011 17.2 / 16%);
        --primary: oklch(0.6522 0.2135 38);
        --primary-foreground: oklch(0.1821 0.0139 94);
        --focus: oklch(0.6522 0.2135 38);
        --shadow: oklch(0.1 0.01 94 / 35%);
      }
    }

    * {
      box-sizing: border-box;
    }

    html {
      min-height: 100%;
      background: var(--background);
    }

    body {
      min-height: 100vh;
      margin: 0;
      display: grid;
      place-items: center;
      padding: 24px;
      background: var(--background);
      color: var(--foreground);
      font-family: var(--font-body);
      -webkit-font-smoothing: antialiased;
      text-rendering: optimizeLegibility;
    }

    .shell {
      width: min(100%, 416px);
    }

    .brand {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 10px;
      margin-bottom: 24px;
      color: var(--foreground);
      font-size: 15px;
      font-weight: 650;
      letter-spacing: -0.01em;
    }

    .brand svg {
      width: 22px;
      height: 22px;
    }

    .card {
      padding: 32px;
      background: var(--surface);
      border: 1px solid var(--color-border);
      border-radius: var(--radius-xl);
      box-shadow: 0 16px 48px var(--shadow);
    }

    h1 {
      margin: 0;
      color: var(--foreground);
      font-size: clamp(24px, 5vw, 30px);
      font-weight: 650;
      letter-spacing: -0.04em;
      line-height: 1.15;
    }

    .description {
      margin: 16px 0 0;
      color: var(--text-secondary);
      font-size: 14px;
      line-height: 1.6;
    }

    .consent-note {
      margin: 18px 0 0;
      color: var(--text-secondary);
      font-size: 13px;
      line-height: 1.55;
    }

    .permissions {
      margin: 24px 0 0;
      padding: 0;
      list-style: none;
      border: 1px solid var(--color-border);
      border-radius: var(--radius-xl);
      overflow: hidden;
    }

    .permissions li {
      display: flex;
      gap: 12px;
      align-items: flex-start;
      padding: 14px;
      background: var(--surface-muted);
      color: var(--foreground);
      font-size: 14px;
      line-height: 1.45;
    }

    .permissions li + li {
      border-top: 1px solid var(--color-border);
    }

    .permissions svg {
      flex: 0 0 auto;
      width: 18px;
      height: 18px;
      margin-top: 1px;
      color: var(--text-secondary);
    }

    form {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 10px;
      margin-top: 24px;
    }

    button {
      min-height: 44px;
      padding: 10px 16px;
      border: 1px solid transparent;
      border-radius: var(--radius-lg);
      font: inherit;
      font-size: 14px;
      font-weight: 650;
      cursor: pointer;
      transition: background-color 140ms ease, border-color 140ms ease, transform 140ms ease;
    }

    button:focus-visible {
      outline: 3px solid var(--focus);
      outline-offset: 3px;
    }

    button:active {
      transform: translateY(1px);
    }

    .allow {
      background: var(--primary);
      color: var(--primary-foreground);
    }

    .allow:hover {
      background: color-mix(in oklch, var(--primary), var(--foreground) 8%);
    }

    .deny {
      background: var(--surface);
      color: var(--foreground);
      border-color: var(--color-border-strong);
    }

    .deny:hover {
      background: var(--surface-muted);
    }

    @supports (corner-shape: squircle) {
      :root {
        --radius-scale: 2.5;
      }

      .card,
      .permissions,
      button {
        corner-shape: squircle;
      }
    }

    @media (max-width: 520px) {
      body {
        padding: 16px;
      }

      .card {
        padding: 24px;
      }

      form {
        grid-template-columns: 1fr;
      }

      .allow {
        order: -1;
      }
    }

    @media (prefers-reduced-motion: reduce) {
      button {
        transition: none;
      }
    }
  </style>
</head>
<body>
  <main class="shell">
    <div class="brand" aria-label="FortyOne">
      <svg viewBox="0 0 607 587" aria-hidden="true">
        <path fill="currentColor" d="M490.12 0C544.803 0 572.144.0005 589.132 16.9883 606.12 33.9761 606.12 61.3172 606.12 116v365.375c0 44.669 0 67.004-11.649 82.49-3.271 4.348-7.137 8.215-11.486 11.486C567.499 587 545.164 587 500.495 587h-21.16c-24.647 0-36.972 0-46.81-3.685-15.701-5.881-28.089-18.27-33.97-33.971-3.685-9.838-3.686-22.162-3.686-46.809 0-13.173.001-19.76-1.969-25.018-3.143-8.393-9.765-15.014-18.157-18.158-5.258-1.969-11.845-1.969-25.018-1.969H217.036c-127.665 0-191.498 0-211.269-40.267-19.771-40.267 19.263-90.774 97.33-191.789L213.589 82.362c31.254-40.44 46.881-60.661 68.98-71.511C304.668.0001 330.223 0 381.333 0H490.12ZM374.526 130.627c-21.02-4.316-44.267 23.382-90.761 78.778l-8.733 10.405c-37.534 44.721-56.301 67.082-50.498 85.834.968 3.126 2.357 6.105 4.129 8.856 10.632 16.5 39.825 16.5 98.21 16.5h8.732c29.761 0 44.641 0 54.627-8.262a41.65 41.65 0 0 0 4.792-4.791c8.262-9.986 8.262-24.867 8.262-54.627v-10.405c0-72.322 0-108.483-18.876-118.69a30.753 30.753 0 0 0-9.884-3.598Z"/>
      </svg>
      <span>FortyOne</span>
    </div>

    <section class="card" aria-labelledby="approval-title">
      <h1 id="approval-title">{{.ClientName}} would like to connect to FortyOne</h1>
      <p class="description">This connection will be able to:</p>

      <ul class="permissions" aria-label="Connection permissions">
        <li>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" aria-hidden="true"><path d="M3 12s3.25-6 9-6 9 6 9 6-3.25 6-9 6-9-6-9-6Z"/><circle cx="12" cy="12" r="2.5"/></svg>
          <span>Access workspaces, teams, stories, sprints, objectives, and key results you can already access</span>
        </li>
        <li>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" aria-hidden="true"><path d="M12 3v18M3 12h18"/></svg>
          <span>Create and organize work in FortyOne after you approve each action</span>
        </li>
      </ul>

      <p class="consent-note">By allowing access, you let {{.ClientName}} read your accessible FortyOne data and request changes on your behalf. You stay in control: changes require your approval in the connected client.</p>

      <form method="post" action="/oauth/authorize">
        <input type="hidden" name="approval" value="{{.Approval}}">
        <button class="deny" type="submit" name="decision" value="deny">Cancel</button>
        <button class="allow" type="submit" name="decision" value="allow">Agree &amp; allow</button>
      </form>
    </section>
  </main>
</body>
</html>`))

func renderApprovalPage(w http.ResponseWriter, data approvalPageData) error {
	redirectOrigin, err := oauthRedirectOrigin(data.RedirectURI)
	if err != nil {
		return err
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Language", "en")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; form-action 'self' "+redirectOrigin+"; base-uri 'none'; frame-ancestors 'none'")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	return approvalPage.Execute(w, data)
}

func (h *Handler) ProtectedResourceMetadata(_ context.Context, w http.ResponseWriter, _ *http.Request) error {
	base := strings.TrimRight(h.cfg.APIPublicURL, "/")
	return writeJSON(w, http.StatusOK, map[string]any{"resource": h.resource, "authorization_servers": []string{base}, "scopes_supported": []string{mcpScope, "offline_access"}, "bearer_methods_supported": []string{"header"}})
}

func (h *Handler) AuthorizationServerMetadata(_ context.Context, w http.ResponseWriter, _ *http.Request) error {
	base := strings.TrimRight(h.cfg.APIPublicURL, "/")
	return writeJSON(w, http.StatusOK, map[string]any{"issuer": base, "authorization_endpoint": base + "/oauth/authorize", "token_endpoint": base + "/oauth/token", "registration_endpoint": base + "/oauth/register", "revocation_endpoint": base + "/oauth/revoke", "response_types_supported": []string{"code"}, "grant_types_supported": []string{"authorization_code", "refresh_token"}, "code_challenge_methods_supported": []string{"S256"}, "token_endpoint_auth_methods_supported": []string{"none"}, "authorization_response_iss_parameter_supported": true, "scopes_supported": []string{mcpScope, "offline_access"}})
}

func (h *Handler) RegisterClient(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var request struct {
		ClientName              string   `json:"client_name"`
		RedirectURIs            []string `json:"redirect_uris"`
		TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&request); err != nil {
		return h.oauthError(ctx, w, "register_client", http.StatusBadRequest, "invalid_client_metadata", "invalid JSON registration")
	}
	if request.TokenEndpointAuthMethod != "" && request.TokenEndpointAuthMethod != "none" {
		return h.oauthError(ctx, w, "register_client", http.StatusBadRequest, "invalid_client_metadata", "only public PKCE clients are supported")
	}
	if len(request.RedirectURIs) == 0 {
		return h.oauthError(ctx, w, "register_client", http.StatusBadRequest, "invalid_redirect_uri", "at least one redirect URI is required")
	}
	for _, redirectURI := range request.RedirectURIs {
		if err := validateRedirectURI(redirectURI); err != nil {
			return h.oauthError(ctx, w, "register_client", http.StatusBadRequest, "invalid_redirect_uri", err.Error())
		}
	}
	clientID, err := randomToken(24)
	if err != nil {
		return err
	}
	client := oauthClient{ClientID: clientID, ClientName: strings.TrimSpace(request.ClientName), RedirectURIs: request.RedirectURIs}
	if client.ClientName == "" {
		client.ClientName = "MCP client"
	}
	if err := h.cfg.Cache.Set(ctx, oauthKey("client", clientID), client, 365*24*time.Hour); err != nil {
		return err
	}
	return writeJSON(w, http.StatusCreated, map[string]any{"client_id": clientID, "client_name": client.ClientName, "redirect_uris": client.RedirectURIs, "token_endpoint_auth_method": "none", "grant_types": []string{"authorization_code", "refresh_token"}, "response_types": []string{"code"}})
}

func (h *Handler) Authorize(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodPost {
		return h.approveAuthorization(ctx, w, r)
	}
	query := r.URL.Query()
	if query.Get("response_type") != "code" {
		return h.oauthError(ctx, w, "authorize", http.StatusBadRequest, "unsupported_response_type", "response_type must be code")
	}
	client, err := h.loadClient(ctx, query.Get("client_id"))
	if err != nil {
		return h.oauthError(ctx, w, "authorize", http.StatusBadRequest, "invalid_request", "unknown client_id")
	}
	redirectURI := query.Get("redirect_uri")
	if !slicesContains(client.RedirectURIs, redirectURI) {
		return h.oauthError(ctx, w, "authorize", http.StatusBadRequest, "invalid_request", "redirect_uri is not registered")
	}
	if query.Get("code_challenge_method") != "S256" || query.Get("code_challenge") == "" {
		return redirectOAuthError(w, r, redirectURI, query.Get("state"), "invalid_request", "PKCE S256 is required")
	}
	if resource := query.Get("resource"); resource != h.resource {
		return redirectOAuthError(w, r, redirectURI, query.Get("state"), "invalid_target", "resource does not match the FortyOne MCP endpoint")
	}
	userID, ok, err := mid.ResolveSessionUserID(ctx, r)
	if err != nil {
		return err
	}
	if !ok {
		callback := absoluteRequestURL(h.cfg.APIPublicURL, r)
		http.Redirect(w, r, strings.TrimRight(h.cfg.LoginURL, "/")+"/?callbackUrl="+url.QueryEscape(callback), http.StatusFound)
		return nil
	}
	scope := normalizeScope(query.Get("scope"))
	approval, err := randomToken(24)
	if err != nil {
		return err
	}
	pending := authorizationRequest{ClientID: client.ClientID, RedirectURI: redirectURI, State: query.Get("state"), Scope: scope, CodeChallenge: query.Get("code_challenge"), UserID: userID.String()}
	if err := h.cfg.Cache.Set(ctx, oauthKey("approval", approval), pending, authorizationCodeTTL); err != nil {
		return err
	}
	return renderApprovalPage(w, approvalPageData{ClientName: client.ClientName, Approval: approval, RedirectURI: pending.RedirectURI})
}

func (h *Handler) approveAuthorization(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return h.oauthError(ctx, w, "approve_authorization", http.StatusBadRequest, "invalid_request", "invalid approval")
	}
	approval := r.Form.Get("approval")
	var pending authorizationRequest
	if err := h.cfg.Cache.Get(ctx, oauthKey("approval", approval), &pending); err != nil {
		return h.oauthError(ctx, w, "approve_authorization", http.StatusBadRequest, "invalid_request", "approval expired")
	}
	_ = h.cfg.Cache.Delete(ctx, oauthKey("approval", approval))
	userID, ok, err := mid.ResolveSessionUserID(ctx, r)
	if err != nil {
		return err
	}
	if !ok || userID.String() != pending.UserID {
		return h.oauthError(ctx, w, "approve_authorization", http.StatusUnauthorized, "access_denied", "session is missing or changed")
	}
	if r.Form.Get("decision") != "allow" {
		return redirectOAuthError(w, r, pending.RedirectURI, pending.State, "access_denied", "the user declined access")
	}
	code, err := randomToken(32)
	if err != nil {
		return err
	}
	record := authorizationCode{ClientID: pending.ClientID, RedirectURI: pending.RedirectURI, Scope: pending.Scope, CodeChallenge: pending.CodeChallenge, UserID: pending.UserID}
	if err := h.cfg.Cache.Set(ctx, oauthKey("code", hashToken(code)), record, authorizationCodeTTL); err != nil {
		return err
	}
	destination, _ := url.Parse(pending.RedirectURI)
	values := destination.Query()
	values.Set("code", code)
	if pending.State != "" {
		values.Set("state", pending.State)
	}
	values.Set("iss", strings.TrimRight(h.cfg.APIPublicURL, "/"))
	destination.RawQuery = values.Encode()
	http.Redirect(w, r, destination.String(), http.StatusFound)
	return nil
}

func (h *Handler) Token(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return h.oauthError(ctx, w, "token", http.StatusBadRequest, "invalid_request", "invalid form")
	}
	switch r.Form.Get("grant_type") {
	case "authorization_code":
		return h.exchangeCode(ctx, w, r)
	case "refresh_token":
		return h.exchangeRefreshToken(ctx, w, r)
	default:
		return h.oauthError(ctx, w, "token", http.StatusBadRequest, "unsupported_grant_type", "grant_type is not supported")
	}
}

func (h *Handler) exchangeCode(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if r.Form.Get("resource") != h.resource {
		return h.oauthError(ctx, w, "exchange_code", http.StatusBadRequest, "invalid_target", "resource must match the FortyOne MCP endpoint")
	}
	code := r.Form.Get("code")
	key := oauthKey("code", hashToken(code))
	var record authorizationCode
	if err := h.cfg.Cache.Get(ctx, key, &record); err != nil {
		return h.oauthError(ctx, w, "exchange_code", http.StatusBadRequest, "invalid_grant", "authorization code is invalid or expired")
	}
	_ = h.cfg.Cache.Delete(ctx, key)
	if record.ClientID != r.Form.Get("client_id") || record.RedirectURI != r.Form.Get("redirect_uri") {
		return h.oauthError(ctx, w, "exchange_code", http.StatusBadRequest, "invalid_grant", "authorization code binding does not match")
	}
	digest := sha256.Sum256([]byte(r.Form.Get("code_verifier")))
	if base64.RawURLEncoding.EncodeToString(digest[:]) != record.CodeChallenge {
		return h.oauthError(ctx, w, "exchange_code", http.StatusBadRequest, "invalid_grant", "PKCE verification failed")
	}
	return h.issueTokens(ctx, w, record.UserID, record.ClientID, record.Scope)
}

func (h *Handler) exchangeRefreshToken(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if r.Form.Get("resource") != h.resource {
		return h.oauthError(ctx, w, "exchange_refresh_token", http.StatusBadRequest, "invalid_target", "resource must match the FortyOne MCP endpoint")
	}
	raw := r.Form.Get("refresh_token")
	key := oauthKey("refresh", hashToken(raw))
	var grant refreshGrant
	if err := h.cfg.Cache.Get(ctx, key, &grant); err != nil {
		return h.oauthError(ctx, w, "exchange_refresh_token", http.StatusBadRequest, "invalid_grant", "refresh token is invalid or expired")
	}
	_ = h.cfg.Cache.Delete(ctx, key)
	if grant.ClientID != r.Form.Get("client_id") {
		return h.oauthError(ctx, w, "exchange_refresh_token", http.StatusBadRequest, "invalid_grant", "refresh token client does not match")
	}
	return h.issueTokens(ctx, w, grant.UserID, grant.ClientID, grant.Scope)
}

func (h *Handler) issueTokens(ctx context.Context, w http.ResponseWriter, userID, clientID, scope string) error {
	now := time.Now().UTC()
	claims := mcpClaims{RegisteredClaims: jwt.RegisteredClaims{Issuer: strings.TrimRight(h.cfg.APIPublicURL, "/"), Subject: userID, Audience: jwt.ClaimStrings{h.resource}, IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenTTL))}, Scope: scope}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(h.signingKey())
	if err != nil {
		return err
	}
	refreshToken, err := randomToken(48)
	if err != nil {
		return err
	}
	if err := h.cfg.Cache.Set(ctx, oauthKey("refresh", hashToken(refreshToken)), refreshGrant{ClientID: clientID, Scope: scope, UserID: userID}, refreshTokenTTL); err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, map[string]any{"access_token": accessToken, "token_type": "Bearer", "expires_in": int(accessTokenTTL.Seconds()), "refresh_token": refreshToken, "scope": scope})
}

func (h *Handler) Revoke(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err == nil {
		_ = h.cfg.Cache.Delete(ctx, oauthKey("refresh", hashToken(r.Form.Get("token"))))
	}
	w.WriteHeader(http.StatusOK)
	return nil
}
func (h *Handler) loadClient(ctx context.Context, clientID string) (oauthClient, error) {
	var client oauthClient
	if h.cfg.Cache == nil {
		return client, errors.New("oauth cache unavailable")
	}
	err := h.cfg.Cache.Get(ctx, oauthKey("client", clientID), &client)
	return client, err
}
func normalizeScope(raw string) string {
	fields := strings.Fields(raw)
	if !slicesContains(fields, mcpScope) {
		fields = append(fields, mcpScope)
	}
	if !slicesContains(fields, "offline_access") {
		fields = append(fields, "offline_access")
	}
	return strings.Join(fields, " ")
}
func validateRedirectURI(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || parsed.Fragment != "" {
		return errors.New("redirect URI must be an absolute URI without a fragment")
	}
	if parsed.User != nil {
		return errors.New("redirect URI must not contain user information")
	}
	if parsed.Scheme == "https" {
		return nil
	}
	host := parsed.Hostname()
	if parsed.Scheme == "http" && (host == "localhost" || host == "127.0.0.1" || host == "::1") {
		return nil
	}
	return errors.New("redirect URI must use HTTPS, except for localhost clients")
}

func oauthRedirectOrigin(raw string) (string, error) {
	if err := validateRedirectURI(raw); err != nil {
		return "", err
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	return parsed.Scheme + "://" + parsed.Host, nil
}
func randomToken(bytes int) (string, error) {
	data := make([]byte, bytes)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}
func hashToken(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}
func oauthKey(kind, id string) string { return "mcp:oauth:" + kind + ":" + id }
func slicesContains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
func absoluteRequestURL(base string, r *http.Request) string {
	parsed, _ := url.Parse(strings.TrimRight(base, "/"))
	parsed.Path = r.URL.Path
	parsed.RawQuery = r.URL.RawQuery
	return parsed.String()
}
func redirectOAuthError(w http.ResponseWriter, r *http.Request, redirectURI, state, code, description string) error {
	destination, err := url.Parse(redirectURI)
	if err != nil {
		return oauthError(w, http.StatusBadRequest, code, description)
	}
	values := destination.Query()
	values.Set("error", code)
	values.Set("error_description", description)
	if state != "" {
		values.Set("state", state)
	}
	destination.RawQuery = values.Encode()
	http.Redirect(w, r, destination.String(), http.StatusFound)
	return nil
}
func oauthError(w http.ResponseWriter, status int, code, description string) error {
	return writeJSON(w, status, map[string]any{"error": code, "error_description": description})
}

func (h *Handler) oauthError(ctx context.Context, w http.ResponseWriter, operation string, status int, code, description string) error {
	if h.cfg.Log != nil {
		h.cfg.Log.Warn(ctx, "MCP OAuth request rejected", "operation", operation, "status", status, "oauth_error", code)
	}
	return oauthError(w, status, code, description)
}

func writeJSON(w http.ResponseWriter, status int, value any) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(value)
}
