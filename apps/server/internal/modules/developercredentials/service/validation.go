package developercredentials

import (
	"errors"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
)

const (
	minimumCredentialLifetime = time.Minute
	maximumCredentialLifetime = 365 * 24 * time.Hour
	maximumRotationOverlap    = 24 * time.Hour
	maximumTeamRestrictions   = 100
	maximumNameRunes          = 120
	maximumReasonRunes        = 240
)

func normalizeName(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || utf8.RuneCountInString(value) > maximumNameRunes {
		return "", developercredentialsdomain.ErrInvalidName
	}
	return value, nil
}

func normalizeReason(value string, defaultValue string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = defaultValue
	}
	if value == "" || utf8.RuneCountInString(value) > maximumReasonRunes {
		return "", developercredentialsdomain.ErrInvalidReason
	}
	return value, nil
}

func validateExpiry(now time.Time, expiresAt time.Time) (time.Time, error) {
	if expiresAt.IsZero() {
		return time.Time{}, developercredentialsdomain.ErrExpiryRequired
	}
	expiresAt = expiresAt.UTC()
	lifetime := expiresAt.Sub(now)
	if lifetime < minimumCredentialLifetime {
		return time.Time{}, developercredentialsdomain.ErrExpiryTooSoon
	}
	if lifetime > maximumCredentialLifetime {
		return time.Time{}, developercredentialsdomain.ErrExpiryTooLong
	}
	return expiresAt, nil
}

func normalizeScopes(scopes []platformauth.Scope) ([]platformauth.Scope, error) {
	if len(scopes) == 0 {
		return nil, developercredentialsdomain.ErrNoScopes
	}
	set, err := platformauth.NewScopeSet(scopes...)
	if err != nil {
		return nil, errors.Join(developercredentialsdomain.ErrInvalidScope, err)
	}
	if set.Has(platformauth.ScopeFirstParty) {
		return nil, developercredentialsdomain.ErrInvalidScope
	}
	values := set.Values()
	if len(values) == 0 {
		return nil, developercredentialsdomain.ErrNoScopes
	}
	return values, nil
}

func normalizeTeamIDs(teamIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(teamIDs) > maximumTeamRestrictions {
		return nil, developercredentialsdomain.ErrInvalidTeamRestriction
	}
	unique := make(map[uuid.UUID]struct{}, len(teamIDs))
	for _, teamID := range teamIDs {
		if teamID == uuid.Nil {
			return nil, developercredentialsdomain.ErrInvalidTeamRestriction
		}
		unique[teamID] = struct{}{}
	}
	result := make([]uuid.UUID, 0, len(unique))
	for teamID := range unique {
		result = append(result, teamID)
	}
	sort.Slice(result, func(left, right int) bool {
		return result[left].String() < result[right].String()
	})
	return result, nil
}

func validateServiceAccountRole(role authorization.WorkspaceRole) error {
	if err := authorization.ValidateWorkspaceRole(role); err != nil ||
		(role != authorization.WorkspaceRoleGuest && role != authorization.WorkspaceRoleMember) {
		return developercredentialsdomain.ErrInvalidServiceAccountRole
	}
	return nil
}

func validateServiceAccountScopes(scopes []platformauth.Scope) error {
	for _, scope := range scopes {
		if scope == platformauth.ScopeServiceAccountsManage {
			return developercredentialsdomain.ErrInvalidScope
		}
	}
	return nil
}

func validateRotationOverlap(value time.Duration, allowOverlap bool) error {
	if value < 0 || value > maximumRotationOverlap || (!allowOverlap && value != 0) {
		return errors.Join(
			developercredentialsdomain.ErrInvalidRotationOverlap,
			errors.New("overlap must be between zero and 24 hours for service-account keys and zero for personal tokens"),
		)
	}
	return nil
}
