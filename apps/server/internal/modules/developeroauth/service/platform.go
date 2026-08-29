package developeroauth

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
)

// Platform routes one authorization server across exact audience-specific
// services. Registration metadata is shared, while grants, tokens, and scope
// policies remain bound to the requested resource.
type Platform struct {
	defaultService *Service
	resources      map[string]*Service
}

func NewPlatform(defaultService *Service, services ...*Service) (*Platform, error) {
	if defaultService == nil {
		return nil, errors.New("default developer OAuth resource service is required")
	}
	resources := make(map[string]*Service, len(services)+1)
	for _, service := range append([]*Service{defaultService}, services...) {
		if service == nil {
			return nil, errors.New("developer OAuth resource service is required")
		}
		resource := strings.TrimSpace(service.Resource())
		if resource == "" {
			return nil, errors.New("developer OAuth resource is required")
		}
		if _, duplicate := resources[resource]; duplicate {
			return nil, fmt.Errorf("developer OAuth resource %q is duplicated", resource)
		}
		resources[resource] = service
	}
	return &Platform{defaultService: defaultService, resources: resources}, nil
}

// Resource preserves the default-resource contract used by the MCP protected
// resource. Authorization requests may target any exact resource registered in
// this platform.
func (platform *Platform) Resource() string {
	if platform == nil || platform.defaultService == nil {
		return ""
	}
	return platform.defaultService.Resource()
}

func (platform *Platform) Resources() []string {
	if platform == nil {
		return nil
	}
	resources := make([]string, 0, len(platform.resources))
	for resource := range platform.resources {
		resources = append(resources, resource)
	}
	sort.Strings(resources)
	return resources
}

func (platform *Platform) SupportedScopes(resource string) []string {
	service, ok := platform.resourceService(resource)
	if !ok {
		return nil
	}
	return service.SupportedScopes()
}

func (platform *Platform) RegisterPublicApplication(
	ctx context.Context,
	name string,
	redirectURIs []string,
) (developeroauthdomain.Application, error) {
	if platform == nil || platform.defaultService == nil {
		return developeroauthdomain.Application{}, developeroauthdomain.ErrApplicationNotFound
	}
	return platform.defaultService.RegisterPublicApplication(ctx, name, redirectURIs)
}

func (platform *Platform) PrepareAuthorization(
	ctx context.Context,
	request AuthorizationRequest,
) (developeroauthdomain.Application, []string, error) {
	service, ok := platform.resourceService(request.Resource)
	if !ok {
		// Preserve the registered redirect safety ordering even when the audience
		// is unknown. The default service validates the client and exact redirect
		// before it returns ErrInvalidResource, allowing the HTTP boundary to
		// redirect only a trusted client URI and never an attacker-controlled one.
		return platform.defaultService.PrepareAuthorization(ctx, request)
	}
	return service.PrepareAuthorization(ctx, request)
}

func (platform *Platform) AuthorizeUser(
	ctx context.Context,
	request AuthorizationRequest,
) (developeroauthdomain.PlaintextSecret, error) {
	service, ok := platform.resourceService(request.Resource)
	if !ok {
		return developeroauthdomain.PlaintextSecret{}, developeroauthdomain.ErrInvalidResource
	}
	return service.AuthorizeUser(ctx, request)
}

func (platform *Platform) ExchangeAuthorizationCode(
	ctx context.Context,
	request AuthorizationCodeExchange,
) (developeroauthdomain.TokenPair, error) {
	service, ok := platform.resourceService(request.Resource)
	if !ok {
		return developeroauthdomain.TokenPair{}, developeroauthdomain.ErrInvalidResource
	}
	return service.ExchangeAuthorizationCode(ctx, request)
}

func (platform *Platform) ExchangeRefreshToken(
	ctx context.Context,
	request RefreshExchange,
) (developeroauthdomain.TokenPair, error) {
	service, ok := platform.resourceService(request.Resource)
	if !ok {
		return developeroauthdomain.TokenPair{}, developeroauthdomain.ErrInvalidResource
	}
	return service.ExchangeRefreshToken(ctx, request)
}

func (platform *Platform) ExchangeClientCredentials(
	ctx context.Context,
	request ClientCredentialsExchange,
) (developeroauthdomain.ApplicationAccessToken, error) {
	service, ok := platform.resourceService(request.Resource)
	if !ok {
		return developeroauthdomain.ApplicationAccessToken{}, developeroauthdomain.ErrInvalidResource
	}
	return service.ExchangeClientCredentials(ctx, request)
}

func (platform *Platform) RevokeRefreshToken(ctx context.Context, raw string) error {
	if platform == nil || platform.defaultService == nil {
		return developeroauthdomain.ErrRefreshToken
	}
	// Refresh secrets are globally unique within the shared repository and the
	// revocation operation intentionally reveals no resource-specific result.
	return platform.defaultService.RevokeRefreshToken(ctx, raw)
}

// VerifyAccessToken verifies only the default protected resource. This keeps
// an API-audience token from being accepted by MCP. External API adapters must
// receive and call their exact resource Service instead of this method.
func (platform *Platform) VerifyAccessToken(
	ctx context.Context,
	raw string,
) (developeroauthdomain.AccessIdentity, error) {
	if platform == nil || platform.defaultService == nil {
		return developeroauthdomain.AccessIdentity{}, developeroauthdomain.ErrInvalidGrant
	}
	return platform.defaultService.VerifyAccessToken(ctx, raw)
}

func (platform *Platform) resourceService(resource string) (*Service, bool) {
	if platform == nil {
		return nil, false
	}
	service, ok := platform.resources[strings.TrimSpace(resource)]
	return service, ok
}
