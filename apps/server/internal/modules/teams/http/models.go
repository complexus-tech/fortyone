package teamshttp

import (
	"encoding/json"
	"fmt"
	"time"

	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	"github.com/google/uuid"
)

// AppTeamList represents a team in the application layer.
type AppTeamsList struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Code           string    `json:"code"`
	Color          string    `json:"color"`
	IsPrivate      bool      `json:"isPrivate"`
	Workspace      uuid.UUID `json:"workspaceId"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	MemberCount    int       `json:"memberCount"`
	SprintsEnabled bool      `json:"sprintsEnabled"`
	StoryTerm      *string   `json:"storyTerm"`
}

type AppPagination struct {
	Page     int  `json:"page"`
	PageSize int  `json:"pageSize"`
	HasMore  bool `json:"hasMore"`
	NextPage int  `json:"nextPage"`
}

type AppTeamsResponse struct {
	Teams      []AppTeamsList `json:"teams"`
	Pagination AppPagination  `json:"pagination"`
}

// toAppTeams converts a list of core teams to a list of application teams.
func toAppTeams(teams []teams.CoreTeam) []AppTeamsList {
	appTeams := make([]AppTeamsList, len(teams))
	for i, team := range teams {
		appTeams[i] = AppTeamsList{
			ID:             team.ID,
			Name:           team.Name,
			Code:           team.Code,
			Color:          team.Color,
			IsPrivate:      team.IsPrivate,
			Workspace:      team.Workspace,
			CreatedAt:      team.CreatedAt,
			UpdatedAt:      team.UpdatedAt,
			MemberCount:    team.MemberCount,
			SprintsEnabled: team.SprintsEnabled,
			StoryTerm:      team.StoryTerm,
		}
	}
	return appTeams
}

func toAppTeamsResponse(teams []teams.CoreTeam, page, pageSize int, hasMore bool) AppTeamsResponse {
	nextPage := 0
	if hasMore {
		nextPage = page + 1
	}

	return AppTeamsResponse{
		Teams: toAppTeams(teams),
		Pagination: AppPagination{
			Page:     page,
			PageSize: pageSize,
			HasMore:  hasMore,
			NextPage: nextPage,
		},
	}
}

type AppNewTeam struct {
	Name      string `json:"name" validate:"required"`
	Code      string `json:"code" validate:"required"`
	Color     string `json:"color" validate:"required"`
	IsPrivate bool   `json:"isPrivate"`
}

type AppUpdateTeam struct {
	Name      string            `json:"name,omitempty"`
	Code      string            `json:"code,omitempty"`
	Color     string            `json:"color,omitempty"`
	IsPrivate *bool             `json:"isPrivate,omitempty"`
	StoryTerm OptionalStoryTerm `json:"storyTerm,omitempty"`
}

// A value object receives JSON null, whereas a pointer cannot distinguish
// null from an omitted property. Both are meaningful in a partial update.
type OptionalStoryTerm struct {
	Present bool
	Value   *string
}

func (term *OptionalStoryTerm) UnmarshalJSON(data []byte) error {
	var value *string
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("work naming must be a string or null: %w", err)
	}
	term.Present, term.Value = true, value
	return nil
}

type AppNewTeamMember struct {
	UserID uuid.UUID `json:"userId" validate:"required"`
}

type AppUpdateTeamMemberAIContext struct {
	RoleTitle       string `json:"roleTitle"`
	RoleDescription string `json:"roleDescription"`
}

type AppUpdateTeamOrdering struct {
	TeamIDs []uuid.UUID `json:"teamIds" validate:"required,min=1"`
}
