package mayahttp

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	"github.com/google/uuid"
)

func (h *Handlers) realtimeTerminology(ctx context.Context, workspaceID uuid.UUID) AppRealtimeTerminology {
	settings := workspaces.CoreWorkspaceSettings{
		StoryTerm:     "story",
		SprintTerm:    "sprint",
		ObjectiveTerm: "objective",
		KeyResultTerm: "key result",
	}
	if h.workspaces != nil {
		if fetched, err := h.workspaces.GetOrCreateWorkspaceSettings(ctx, workspaceID); err == nil {
			settings = fetched
		}
	}

	return AppRealtimeTerminology{
		Story:      settings.StoryTerm,
		Stories:    pluralizeTerm(settings.StoryTerm),
		Sprint:     settings.SprintTerm,
		Sprints:    pluralizeTerm(settings.SprintTerm),
		Objective:  settings.ObjectiveTerm,
		Objectives: pluralizeTerm(settings.ObjectiveTerm),
		KeyResult:  settings.KeyResultTerm,
		KeyResults: pluralizeTerm(settings.KeyResultTerm),
	}
}

func pluralizeTerm(term string) string {
	term = strings.TrimSpace(term)
	if term == "" {
		return term
	}
	if strings.HasSuffix(term, "y") {
		return strings.TrimSuffix(term, "y") + "ies"
	}
	if term == "focus area" {
		return "focus areas"
	}
	return term + "s"
}

func termForCount(singular, plural string, count int) string {
	if count == 1 {
		return singular
	}
	return plural
}

func pluralSuffix(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

func teamContextMessage(teamList []teams.CoreTeam) string {
	switch len(teamList) {
	case 0:
		return "The user does not belong to any teams."
	case 1:
		return fmt.Sprintf("The user belongs to one team: %s. Default to this team when team selection is needed.", teamList[0].Name)
	default:
		return fmt.Sprintf("The user belongs to %d teams. Resolve close team name matches by name or code, and ask only when ambiguous.", len(teamList))
	}
}

func teamSelectionInstruction(teamList []teams.CoreTeam) string {
	switch len(teamList) {
	case 0:
		return "The user currently belongs to no teams, so team-scoped actions may not be possible."
	case 1:
		return fmt.Sprintf("If team selection is needed and the user does not specify a team, use %s.", teamList[0].Name)
	default:
		return fmt.Sprintf("The user belongs to these teams: %s. Resolve close team matches by name or code; ask which team only when the match is missing or ambiguous.", strings.Join(teamNames(teamList), ", "))
	}
}

func teamNames(teamList []teams.CoreTeam) []string {
	names := make([]string, len(teamList))
	for i, team := range teamList {
		if strings.TrimSpace(team.Code) == "" {
			names[i] = team.Name
			continue
		}
		names[i] = fmt.Sprintf("%s (%s)", team.Name, team.Code)
	}
	return names
}

func clampLimit(limit, fallback int) int {
	if fallback <= 0 {
		fallback = 10
	}
	if limit <= 0 {
		return fallback
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func ptr[T any](value T) *T {
	return &value
}

func normalizePriority(priority string) string {
	priority = strings.TrimSpace(priority)
	if priority == "" {
		return "No Priority"
	}
	if _, ok := realtimeStoryPriorities[priority]; ok {
		return priority
	}
	return "No Priority"
}

func optionalString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func normalizeName(value string) string {
	var b strings.Builder
	lastWasSpace := true
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastWasSpace = false
		case unicode.IsSpace(r):
			if !lastWasSpace {
				b.WriteByte(' ')
				lastWasSpace = true
			}
		}
	}
	return strings.TrimSpace(b.String())
}

func resolveRealtimeTeam(teamList []teams.CoreTeam, teamName string) *teams.CoreTeam {
	if strings.TrimSpace(teamName) == "" {
		if len(teamList) != 1 {
			return nil
		}
		return &teamList[0]
	}

	normalizedTeamName := normalizeName(teamName)
	for i := range teamList {
		if normalizeName(teamList[i].Name) == normalizedTeamName || normalizeName(teamList[i].Code) == normalizedTeamName {
			return &teamList[i]
		}
	}

	var matches []int
	for i := range teamList {
		if strings.Contains(normalizeName(teamList[i].Name), normalizedTeamName) || strings.Contains(normalizeName(teamList[i].Code), normalizedTeamName) {
			matches = append(matches, i)
		}
	}
	if len(matches) != 1 {
		return nil
	}
	return &teamList[matches[0]]
}

func findDefaultRealtimeStatus(statuses []states.CoreState) *states.CoreState {
	if len(statuses) == 0 {
		return nil
	}
	for i := range statuses {
		if statuses[i].IsDefault {
			return &statuses[i]
		}
	}
	for _, category := range []string{"unstarted", "backlog"} {
		for i := range statuses {
			if statuses[i].Category == category {
				return &statuses[i]
			}
		}
	}
	return &statuses[0]
}

func toRealtimeVoiceTeams(teamList []teams.CoreTeam) []AppRealtimeVoiceTeam {
	result := make([]AppRealtimeVoiceTeam, len(teamList))
	for i, team := range teamList {
		result[i] = AppRealtimeVoiceTeam{
			Name:        team.Name,
			Code:        team.Code,
			MemberCount: team.MemberCount,
			IsPrivate:   team.IsPrivate,
		}
	}
	return result
}

func toRealtimeVoiceStatus(status states.CoreState) *AppRealtimeVoiceStatus {
	return &AppRealtimeVoiceStatus{
		Name:     status.Name,
		Category: status.Category,
	}
}

func storyReference(teamCode string, sequenceID int) string {
	teamCode = strings.TrimSpace(teamCode)
	if teamCode == "" || sequenceID <= 0 {
		return ""
	}
	return fmt.Sprintf("%s-%d", teamCode, sequenceID)
}

func toRealtimeVoiceStory(story stories.CoreStoryList, statusesByID map[uuid.UUID]states.CoreState) AppRealtimeVoiceStory {
	var teamName, teamCode string
	if story.TeamSummary != nil {
		teamName = story.TeamSummary.Name
		teamCode = story.TeamSummary.Code
	}

	var status *AppRealtimeVoiceStatus
	if story.Status != nil {
		if matchedStatus, ok := statusesByID[*story.Status]; ok {
			status = toRealtimeVoiceStatus(matchedStatus)
		}
	}

	return AppRealtimeVoiceStory{
		Reference:     storyReference(teamCode, story.SequenceID),
		Title:         story.Title,
		Priority:      story.Priority,
		EstimateLabel: story.EstimateLabel,
		EstimateValue: story.EstimateValue,
		Team:          teamName,
		Status:        status,
		StartDate:     story.StartDate,
		EndDate:       story.EndDate,
		CompletedAt:   story.CompletedAt,
	}
}

func toRealtimeVoiceCreatedStory(story stories.CoreSingleStory, team teams.CoreTeam, status states.CoreState, assigneeName string) AppRealtimeVoiceStory {
	return AppRealtimeVoiceStory{
		Reference:     storyReference(team.Code, story.SequenceID),
		Title:         story.Title,
		Priority:      story.Priority,
		EstimateLabel: story.EstimateLabel,
		EstimateValue: story.EstimateValue,
		Team:          team.Name,
		Assignee:      assigneeName,
		Status:        toRealtimeVoiceStatus(status),
		StartDate:     story.StartDate,
		EndDate:       story.EndDate,
		CompletedAt:   story.CompletedAt,
	}
}
