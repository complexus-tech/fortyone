package agentreadinesshttp

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	objectivestatus "github.com/complexus-tech/projects-api/internal/modules/objectivestatus/service"
	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	sprints "github.com/complexus-tech/projects-api/internal/modules/sprints/service"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	mcpauth "github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	mcpScope        = "mcp:access"
	mcpInstructions = "Use FortyOne as the system of record for delivery. Read before writing, preserve scheduling fields, and ask the user before invoking a write tool. Treat task, issue, ticket, and work item as story; project and goal as objective; iteration and cycle as sprint; and KR, outcome, measure, and target as key result. For 'my work', set assignedToMe. For work due on a named day such as today, resolve it to YYYY-MM-DD and set dueOn."
)

//go:embed openapi.json
var openAPIDescription []byte

type Config struct {
	SecretKey         string
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
	Cache             oauthStore
	LoginURL          string
	Log               *logger.Logger
}

type oauthStore interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Get(ctx context.Context, key string, dest any) error
	Delete(ctx context.Context, key string) error
}

type Handler struct {
	cfg      Config
	resource string
	handler  http.Handler
}

type mcpClaims struct {
	jwt.RegisteredClaims
	Scope string `json:"scope"`
}

func New(cfg Config) *Handler {
	resource := strings.TrimRight(cfg.APIPublicURL, "/") + "/mcp"
	h := &Handler{cfg: cfg, resource: resource}
	server := h.newMCPServer()
	transport := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true})
	h.handler = mcpauth.RequireBearerToken(h.verifyToken, &mcpauth.RequireBearerTokenOptions{
		ResourceMetadataURL: strings.TrimRight(cfg.APIPublicURL, "/") + "/.well-known/oauth-protected-resource",
		Scopes:              []string{mcpScope}, ClockSkew: 30 * time.Second,
	})(transport)
	return h
}

func (h *Handler) newMCPServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "app.fortyone", Version: "1.0.0"}, &mcp.ServerOptions{
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

func (h *Handler) verifyToken(_ context.Context, raw string, _ *http.Request) (*mcpauth.TokenInfo, error) {
	claims := &mcpClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, mcpauth.ErrInvalidToken
		}
		return h.signingKey(), nil
	}, jwt.WithAudience(h.resource), jwt.WithExpirationRequired())
	if err != nil || !token.Valid || claims.Subject == "" || claims.ExpiresAt == nil {
		return nil, fmt.Errorf("%w: bearer token is invalid or expired", mcpauth.ErrInvalidToken)
	}
	if !slices.Contains(strings.Fields(claims.Scope), mcpScope) {
		return nil, fmt.Errorf("%w: bearer token does not grant MCP access", mcpauth.ErrInvalidToken)
	}
	if _, err := uuid.Parse(claims.Subject); err != nil {
		return nil, fmt.Errorf("%w: bearer token subject is invalid", mcpauth.ErrInvalidToken)
	}
	return &mcpauth.TokenInfo{Scopes: strings.Fields(claims.Scope), Expiration: claims.ExpiresAt.Time, UserID: claims.Subject}, nil
}

func (h *Handler) signingKey() []byte {
	digest := sha256.Sum256([]byte("fortyone:mcp:access-token:v1\x00" + h.cfg.SecretKey))
	return digest[:]
}
