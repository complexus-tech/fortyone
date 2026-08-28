package developeroauth

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
)

const maxApplicationRedirectURIs = 10

func (service *Service) RegisterPublicApplication(
	ctx context.Context,
	name string,
	redirectURIs []string,
) (developeroauthdomain.Application, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "MCP client"
	}
	if len(name) > 120 {
		return developeroauthdomain.Application{}, errors.New("OAuth application name must not exceed 120 characters")
	}
	normalizedRedirects, err := normalizeRedirectURIs(redirectURIs)
	if err != nil {
		return developeroauthdomain.Application{}, err
	}
	id, err := service.nextID()
	if err != nil {
		return developeroauthdomain.Application{}, fmt.Errorf("generate OAuth application ID: %w", err)
	}
	clientIDBytes := make([]byte, 24)
	if _, err := io.ReadFull(service.random, clientIDBytes); err != nil {
		return developeroauthdomain.Application{}, fmt.Errorf("generate OAuth client ID: %w", err)
	}
	defer zeroByteSlices([][]byte{clientIDBytes})
	now := service.clock.Now().UTC()
	return service.repository.CreateApplication(ctx, developeroauthdomain.RegisterApplication{
		ID:       id,
		ClientID: "f41_oauth_" + base64.RawURLEncoding.EncodeToString(clientIDBytes),
		Name:     name, RegistrationKind: "dynamic_public", RedirectURIs: normalizedRedirects,
		ExpiresAt: now.Add(service.dynamicClientTTL), CreatedAt: now,
	})
}

func (service *Service) GetApplication(ctx context.Context, clientID string) (developeroauthdomain.Application, error) {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return developeroauthdomain.Application{}, developeroauthdomain.ErrApplicationNotFound
	}
	return service.repository.GetActiveApplication(ctx, clientID, service.clock.Now().UTC())
}

func normalizeRedirectURIs(values []string) ([]string, error) {
	if len(values) == 0 || len(values) > maxApplicationRedirectURIs {
		return nil, fmt.Errorf("%w: provide between 1 and %d redirect URIs", developeroauthdomain.ErrInvalidRedirectURI, maxApplicationRedirectURIs)
	}
	unique := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if err := ValidateRedirectURI(value); err != nil {
			return nil, err
		}
		unique[value] = struct{}{}
	}
	result := make([]string, 0, len(unique))
	for value := range unique {
		result = append(result, value)
	}
	sort.Strings(result)
	return result, nil
}
