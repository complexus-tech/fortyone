package agentreadinesshttp

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"strings"
	"time"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	developeroauth "github.com/complexus-tech/projects-api/internal/modules/developeroauth/service"
	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	objectivestatus "github.com/complexus-tech/projects-api/internal/modules/objectivestatus/service"
	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	sprints "github.com/complexus-tech/projects-api/internal/modules/sprints/service"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	mcpauth "github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	mcpScope         = "mcp:access"
	mcpServerVersion = "1.1.0"
	mcpInstructions  = "Use FortyOne as the system of record for delivery. Read before writing, preserve scheduling fields, and ask the user before invoking a write tool. Treat task, issue, ticket, and work item as story; project and goal as objective; iteration and cycle as sprint; and KR, outcome, measure, and target as key result. Use the matching update tool to edit stories, objectives, and key results. Sprint editing is not available, so never claim or imply that an existing sprint was changed. Resolve requested story or objective workflow states with the matching status-list tool before updating. For 'my work', set assignedToMe. For work due on a named day such as today, resolve it to YYYY-MM-DD and set dueOn. Respect page boundaries and follow hasMore only when more results are needed."
)

//go:embed openapi.json
var openAPIDescription []byte

type Config struct {
	APIPublicURL      string
	Workspaces        *workspaces.Service
	Teams             *teams.Service
	States            *states.Service
	Stories           *stories.Service
	Sprints           *sprints.Service
	Objectives        *objectives.Service
	ObjectiveStatuses *objectivestatus.Service
	KeyResults        *keyresults.Service
	Reports           *reports.Service
	OAuth             oauthPlatform
	Cache             oauthStore
	BrowserSessions   mid.SessionResolver
	LoginURL          string
	Log               *logger.Logger
}

type oauthPlatform interface {
	Resource() string
	RegisterPublicApplication(context.Context, string, []string) (developeroauthdomain.Application, error)
	PrepareAuthorization(context.Context, developeroauth.AuthorizationRequest) (developeroauthdomain.Application, []string, error)
	AuthorizeUser(context.Context, developeroauth.AuthorizationRequest) (developeroauthdomain.PlaintextSecret, error)
	ExchangeAuthorizationCode(context.Context, developeroauth.AuthorizationCodeExchange) (developeroauthdomain.TokenPair, error)
	ExchangeRefreshToken(context.Context, developeroauth.RefreshExchange) (developeroauthdomain.TokenPair, error)
	ExchangeClientCredentials(context.Context, developeroauth.ClientCredentialsExchange) (developeroauthdomain.ApplicationAccessToken, error)
	RevokeRefreshToken(context.Context, string) error
	VerifyAccessToken(context.Context, string) (developeroauthdomain.AccessIdentity, error)
}

type oauthResourceCatalog interface {
	Resources() []string
	SupportedScopes(string) []string
}

type oauthStore interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Get(ctx context.Context, key string, dest any) error
	Take(ctx context.Context, key string, dest any) error
	IncrementWithTTL(ctx context.Context, key string, ttl time.Duration) (int64, error)
}

type Handler struct {
	cfg         Config
	resource    string
	apiResource string
	handler     http.Handler
}

func New(cfg Config) *Handler {
	resource := strings.TrimRight(cfg.APIPublicURL, "/") + "/mcp"
	if cfg.OAuth != nil {
		resource = cfg.OAuth.Resource()
	}
	h := &Handler{
		cfg: cfg, resource: resource,
		apiResource: strings.TrimRight(cfg.APIPublicURL, "/") + "/api/v1",
	}
	server := h.newMCPServer()
	transport := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true})
	h.handler = mcpauth.RequireBearerToken(h.verifyToken, &mcpauth.RequireBearerTokenOptions{
		ResourceMetadataURL: strings.TrimRight(cfg.APIPublicURL, "/") + "/.well-known/oauth-protected-resource",
		Scopes:              []string{mcpScope}, ClockSkew: 30 * time.Second,
	})(h.mcpGrantRateLimit(transport, mcpGrantRequestLimit, mcpGrantRequestWindow))
	return h
}

func (h *Handler) newMCPServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "app.fortyone", Version: mcpServerVersion}, &mcp.ServerOptions{
		Instructions: mcpInstructions,
	})
	h.addTools(server)
	return server
}

func (h *Handler) OpenAPI(_ context.Context, w http.ResponseWriter, _ *http.Request) error {
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(openAPIDescription)
	return err
}

func (h *Handler) MCP(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	h.handler.ServeHTTP(w, r.WithContext(ctx))
	return nil
}

func (h *Handler) verifyToken(ctx context.Context, raw string, _ *http.Request) (*mcpauth.TokenInfo, error) {
	if h.cfg.OAuth == nil {
		return nil, fmt.Errorf("%w: OAuth service is unavailable", mcpauth.ErrInvalidToken)
	}
	identity, err := h.cfg.OAuth.VerifyAccessToken(ctx, raw)
	if err != nil {
		return nil, fmt.Errorf("%w: bearer token is invalid or expired", mcpauth.ErrInvalidToken)
	}
	return &mcpauth.TokenInfo{
		Scopes: identity.Scopes, Expiration: identity.ExpiresAt, UserID: identity.UserID.String(),
		Extra: map[string]any{"grant_id": identity.GrantID.String()},
	}, nil
}
