package auth

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
)

// PrincipalKind preserves the identity that performed an operation. Provider,
// application, and system actions must never be silently attributed to the
// human who installed or configured them.
type PrincipalKind string

const (
	PrincipalHumanUser           PrincipalKind = "human_user"
	PrincipalPersonalToken       PrincipalKind = "personal_token"
	PrincipalServiceAccount      PrincipalKind = "service_account"
	PrincipalOAuthUser           PrincipalKind = "oauth_user"
	PrincipalOAuthApplication    PrincipalKind = "oauth_application"
	PrincipalSystem              PrincipalKind = "system"
	PrincipalExternalContributor PrincipalKind = "external_contributor"
)

// Scope identifies a stable resource/action permission. Scopes only narrow
// product permissions; possessing a scope never grants workspace membership or
// resource ownership.
type Scope string

const (
	ScopeFirstParty            Scope = "first_party:*"
	ScopeWorkspacesRead        Scope = "workspaces:read"
	ScopeTeamsRead             Scope = "teams:read"
	ScopeStoriesRead           Scope = "stories:read"
	ScopeStoriesWrite          Scope = "stories:write"
	ScopeCommentsRead          Scope = "comments:read"
	ScopeCommentsWrite         Scope = "comments:write"
	ScopeLabelsRead            Scope = "labels:read"
	ScopeLabelsWrite           Scope = "labels:write"
	ScopeSprintsRead           Scope = "sprints:read"
	ScopeObjectivesRead        Scope = "objectives:read"
	ScopeObjectivesWrite       Scope = "objectives:write"
	ScopeWebhooksManage        Scope = "webhooks:manage"
	ScopeIntegrationsManage    Scope = "integrations:manage"
	ScopeServiceAccountsManage Scope = "service_accounts:manage"
)

var (
	ErrActorNotFound             = errors.New("actor not found in context")
	ErrInvalidActor              = errors.New("invalid actor")
	ErrInvalidPrincipalKind      = errors.New("invalid principal kind")
	ErrInvalidScope              = errors.New("invalid scope")
	ErrActorWorkspaceUnavailable = errors.New("actor workspace is not selected")
)

var knownPrincipalKinds = map[PrincipalKind]struct{}{
	PrincipalHumanUser:           {},
	PrincipalPersonalToken:       {},
	PrincipalServiceAccount:      {},
	PrincipalOAuthUser:           {},
	PrincipalOAuthApplication:    {},
	PrincipalSystem:              {},
	PrincipalExternalContributor: {},
}

var knownScopes = map[Scope]struct{}{
	ScopeFirstParty:            {},
	ScopeWorkspacesRead:        {},
	ScopeTeamsRead:             {},
	ScopeStoriesRead:           {},
	ScopeStoriesWrite:          {},
	ScopeCommentsRead:          {},
	ScopeCommentsWrite:         {},
	ScopeLabelsRead:            {},
	ScopeLabelsWrite:           {},
	ScopeSprintsRead:           {},
	ScopeObjectivesRead:        {},
	ScopeObjectivesWrite:       {},
	ScopeWebhooksManage:        {},
	ScopeIntegrationsManage:    {},
	ScopeServiceAccountsManage: {},
}

// ScopeSet is an immutable-by-convention set. Constructors and accessors copy
// their data so callers cannot mutate an actor through a shared map or slice.
type ScopeSet struct {
	values map[Scope]struct{}
}

func NewScopeSet(scopes ...Scope) (ScopeSet, error) {
	values := make(map[Scope]struct{}, len(scopes))
	for _, scope := range scopes {
		if _, known := knownScopes[scope]; !known {
			return ScopeSet{}, fmt.Errorf("%w: %q", ErrInvalidScope, scope)
		}
		values[scope] = struct{}{}
	}
	return ScopeSet{values: values}, nil
}

func MustScopeSet(scopes ...Scope) ScopeSet {
	set, err := NewScopeSet(scopes...)
	if err != nil {
		panic(err)
	}
	return set
}

func (s ScopeSet) Has(scope Scope) bool {
	if _, firstParty := s.values[ScopeFirstParty]; firstParty {
		return true
	}
	_, ok := s.values[scope]
	return ok
}

func (s ScopeSet) ContainsAll(required ...Scope) bool {
	for _, scope := range required {
		if !s.Has(scope) {
			return false
		}
	}
	return true
}

func (s ScopeSet) Values() []Scope {
	values := make([]Scope, 0, len(s.values))
	for scope := range s.values {
		values = append(values, scope)
	}
	sort.Slice(values, func(left, right int) bool { return values[left] < values[right] })
	return values
}

func (s ScopeSet) clone() ScopeSet {
	values := make(map[Scope]struct{}, len(s.values))
	for scope := range s.values {
		values[scope] = struct{}{}
	}
	return ScopeSet{values: values}
}

// TeamAccess is a credential-level restriction. It does not replace the
// product's team-membership check.
type TeamAccess struct {
	unrestricted bool
	allowed      map[uuid.UUID]struct{}
}

func UnrestrictedTeamAccess() TeamAccess {
	return TeamAccess{unrestricted: true}
}

func RestrictedTeamAccess(teamIDs ...uuid.UUID) (TeamAccess, error) {
	allowed := make(map[uuid.UUID]struct{}, len(teamIDs))
	for _, teamID := range teamIDs {
		if teamID == uuid.Nil {
			return TeamAccess{}, fmt.Errorf("%w: team restriction contains a zero id", ErrInvalidActor)
		}
		allowed[teamID] = struct{}{}
	}
	return TeamAccess{allowed: allowed}, nil
}

func (a TeamAccess) Allows(teamID uuid.UUID) bool {
	if teamID == uuid.Nil {
		return false
	}
	if a.unrestricted {
		return true
	}
	_, ok := a.allowed[teamID]
	return ok
}

func (a TeamAccess) RestrictedTeamIDs() []uuid.UUID {
	teamIDs := make([]uuid.UUID, 0, len(a.allowed))
	for teamID := range a.allowed {
		teamIDs = append(teamIDs, teamID)
	}
	sort.Slice(teamIDs, func(left, right int) bool {
		return strings.Compare(teamIDs[left].String(), teamIDs[right].String()) < 0
	})
	return teamIDs
}

func (a TeamAccess) IsUnrestricted() bool {
	return a.unrestricted
}

func (a TeamAccess) clone() TeamAccess {
	allowed := make(map[uuid.UUID]struct{}, len(a.allowed))
	for teamID := range a.allowed {
		allowed[teamID] = struct{}{}
	}
	return TeamAccess{unrestricted: a.unrestricted, allowed: allowed}
}

// Actor is the immutable authentication and delegation context propagated into
// use cases. Workspace role and resource ownership remain policy inputs because
// they must be loaded from current authoritative state.
type Actor struct {
	PrincipalID  uuid.UUID
	Kind         PrincipalKind
	WorkspaceID  uuid.UUID
	CredentialID uuid.UUID
	Scopes       ScopeSet
	TeamAccess   TeamAccess
}

func NewActor(
	principalID uuid.UUID,
	kind PrincipalKind,
	credentialID uuid.UUID,
	scopes ScopeSet,
	teamAccess TeamAccess,
) (Actor, error) {
	actor := Actor{
		PrincipalID:  principalID,
		Kind:         kind,
		CredentialID: credentialID,
		Scopes:       scopes.clone(),
		TeamAccess:   teamAccess.clone(),
	}
	if err := actor.Validate(); err != nil {
		return Actor{}, err
	}
	return actor, nil
}

func NewHumanActor(userID uuid.UUID) Actor {
	actor, err := NewActor(
		userID,
		PrincipalHumanUser,
		uuid.Nil,
		MustScopeSet(ScopeFirstParty),
		UnrestrictedTeamAccess(),
	)
	if err != nil {
		panic(err)
	}
	return actor
}

func (a Actor) Validate() error {
	if a.PrincipalID == uuid.Nil {
		return fmt.Errorf("%w: principal id is required", ErrInvalidActor)
	}
	if _, known := knownPrincipalKinds[a.Kind]; !known {
		return fmt.Errorf("%w: %q", ErrInvalidPrincipalKind, a.Kind)
	}
	for _, scope := range a.Scopes.Values() {
		if _, known := knownScopes[scope]; !known {
			return fmt.Errorf("%w: %q", ErrInvalidScope, scope)
		}
	}
	if a.TeamAccess.unrestricted && len(a.TeamAccess.allowed) > 0 {
		return fmt.Errorf("%w: team access cannot be both unrestricted and restricted", ErrInvalidActor)
	}
	return nil
}

func (a Actor) WithWorkspace(workspaceID uuid.UUID) (Actor, error) {
	if workspaceID == uuid.Nil {
		return Actor{}, fmt.Errorf("%w: workspace id is required", ErrInvalidActor)
	}
	if err := a.Validate(); err != nil {
		return Actor{}, err
	}
	a.WorkspaceID = workspaceID
	return a.clone(), nil
}

func (a Actor) IsUserActor() bool {
	return a.Kind == PrincipalHumanUser || a.Kind == PrincipalPersonalToken || a.Kind == PrincipalOAuthUser
}

func (a Actor) UserID() (uuid.UUID, error) {
	if !a.IsUserActor() {
		return uuid.Nil, fmt.Errorf("%w: %s is not a user actor", ErrInvalidActor, a.Kind)
	}
	return a.PrincipalID, nil
}

func (a Actor) clone() Actor {
	a.Scopes = a.Scopes.clone()
	a.TeamAccess = a.TeamAccess.clone()
	return a
}
