package workspacesecuritydomain

import (
	"fmt"
	"net/mail"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/net/idna"
)

func NormalizePolicy(input PolicyUpdate) (PolicyUpdate, error) {
	if input.ExpectedVersion < 0 || input.MaxSessionAgeHours < 0 || input.MaxSessionAgeHours > 720 || len(input.AllowedDomains) > 50 {
		return PolicyUpdate{}, ErrInvalid
	}
	seen := make(map[string]bool, len(input.AllowedDomains))
	domains := make([]string, 0, len(input.AllowedDomains))
	for _, value := range input.AllowedDomains {
		value = strings.ToLower(strings.TrimSpace(value))
		ascii, err := idna.Lookup.ToASCII(value)
		if err != nil || ascii == "" || len(ascii) > 253 || strings.ContainsAny(ascii, "@/:* ") || !strings.Contains(ascii, ".") {
			return PolicyUpdate{}, fmt.Errorf("%w: use complete email domains without wildcards", ErrInvalid)
		}
		for _, label := range strings.Split(ascii, ".") {
			if label == "" || len(label) > 63 || strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
				return PolicyUpdate{}, ErrInvalid
			}
			for _, char := range label {
				if (char < 'a' || char > 'z') && (char < '0' || char > '9') && char != '-' {
					return PolicyUpdate{}, ErrInvalid
				}
			}
		}
		if !seen[ascii] {
			domains = append(domains, ascii)
			seen[ascii] = true
		}
	}
	slices.Sort(domains)
	input.AllowedDomains = domains
	return input, nil
}

func EmailAllowed(email string, domains []string) bool {
	if len(domains) == 0 {
		return true
	}
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return false
	}
	_, rawDomain, ok := strings.Cut(email, "@")
	if !ok {
		return false
	}
	domain, err := idna.Lookup.ToASCII(strings.ToLower(rawDomain))
	return err == nil && slices.Contains(domains, domain)
}

// Legacy opaque sessions are accepted by the default policy. A policy requiring
// a true authentication timestamp or a tenant revocation fence requires login.
func CheckAccess(state AccessState, session SessionIdentity, now time.Time) error {
	if !EmailAllowed(state.Email, state.Policy.AllowedDomains) || (state.Role == "guest" && !state.Policy.AllowGuests) || state.RevokedAt != nil {
		return ErrForbidden
	}
	known := session.ID != [16]byte{} && !session.AuthenticatedAt.IsZero() && !session.ExpiresAt.IsZero()
	if !known {
		if state.RevokedBefore != nil || state.Policy.MaxSessionAgeHours > 0 {
			return ErrForbidden
		}
		return nil
	}
	if session.AuthenticatedAt.After(now.Add(30*time.Second)) || !session.ExpiresAt.After(now) || !session.ExpiresAt.After(session.AuthenticatedAt) {
		return ErrForbidden
	}
	if state.RevokedBefore != nil && !session.AuthenticatedAt.After(*state.RevokedBefore) {
		return ErrForbidden
	}
	if state.Policy.MaxSessionAgeHours > 0 && !session.AuthenticatedAt.Add(time.Duration(state.Policy.MaxSessionAgeHours)*time.Hour).After(now) {
		return ErrForbidden
	}
	return nil
}

func ValidateReason(reason string) error {
	if !utf8.ValidString(reason) || len(strings.TrimSpace(reason)) == 0 || utf8.RuneCountInString(reason) > 240 {
		return ErrInvalid
	}
	return nil
}
