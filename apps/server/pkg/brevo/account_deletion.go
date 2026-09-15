package brevo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

var ErrAccountCleanupUnavailable = errors.New("Brevo account cleanup is not configured")

// DeleteAccountContact requires explicit provider confirmation. Disabled
// integrations and transport failures must leave durable cleanup pending.
// Provider errors can contain the contact address, so expose only safe status.
func (s *Service) DeleteAccountContact(ctx context.Context, email string) error {
	if s == nil || !s.enabled || s.client == nil {
		return ErrAccountCleanupUnavailable
	}
	if strings.TrimSpace(email) == "" {
		return errors.New("Brevo account cleanup requires an address")
	}
	response, err := s.client.ContactsApi.DeleteContact(ctx, url.PathEscape(email))
	if response != nil && (response.StatusCode == http.StatusNoContent || response.StatusCode == http.StatusNotFound) {
		return nil
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if response != nil {
		return fmt.Errorf("Brevo account cleanup was not confirmed (status %d)", response.StatusCode)
	}
	if err != nil {
		return errors.New("Brevo account cleanup request failed")
	}
	return errors.New("Brevo account cleanup returned no confirmation")
}
