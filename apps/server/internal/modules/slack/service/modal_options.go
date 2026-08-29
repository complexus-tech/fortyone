package slack

import (
	"fmt"
	"strings"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
)

const (
	slackSelectMaxOptions       = 100
	slackOptionTextMaxRunes     = 75
	slackOptionValueMaxRunes    = 150
	slackExternalSearchMinRunes = 1
)

func limitedSlackOptions(options []map[string]any) []map[string]any {
	if len(options) <= slackSelectMaxOptions {
		return options
	}
	return options[:slackSelectMaxOptions]
}

func slackTeamOption(team slackdomain.Team) map[string]any {
	return toSlackOption(slackTeamOptionText(team), team.ID.String())
}

func slackTeamOptionText(team slackdomain.Team) string {
	name := strings.TrimSpace(team.Name)
	code := strings.TrimSpace(team.Code)
	switch {
	case name == "":
		return truncateRunes(code, slackOptionTextMaxRunes)
	case code == "":
		return truncateRunes(name, slackOptionTextMaxRunes)
	}

	code = truncateRunes(code, slackOptionTextMaxRunes)
	suffix := fmt.Sprintf(" (%s)", code)
	if len([]rune(suffix)) >= slackOptionTextMaxRunes {
		return truncateRunes(code, slackOptionTextMaxRunes)
	}
	name = strings.TrimSpace(truncateRunes(name, slackOptionTextMaxRunes-len([]rune(suffix))))
	return strings.TrimSpace(name + suffix)
}

func slackTeamSuggestionOptions(teams []slackdomain.Team, query string) []map[string]any {
	query = strings.ToLower(strings.TrimSpace(query))
	options := make([]map[string]any, 0, min(len(teams), slackSelectMaxOptions))
	for _, team := range teams {
		if query != "" &&
			!strings.Contains(strings.ToLower(team.Name), query) &&
			!strings.Contains(strings.ToLower(team.Code), query) {
			continue
		}
		options = append(options, slackTeamOption(team))
		if len(options) == slackSelectMaxOptions {
			break
		}
	}
	return options
}

func slackStatusSuggestionOptions(statuses []slackdomain.Status, query string) []map[string]any {
	query = strings.ToLower(strings.TrimSpace(query))
	options := make([]map[string]any, 0, min(len(statuses)+1, slackSelectMaxOptions))
	if query == "" || strings.Contains(strings.ToLower("Request"), query) {
		options = append(options, toSlackOption("Request", slackRequestStatusValue))
	}
	for _, status := range statuses {
		if query != "" && !strings.Contains(strings.ToLower(status.Name), query) {
			continue
		}
		options = append(options, toSlackOption(status.Name, status.ID.String()))
		if len(options) == slackSelectMaxOptions {
			break
		}
	}
	return options
}
