package mayahttp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	search "github.com/complexus-tech/projects-api/internal/modules/search/service"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/google/uuid"
)

func (h *Handlers) currentRealtimeUser(ctx context.Context, userID uuid.UUID) (AppRealtimeVoiceUser, error) {
	user, err := h.users.GetUser(ctx, userID)
	if err != nil {
		return AppRealtimeVoiceUser{}, fmt.Errorf("get current user: %w", err)
	}
	loc := userLocation(user)
	now := h.now().In(loc)
	name := strings.TrimSpace(user.FullName)
	if name == "" {
		name = user.Username
	}
	return AppRealtimeVoiceUser{
		Name:     name,
		Username: user.Username,
		Timezone: loc.String(),
		Today:    now.Format("2006-01-02"),
		Now:      now.Format("15:04"),
	}, nil
}

func userLocation(user users.CoreUser) *time.Location {
	timezone := strings.TrimSpace(user.Timezone)
	if timezone == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.UTC
	}
	return loc
}

func (h *Handlers) resolveRealtimeAssignee(ctx context.Context, workspaceID, userID uuid.UUID, team *teams.CoreTeam, args AppRealtimeCreateTaskArguments) (*uuid.UUID, string, *AppRealtimeToolResponse, error) {
	assigneeName := strings.TrimSpace(args.AssigneeName)
	if args.AssignToMe || isSelfReference(assigneeName) {
		currentUser, err := h.users.GetUser(ctx, userID)
		if err != nil {
			return nil, "", nil, fmt.Errorf("get current user: %w", err)
		}
		return &userID, displayUserName(currentUser), nil, nil
	}
	if assigneeName == "" {
		return nil, "", nil, nil
	}

	members, err := h.users.List(ctx, workspaceID, users.CoreListUsersFilter{
		TeamID: &team.ID,
		Search: assigneeName,
		Limit:  25,
	})
	if err != nil {
		return nil, "", nil, fmt.Errorf("list team members: %w", err)
	}
	matches := resolveRealtimeMemberMatches(members, assigneeName)
	if len(matches) == 1 {
		member := matches[0]
		memberID := member.ID
		return &memberID, displayUserName(member), nil, nil
	}

	message := fmt.Sprintf("Ask which %s member should be assigned.", team.Name)
	if len(matches) == 0 {
		message = fmt.Sprintf("I could not find %q in %s. Ask for the assignee's name again.", assigneeName, team.Name)
	}
	return nil, "", &AppRealtimeToolResponse{
		Success:       false,
		NeedsAssignee: true,
		Members:       toRealtimeVoiceMembers(members),
		Message:       message,
	}, nil
}

func resolveRealtimeMemberMatches(memberList []users.CoreUser, assigneeName string) []users.CoreUser {
	normalized := normalizeName(assigneeName)
	var exact []users.CoreUser
	for _, member := range memberList {
		if normalizeName(member.FullName) == normalized || normalizeName(member.Username) == normalized || normalizeName(member.Email) == normalized {
			exact = append(exact, member)
		}
	}
	if len(exact) > 0 {
		return exact
	}

	var partial []users.CoreUser
	for _, member := range memberList {
		if strings.Contains(normalizeName(member.FullName), normalized) ||
			strings.Contains(normalizeName(member.Username), normalized) ||
			strings.Contains(normalizeName(member.Email), normalized) {
			partial = append(partial, member)
		}
	}
	return partial
}

func (h *Handlers) resolveRealtimeStoryLink(ctx context.Context, workspaceID, userID uuid.UUID, teamList []teams.CoreTeam, team *teams.CoreTeam, value, label string) (*uuid.UUID, string, *AppRealtimeToolResponse, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, "", nil, nil
	}

	teamsByID := indexTeamsByID(teamList)
	if storyRef, ok := normalizeRealtimeStoryReference(value, team); ok {
		story, err := h.stories.QueryByRef(ctx, workspaceID, storyRef)
		if err == nil {
			storyID := story.ID
			return &storyID, storyReference(story.TeamCode, story.SequenceID), nil, nil
		}
	}

	teamID := (*uuid.UUID)(nil)
	if team != nil {
		teamID = &team.ID
	}
	searchResult, err := h.search.Search(ctx, workspaceID, userID, search.SearchParams{
		Type:     search.SearchTypeStories,
		Query:    value,
		TeamID:   teamID,
		SortBy:   search.SortByRelevance,
		Page:     1,
		PageSize: 5,
	})
	if err != nil {
		return nil, "", nil, fmt.Errorf("search %s story reference: %w", label, err)
	}

	matches := resolveRealtimeStoryMatches(searchResult.Stories, value, teamsByID)
	if len(matches) == 1 {
		story := matches[0]
		storyID := story.ID
		return &storyID, realtimeSearchStoryReference(story, teamsByID), nil, nil
	}

	statusesByID, err := h.statusesByID(ctx, workspaceID, userID)
	if err != nil {
		return nil, "", nil, err
	}
	voiceStories := make([]AppRealtimeVoiceStory, 0, len(searchResult.Stories))
	for _, story := range searchResult.Stories {
		voiceStories = append(voiceStories, toRealtimeVoiceSearchStory(story, teamsByID, statusesByID))
	}

	message := fmt.Sprintf("I could not find %q. Ask for the existing %s reference or title.", value, label)
	if len(voiceStories) > 1 {
		message = fmt.Sprintf("Ask which existing %s the user meant.", label)
	}
	return nil, "", &AppRealtimeToolResponse{
		Success:             false,
		NeedsStoryReference: true,
		Stories:             voiceStories,
		Count:               len(voiceStories),
		Message:             message,
	}, nil
}

func resolveRealtimeStoryMatches(storyList []search.CoreSearchStory, value string, teamsByID map[uuid.UUID]teams.CoreTeam) []search.CoreSearchStory {
	if len(storyList) == 0 {
		return nil
	}

	normalizedValue := normalizeName(value)
	var exact []search.CoreSearchStory
	for _, story := range storyList {
		if normalizeName(story.Title) == normalizedValue || normalizeName(realtimeSearchStoryReference(story, teamsByID)) == normalizedValue {
			exact = append(exact, story)
		}
	}
	if len(exact) > 0 {
		return exact
	}
	if len(storyList) == 1 {
		return storyList
	}
	return storyList
}

func normalizeRealtimeStoryReference(value string, team *teams.CoreTeam) (string, bool) {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return "", false
	}

	var letters strings.Builder
	var digits strings.Builder
	seenDigit := false
	for _, r := range value {
		switch {
		case r >= 'A' && r <= 'Z':
			if seenDigit {
				return "", false
			}
			letters.WriteRune(r)
		case r >= '0' && r <= '9':
			seenDigit = true
			digits.WriteRune(r)
		case r == '-' || unicode.IsSpace(r):
			continue
		default:
			return "", false
		}
	}

	if digits.Len() == 0 {
		return "", false
	}
	sequenceID, err := strconv.Atoi(digits.String())
	if err != nil || sequenceID <= 0 {
		return "", false
	}

	teamCode := letters.String()
	if teamCode == "" {
		if team == nil || strings.TrimSpace(team.Code) == "" {
			return "", false
		}
		teamCode = strings.ToUpper(strings.TrimSpace(team.Code))
	}
	return storyReference(teamCode, sequenceID), true
}

func realtimeSearchStoryReference(story search.CoreSearchStory, teamsByID map[uuid.UUID]teams.CoreTeam) string {
	team, ok := teamsByID[story.Team]
	if !ok {
		return ""
	}
	return storyReference(team.Code, story.SequenceID)
}

func isSelfReference(value string) bool {
	switch normalizeName(value) {
	case "me", "myself", "self", "i":
		return true
	default:
		return false
	}
}

func displayUserName(user users.CoreUser) string {
	name := strings.TrimSpace(user.FullName)
	if name != "" {
		return name
	}
	return user.Username
}

func parseRealtimeDate(value string, loc *time.Location, now time.Time) (*time.Time, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return nil, nil
	}
	if parsed, err := time.ParseInLocation("2006-01-02", raw, loc); err == nil {
		return &parsed, nil
	}
	if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
		local := parsed.In(loc)
		date := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
		return &date, nil
	}

	normalized := normalizeDatePhrase(raw)
	switch normalized {
	case "today":
		date := dateOnly(now, loc)
		return &date, nil
	case "tomorrow":
		date := dateOnly(now.AddDate(0, 0, 1), loc)
		return &date, nil
	case "next week":
		date := dateOnly(now.AddDate(0, 0, 7), loc)
		return &date, nil
	}

	if weekday, ok := weekdayFromPhrase(normalized); ok {
		date := dateOnly(nextWeekday(now, loc, weekday, strings.HasPrefix(normalized, "next ")), loc)
		return &date, nil
	}

	return nil, fmt.Errorf("use YYYY-MM-DD or a relative date like today, tomorrow, this Friday, or next Friday")
}

func normalizeDatePhrase(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	for _, prefix := range []string{"due ", "on ", "by ", "deadline ", "start ", "starting "} {
		value = strings.TrimPrefix(value, prefix)
	}
	return strings.Join(strings.Fields(value), " ")
}

func weekdayFromPhrase(value string) (time.Weekday, bool) {
	value = strings.TrimPrefix(value, "this ")
	value = strings.TrimPrefix(value, "next ")
	weekdays := map[string]time.Weekday{
		"sunday":    time.Sunday,
		"monday":    time.Monday,
		"tuesday":   time.Tuesday,
		"wednesday": time.Wednesday,
		"thursday":  time.Thursday,
		"friday":    time.Friday,
		"saturday":  time.Saturday,
	}
	weekday, ok := weekdays[value]
	return weekday, ok
}

func nextWeekday(now time.Time, loc *time.Location, weekday time.Weekday, forceNext bool) time.Time {
	local := now.In(loc)
	days := (int(weekday) - int(local.Weekday()) + 7) % 7
	if forceNext {
		if days == 0 {
			days = 7
		} else {
			days += 7
		}
	}
	return local.AddDate(0, 0, days)
}

func dateOnly(value time.Time, loc *time.Location) time.Time {
	local := value.In(loc)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
}

func toRealtimeVoiceMembers(memberList []users.CoreUser) []AppRealtimeVoiceMember {
	result := make([]AppRealtimeVoiceMember, len(memberList))
	for i, member := range memberList {
		name := strings.TrimSpace(member.FullName)
		if name == "" {
			name = member.Username
		}
		roleTitle := strings.TrimSpace(member.TeamAIRoleTitle)
		if roleTitle == "" {
			roleTitle = strings.TrimSpace(member.InferredTeamAIRoleTitle)
		}
		role := ""
		if member.Role != nil {
			role = *member.Role
		}
		result[i] = AppRealtimeVoiceMember{
			Name:         name,
			Username:     member.Username,
			Role:         role,
			RoleTitle:    roleTitle,
			LastActiveAt: member.LastStoryActivityAt,
		}
	}
	return result
}

func toRealtimeVoiceObjective(objective objectives.CoreObjective, teamsByID map[uuid.UUID]teams.CoreTeam) AppRealtimeVoiceObjective {
	teamName := ""
	if team, ok := teamsByID[objective.Team]; ok {
		teamName = team.Name
	}
	health := ""
	if objective.Health != nil {
		health = string(*objective.Health)
	}
	return AppRealtimeVoiceObjective{
		Name:             objective.Name,
		Description:      objective.Description,
		Team:             teamName,
		Priority:         objective.Priority,
		Health:           health,
		StartDate:        objective.StartDate,
		EndDate:          objective.EndDate,
		TotalStories:     objective.TotalStories,
		CompletedStories: objective.CompletedStories,
	}
}

func toRealtimeVoiceKeyResult(keyResult keyresults.CoreKeyResultWithObjective) AppRealtimeVoiceKeyResult {
	return AppRealtimeVoiceKeyResult{
		Name:            keyResult.Name,
		ObjectiveName:   keyResult.ObjectiveName,
		Team:            keyResult.TeamName,
		MeasurementType: keyResult.MeasurementType,
		StartValue:      keyResult.StartValue,
		CurrentValue:    keyResult.CurrentValue,
		TargetValue:     keyResult.TargetValue,
		StartDate:       keyResult.StartDate,
		EndDate:         keyResult.EndDate,
	}
}

func toRealtimeVoiceSearchStory(story search.CoreSearchStory, teamsByID map[uuid.UUID]teams.CoreTeam, statusesByID map[uuid.UUID]states.CoreState) AppRealtimeVoiceStory {
	teamName, teamCode := "", ""
	if team, ok := teamsByID[story.Team]; ok {
		teamName = team.Name
		teamCode = team.Code
	}
	var status *AppRealtimeVoiceStatus
	if story.Status != nil {
		if matchedStatus, ok := statusesByID[*story.Status]; ok {
			status = toRealtimeVoiceStatus(matchedStatus)
		}
	}
	return AppRealtimeVoiceStory{
		Reference: storyReference(teamCode, story.SequenceID),
		Title:     story.Title,
		Priority:  story.Priority,
		Team:      teamName,
		Status:    status,
		StartDate: story.StartDate,
		EndDate:   story.EndDate,
	}
}

func toRealtimeVoiceSearchObjective(objective search.CoreSearchObjective, teamsByID map[uuid.UUID]teams.CoreTeam) AppRealtimeVoiceObjective {
	teamName := ""
	if team, ok := teamsByID[objective.Team]; ok {
		teamName = team.Name
	}
	health := ""
	if objective.Health != nil {
		health = *objective.Health
	}
	return AppRealtimeVoiceObjective{
		Name:        objective.Name,
		Description: objective.Description,
		Team:        teamName,
		Priority:    objective.Priority,
		Health:      health,
		StartDate:   objective.StartDate,
		EndDate:     objective.EndDate,
	}
}

func indexTeamsByID(teamList []teams.CoreTeam) map[uuid.UUID]teams.CoreTeam {
	teamsByID := make(map[uuid.UUID]teams.CoreTeam, len(teamList))
	for _, team := range teamList {
		teamsByID[team.ID] = team
	}
	return teamsByID
}

func (h *Handlers) statusesByID(ctx context.Context, workspaceID, userID uuid.UUID) (map[uuid.UUID]states.CoreState, error) {
	statusList, err := h.states.List(ctx, workspaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("list statuses: %w", err)
	}
	statusesByID := make(map[uuid.UUID]states.CoreState, len(statusList))
	for _, status := range statusList {
		statusesByID[status.ID] = status
	}
	return statusesByID, nil
}

func (h *Handlers) invalidateStoryListCaches(ctx context.Context, workspaceID uuid.UUID) {
	if h.cache == nil {
		return
	}

	storyListCachePattern := fmt.Sprintf(cache.StoryListKey+"*", workspaceID.String())
	if err := h.cache.DeleteByPattern(ctx, storyListCachePattern); err != nil {
		h.log.Warn(ctx, "failed to invalidate maya story list cache", "workspace_id", workspaceID, "error", err)
	}

	myStoriesCachePattern := fmt.Sprintf(cache.MyStoriesKey+"*", workspaceID.String())
	if err := h.cache.DeleteByPattern(ctx, myStoriesCachePattern); err != nil {
		h.log.Warn(ctx, "failed to invalidate maya assigned story cache", "workspace_id", workspaceID, "error", err)
	}
}

func safetyIdentifier(userID uuid.UUID) string {
	sum := sha256.Sum256([]byte(userID.String()))
	return hex.EncodeToString(sum[:])
}
