package slack

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func findStatusByID(statuses []slackStatusRecord, statusID uuid.UUID) (slackStatusRecord, bool) {
	for _, status := range statuses {
		if status.ID == statusID {
			return status, true
		}
	}
	return slackStatusRecord{}, false
}

func normalizeSlackPriority(value string) string {
	switch strings.TrimSpace(value) {
	case "Low", "Medium", "High", "Urgent", slackPriorityNoPriority:
		return strings.TrimSpace(value)
	default:
		return slackPriorityNoPriority
	}
}

func selectTeam(teams []slackTeamRecord, preferredTeamID uuid.UUID) slackTeamRecord {
	if preferredTeamID != uuid.Nil {
		for _, team := range teams {
			if team.ID == preferredTeamID {
				return team
			}
		}
	}
	return teams[0]
}

func teamMemberDisplayName(member slackTeamMemberRecord) string {
	if fullName := strings.TrimSpace(member.FullName); fullName != "" {
		if email := strings.TrimSpace(member.Email); email != "" {
			return fmt.Sprintf("%s (%s)", fullName, email)
		}
		return fullName
	}
	if username := strings.TrimSpace(member.Username); username != "" {
		if email := strings.TrimSpace(member.Email); email != "" {
			return fmt.Sprintf("%s (%s)", username, email)
		}
		return username
	}
	if email := strings.TrimSpace(member.Email); email != "" {
		return email
	}
	return member.UserID.String()
}

func storyCreatorDisplayName(member slackTeamMemberRecord) string {
	if fullName := strings.TrimSpace(member.FullName); fullName != "" {
		return fullName
	}
	if username := strings.TrimSpace(member.Username); username != "" {
		return username
	}
	if email := strings.TrimSpace(member.Email); email != "" {
		return email
	}
	return "A team member"
}

func teamMemberExists(members []slackTeamMemberRecord, userID uuid.UUID) bool {
	for _, member := range members {
		if member.UserID == userID {
			return true
		}
	}
	return false
}

func filterValidLabelIDs(labels []slackLabelRecord, selected []uuid.UUID) []uuid.UUID {
	if len(selected) == 0 {
		return nil
	}
	valid := make(map[uuid.UUID]struct{}, len(labels))
	for _, label := range labels {
		valid[label.ID] = struct{}{}
	}
	result := make([]uuid.UUID, 0, len(selected))
	seen := make(map[uuid.UUID]struct{}, len(selected))
	for _, labelID := range selected {
		if _, alreadySeen := seen[labelID]; alreadySeen {
			continue
		}
		seen[labelID] = struct{}{}
		if _, ok := valid[labelID]; ok {
			result = append(result, labelID)
		}
	}
	return result
}

func uuidStrings(values []uuid.UUID) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.String())
	}
	return result
}

func optionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeEmail(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
