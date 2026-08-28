package agentreadinesshttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	developeroauth "github.com/complexus-tech/projects-api/internal/modules/developeroauth/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

const maximumOAuthAuthorizationHeaderBytes = 16 << 10

const (
	maximumOAuthInstallationIDBytes = 64
	maximumOAuthResourceBytes       = 2048
	maximumOAuthScopeBytes          = 4096
)

func (h *Handler) exchangeClientCredentials(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
) error {
	clientID, clientSecret, ok := clientSecretBasic(r)
	if !ok || r.PostForm.Has("client_id") || r.PostForm.Has("client_secret") ||
		r.URL.Query().Has("client_id") || r.URL.Query().Has("client_secret") {
		return h.invalidOAuthClient(ctx, w)
	}
	installationValue, ok := singleOAuthFormValue(r, "installation_id", maximumOAuthInstallationIDBytes)
	if !ok {
		return h.oauthError(ctx, w, "exchange_client_credentials", http.StatusBadRequest, "invalid_request", "installation_id must appear exactly once")
	}
	resource, ok := singleOAuthFormValue(r, "resource", maximumOAuthResourceBytes)
	if !ok {
		return h.oauthError(ctx, w, "exchange_client_credentials", http.StatusBadRequest, "invalid_request", "resource must appear exactly once")
	}
	scope, ok := singleOAuthFormValue(r, "scope", maximumOAuthScopeBytes)
	if !ok {
		return h.oauthError(ctx, w, "exchange_client_credentials", http.StatusBadRequest, "invalid_request", "scope must appear exactly once")
	}
	installationID, err := uuid.Parse(installationValue)
	if err != nil || installationID == uuid.Nil {
		return h.oauthError(ctx, w, "exchange_client_credentials", http.StatusBadRequest, "invalid_request", "installation_id must be a UUID")
	}
	if h.cfg.OAuth == nil {
		return h.oauthError(ctx, w, "exchange_client_credentials", http.StatusServiceUnavailable, "temporarily_unavailable", "OAuth service is unavailable")
	}
	issued, err := h.cfg.OAuth.ExchangeClientCredentials(ctx, developeroauth.ClientCredentialsExchange{
		ClientID: clientID, ClientSecret: clientSecret, InstallationID: installationID,
		Resource: resource, Scopes: strings.Fields(scope),
		RequestID: web.GetRequestID(ctx),
	})
	if err != nil {
		switch {
		case errors.Is(err, developeroauthdomain.ErrInvalidResource):
			return h.oauthError(ctx, w, "exchange_client_credentials", http.StatusBadRequest, "invalid_target", "resource is not a supported FortyOne OAuth audience")
		case errors.Is(err, developeroauthdomain.ErrInvalidScope):
			return h.oauthError(ctx, w, "exchange_client_credentials", http.StatusBadRequest, "invalid_scope", "requested scope is not installed for this application")
		case errors.Is(err, developeroauthdomain.ErrApplicationActorUnavailable):
			return h.oauthError(ctx, w, "exchange_client_credentials", http.StatusBadRequest, "unauthorized_client", "application actors are not enabled for this resource")
		case errors.Is(err, developeroauthdomain.ErrInvalidClient):
			return h.invalidOAuthClient(ctx, w)
		default:
			return fmt.Errorf("exchange OAuth client credentials: %w", err)
		}
	}
	return writeApplicationAccessToken(w, issued)
}

func clientSecretBasic(request *http.Request) (string, string, bool) {
	if request == nil {
		return "", "", false
	}
	values := request.Header.Values("Authorization")
	if len(values) != 1 || len(values[0]) == 0 || len(values[0]) > maximumOAuthAuthorizationHeaderBytes {
		return "", "", false
	}
	clientID, clientSecret, ok := request.BasicAuth()
	if !ok || clientID == "" || clientSecret == "" || strings.TrimSpace(clientID) != clientID ||
		strings.TrimSpace(clientSecret) != clientSecret {
		return "", "", false
	}
	return clientID, clientSecret, true
}

func (h *Handler) invalidOAuthClient(ctx context.Context, writer http.ResponseWriter) error {
	writer.Header().Set("WWW-Authenticate", `Basic realm="FortyOne OAuth", error="invalid_client"`)
	return h.oauthError(
		ctx,
		writer,
		"exchange_client_credentials",
		http.StatusUnauthorized,
		"invalid_client",
		"client authentication failed",
	)
}

func writeApplicationAccessToken(
	writer http.ResponseWriter,
	issued developeroauthdomain.ApplicationAccessToken,
) error {
	return writeJSON(writer, http.StatusOK, map[string]any{
		"access_token": issued.AccessToken.Reveal(),
		"token_type":   "Bearer",
		"expires_in":   int(issued.ExpiresIn.Seconds()),
		"scope":        developeroauth.ScopeString(issued.Scopes),
	})
}
