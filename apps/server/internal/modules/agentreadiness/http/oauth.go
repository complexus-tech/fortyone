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

var approvalPage = template.Must(template.New("mcp-approval").Parse(`<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width"><title>Connect to FortyOne</title><style>body{font:16px system-ui;margin:0;background:#f7f6f2;color:#191918}.card{max-width:460px;margin:12vh auto;padding:28px;background:white;border:1px solid #dedbd2;border-radius:16px}h1{font-size:22px}p{line-height:1.5;color:#555}button{border:0;border-radius:9px;padding:11px 16px;font-weight:650;cursor:pointer}.allow{background:#191918;color:white}.deny{background:#eeeae2;margin-left:8px}</style></head><body><main class="card"><h1>Connect {{.ClientName}} to FortyOne?</h1><p>This connection can read your accessible work and, when you explicitly approve a tool call, create stories, sprints, objectives, and key results.</p><form method="post" action="/oauth/authorize"><input type="hidden" name="approval" value="{{.Approval}}"><button class="allow" name="decision" value="allow">Allow</button><button class="deny" name="decision" value="deny">Cancel</button></form></main></body></html>`))

func (h *Handler) ProtectedResourceMetadata(_ context.Context, w http.ResponseWriter, _ *http.Request) error {
	base := strings.TrimRight(h.cfg.APIPublicURL, "/")
	return writeJSON(w, http.StatusOK, map[string]any{"resource": h.resource, "authorization_servers": []string{base}, "scopes_supported": []string{mcpScope, "offline_access"}, "bearer_methods_supported": []string{"header"}})
}

func (h *Handler) AuthorizationServerMetadata(_ context.Context, w http.ResponseWriter, _ *http.Request) error {
	base := strings.TrimRight(h.cfg.APIPublicURL, "/")
	return writeJSON(w, http.StatusOK, map[string]any{"issuer": base, "authorization_endpoint": base + "/oauth/authorize", "token_endpoint": base + "/oauth/token", "registration_endpoint": base + "/oauth/register", "revocation_endpoint": base + "/oauth/revoke", "response_types_supported": []string{"code"}, "grant_types_supported": []string{"authorization_code", "refresh_token"}, "code_challenge_methods_supported": []string{"S256"}, "token_endpoint_auth_methods_supported": []string{"none"}, "scopes_supported": []string{mcpScope, "offline_access"}})
}

func (h *Handler) RegisterClient(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var request struct {
		ClientName              string   `json:"client_name"`
		RedirectURIs            []string `json:"redirect_uris"`
		TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&request); err != nil {
		return oauthError(w, http.StatusBadRequest, "invalid_client_metadata", "invalid JSON registration")
	}
	if request.TokenEndpointAuthMethod != "" && request.TokenEndpointAuthMethod != "none" {
		return oauthError(w, http.StatusBadRequest, "invalid_client_metadata", "only public PKCE clients are supported")
	}
	if len(request.RedirectURIs) == 0 {
		return oauthError(w, http.StatusBadRequest, "invalid_redirect_uri", "at least one redirect URI is required")
	}
	for _, redirectURI := range request.RedirectURIs {
		if err := validateRedirectURI(redirectURI); err != nil {
			return oauthError(w, http.StatusBadRequest, "invalid_redirect_uri", err.Error())
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
		return oauthError(w, http.StatusBadRequest, "unsupported_response_type", "response_type must be code")
	}
	client, err := h.loadClient(ctx, query.Get("client_id"))
	if err != nil {
		return oauthError(w, http.StatusBadRequest, "invalid_request", "unknown client_id")
	}
	redirectURI := query.Get("redirect_uri")
	if !slicesContains(client.RedirectURIs, redirectURI) {
		return oauthError(w, http.StatusBadRequest, "invalid_request", "redirect_uri is not registered")
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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	return approvalPage.Execute(w, map[string]string{"ClientName": client.ClientName, "Approval": approval})
}

func (h *Handler) approveAuthorization(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return oauthError(w, http.StatusBadRequest, "invalid_request", "invalid approval")
	}
	approval := r.Form.Get("approval")
	var pending authorizationRequest
	if err := h.cfg.Cache.Get(ctx, oauthKey("approval", approval), &pending); err != nil {
		return oauthError(w, http.StatusBadRequest, "invalid_request", "approval expired")
	}
	_ = h.cfg.Cache.Delete(ctx, oauthKey("approval", approval))
	userID, ok, err := mid.ResolveSessionUserID(ctx, r)
	if err != nil {
		return err
	}
	if !ok || userID.String() != pending.UserID {
		return oauthError(w, http.StatusUnauthorized, "access_denied", "session is missing or changed")
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
		return oauthError(w, http.StatusBadRequest, "invalid_request", "invalid form")
	}
	switch r.Form.Get("grant_type") {
	case "authorization_code":
		return h.exchangeCode(ctx, w, r)
	case "refresh_token":
		return h.exchangeRefreshToken(ctx, w, r)
	default:
		return oauthError(w, http.StatusBadRequest, "unsupported_grant_type", "grant_type is not supported")
	}
}

func (h *Handler) exchangeCode(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if r.Form.Get("resource") != h.resource {
		return oauthError(w, http.StatusBadRequest, "invalid_target", "resource must match the FortyOne MCP endpoint")
	}
	code := r.Form.Get("code")
	key := oauthKey("code", hashToken(code))
	var record authorizationCode
	if err := h.cfg.Cache.Get(ctx, key, &record); err != nil {
		return oauthError(w, http.StatusBadRequest, "invalid_grant", "authorization code is invalid or expired")
	}
	_ = h.cfg.Cache.Delete(ctx, key)
	if record.ClientID != r.Form.Get("client_id") || record.RedirectURI != r.Form.Get("redirect_uri") {
		return oauthError(w, http.StatusBadRequest, "invalid_grant", "authorization code binding does not match")
	}
	digest := sha256.Sum256([]byte(r.Form.Get("code_verifier")))
	if base64.RawURLEncoding.EncodeToString(digest[:]) != record.CodeChallenge {
		return oauthError(w, http.StatusBadRequest, "invalid_grant", "PKCE verification failed")
	}
	return h.issueTokens(ctx, w, record.UserID, record.ClientID, record.Scope)
}

func (h *Handler) exchangeRefreshToken(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if r.Form.Get("resource") != h.resource {
		return oauthError(w, http.StatusBadRequest, "invalid_target", "resource must match the FortyOne MCP endpoint")
	}
	raw := r.Form.Get("refresh_token")
	key := oauthKey("refresh", hashToken(raw))
	var grant refreshGrant
	if err := h.cfg.Cache.Get(ctx, key, &grant); err != nil {
		return oauthError(w, http.StatusBadRequest, "invalid_grant", "refresh token is invalid or expired")
	}
	_ = h.cfg.Cache.Delete(ctx, key)
	if grant.ClientID != r.Form.Get("client_id") {
		return oauthError(w, http.StatusBadRequest, "invalid_grant", "refresh token client does not match")
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
	if parsed.Scheme == "https" {
		return nil
	}
	host := parsed.Hostname()
	if parsed.Scheme == "http" && (host == "localhost" || host == "127.0.0.1" || host == "::1") {
		return nil
	}
	return errors.New("redirect URI must use HTTPS, except for localhost clients")
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
func writeJSON(w http.ResponseWriter, status int, value any) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(value)
}
