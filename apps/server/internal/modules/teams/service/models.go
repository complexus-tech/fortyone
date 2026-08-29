package teams

import teamsdomain "github.com/complexus-tech/projects-api/internal/modules/teams/domain"

var (
	ErrTeamCodeExists     = teamsdomain.ErrCodeExists
	ErrTeamMemberExists   = teamsdomain.ErrMemberExists
	ErrTeamNotFound       = teamsdomain.ErrNotFound
	ErrTeamMemberNotFound = teamsdomain.ErrMemberNotFound
)

type CoreTeam = teamsdomain.Team
type CoreListTeamsFilter = teamsdomain.ListFilter
type CoreTeamMemberAIContext = teamsdomain.MemberAIContext
type CorePublicTeamJoin = teamsdomain.PublicTeamJoin
type CoreTeamSelfLeave = teamsdomain.TeamSelfLeave
type DefaultStatus = teamsdomain.DefaultStatus

var DefaultStoryStatuses = teamsdomain.DefaultStoryStatuses
