// Package messagingcontext builds server-authoritative, provider-neutral
// context for messaging assistants.
package messagingcontext

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	"github.com/google/uuid"
)

// UserReader is the active-user lookup required for assistant identity and
// timezone context.
type UserReader interface {
	GetUser(ctx context.Context, userID uuid.UUID) (users.CoreUser, error)
}

// WorkspaceReader is the membership-scoped workspace and terminology lookup
// required for assistant context.
type WorkspaceReader interface {
	Get(ctx context.Context, workspaceID, userID uuid.UUID) (workspaces.CoreWorkspace, error)
	GetWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID) (workspaces.CoreWorkspaceSettings, error)
}

// TeamReader returns teams in the user's configured order. The provider still
// post-filters every row against the authoritative messaging audience.
type TeamReader interface {
	List(ctx context.Context, workspaceID, userID uuid.UUID, filters ...teams.CoreListTeamsFilter) ([]teams.CoreTeam, error)
}

// Provider loads display-safe runtime context from authoritative FortyOne
// services. It does not make authorization decisions; callers must supply the
// already-resolved team audience and tools independently reauthorize access.
type Provider struct {
	users      UserReader
	workspaces WorkspaceReader
	teams      TeamReader
}

// New constructs a runtime-context provider.
func New(userReader UserReader, workspaceReader WorkspaceReader, teamReader TeamReader) (*Provider, error) {
	if userReader == nil || workspaceReader == nil || teamReader == nil {
		return nil, errors.New("messaging context user, workspace, and team readers are required")
	}
	return &Provider{users: userReader, workspaces: workspaceReader, teams: teamReader}, nil
}

// Load returns context for one authenticated assistant turn. allowedTeamIDs
// follows messaging scope semantics: nil means every currently joined team,
// while an explicit empty slice means no team is available on this surface.
func (p *Provider) Load(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	allowedTeamIDs []uuid.UUID,
	surface messaging.RuntimeSurfaceContext,
	now time.Time,
) (*messaging.RuntimeContext, error) {
	if workspaceID == uuid.Nil || userID == uuid.Nil {
		return nil, errors.New("messaging context workspace and user are required")
	}
	if now.IsZero() {
		return nil, errors.New("messaging context current time is required")
	}

	actor, err := p.users.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load messaging actor context: %w", err)
	}
	workspace, err := p.workspaces.Get(ctx, workspaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("load messaging workspace context: %w", err)
	}
	settings, err := p.workspaces.GetWorkspaceSettings(ctx, workspaceID)
	if err != nil {
		if !errors.Is(err, workspaces.ErrNotFound) {
			return nil, fmt.Errorf("load messaging terminology context: %w", err)
		}
		settings = workspaces.CoreWorkspaceSettings{}
	}
	teamList, err := p.teams.List(ctx, workspaceID, userID, teams.CoreListTeamsFilter{JoinedOnly: true})
	if err != nil {
		return nil, fmt.Errorf("load messaging team context: %w", err)
	}

	location := time.UTC
	if timezone := strings.TrimSpace(actor.Timezone); timezone != "" {
		if parsed, parseErr := time.LoadLocation(timezone); parseErr == nil {
			location = parsed
		}
	}

	runtime := &messaging.RuntimeContext{
		Actor: messaging.RuntimeActorContext{
			DisplayName: actor.FullName,
			Username:    actor.Username,
		},
		Workspace: messaging.RuntimeWorkspaceContext{
			Name: workspace.Name,
			Slug: workspace.Slug,
			Role: workspace.UserRole,
		},
		LocalTime: now.In(location),
		Terminology: messaging.RuntimeTerminologyContext{
			Story:     runtimeTerm(settings.StoryTerm, "story"),
			Sprint:    runtimeTerm(settings.SprintTerm, "sprint"),
			Objective: runtimeTerm(settings.ObjectiveTerm, "objective"),
			KeyResult: runtimeTerm(settings.KeyResultTerm, "key result"),
		},
		Surface: surface,
	}
	runtime.TeamHints = authorizedTeamHints(teamList, workspaceID, allowedTeamIDs)
	return runtime, nil
}

func authorizedTeamHints(teamList []teams.CoreTeam, workspaceID uuid.UUID, allowedTeamIDs []uuid.UUID) []messaging.RuntimeTeamHint {
	var allowed map[uuid.UUID]struct{}
	if allowedTeamIDs != nil {
		allowed = make(map[uuid.UUID]struct{}, len(allowedTeamIDs))
		for _, teamID := range allowedTeamIDs {
			if teamID != uuid.Nil {
				allowed[teamID] = struct{}{}
			}
		}
	}

	hints := make([]messaging.RuntimeTeamHint, 0, min(len(teamList), messaging.MaximumRuntimeContextTeamHints))
	seen := make(map[uuid.UUID]struct{}, len(teamList))
	for _, team := range teamList {
		if len(hints) == messaging.MaximumRuntimeContextTeamHints {
			break
		}
		if team.ID == uuid.Nil || team.Workspace != workspaceID {
			continue
		}
		if _, duplicate := seen[team.ID]; duplicate {
			continue
		}
		if allowed != nil {
			if _, ok := allowed[team.ID]; !ok {
				continue
			}
		}
		seen[team.ID] = struct{}{}
		hints = append(hints, messaging.RuntimeTeamHint{Name: team.Name, Code: team.Code})
	}
	return hints
}

func runtimeTerm(value, fallback string) messaging.RuntimeTerm {
	singular := strings.TrimSpace(value)
	if singular == "" {
		singular = fallback
	}
	return messaging.RuntimeTerm{Singular: singular, Plural: pluralize(singular)}
}

func pluralize(value string) string {
	if strings.EqualFold(value, "focus area") {
		return value + "s"
	}
	runes := []rune(value)
	if len(runes) > 0 && (runes[len(runes)-1] == 'y' || runes[len(runes)-1] == 'Y') {
		return string(runes[:len(runes)-1]) + "ies"
	}
	return value + "s"
}
