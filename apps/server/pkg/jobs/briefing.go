package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	"github.com/complexus-tech/projects-api/pkg/emailthread"
	"github.com/complexus-tech/projects-api/pkg/mailer"
)

// BriefingSources retains each category's existing, permission- and preference-
// filtered queries. Combining delivery must never broaden an email's audience.
type BriefingSources struct {
	Stories    OverdueStoryStore
	Objectives ObjectiveOverdueStore
	Weekly     WeeklyDigestStore
}

type BriefingContent struct {
	Sections []mailer.Digest
	Targets  []emailthread.TargetContext
}

// BuildBriefing consolidates daily deadlines and Monday's planning guidance.
// asOf is the recipient's calendar date, represented at UTC midnight for SQL date comparisons.
func (s BriefingSources) BuildBriefing(ctx context.Context, recipient notifications.RoutineRecipient, asOf time.Time) (BriefingContent, error) {
	var content BriefingContent
	if s.Stories == nil || s.Objectives == nil || s.Weekly == nil {
		return content, errors.New("briefing sources are required")
	}
	base := "https://" + recipient.WorkspaceSlug + ".fortyone.app"
	stories, err := s.Stories.ListOverdueStoryGuidanceItems(ctx, asOf, recipient.UserID, recipient.WorkspaceID)
	if err != nil {
		return content, fmt.Errorf("load briefing stories: %w", err)
	}
	if len(stories) > 0 {
		section := mailer.Digest{Intro: fmt.Sprintf("%d %s need your attention.", len(stories), pluralize(len(stories), "story", "stories"))}
		for _, story := range stories[:min(len(stories), mailer.DigestDetailLimit)] {
			section.Rows = append(section.Rows, mailer.DigestRow{Label: story.Title, Text: overdueStoryEmailCopyFact(story), URL: overdueStoryURL(base, story), Icon: "calendar"})
			content.Targets = append(content.Targets, emailthread.TargetContext{Kind: "story", ID: story.ID, TeamID: story.TeamID, DisplayName: story.Title})
		}
		section.Rows = appendBriefingMore(section.Rows, len(stories), base+"/my-work?tab=assigned", "stories")
		content.Sections = append(content.Sections, section)
	}
	if asOf.Weekday() != time.Monday {
		return content, nil
	}
	objectiveItems, err := s.Objectives.ListOverdueObjectiveGuidanceItems(ctx, asOf, recipient.UserID, recipient.WorkspaceID)
	if err != nil {
		return content, fmt.Errorf("load briefing objectives: %w", err)
	}
	var objectiveRows []mailer.DigestRow
	var targets []emailthread.TargetContext
	for _, objective := range objectiveItems {
		url := fmt.Sprintf("%s/teams/%s/objectives/%s", base, objective.TeamID, objective.ID)
		if objective.DeadlineStatus != "" && objective.DeadlineStatus != "not_due" && objective.DeadlineStatus != "other" && objective.DeadlineStatus != "future" && !objective.EndDate.IsZero() {
			objectiveRows = append(objectiveRows, mailer.DigestRow{Label: objective.Name, Text: objectiveEmailCopyFactText(objective), URL: url, Icon: "calendar"})
			targets = append(targets, emailthread.TargetContext{Kind: "objective", ID: objective.ID, TeamID: objective.TeamID, DisplayName: objective.Name})
		}
		var results []objectives.OverdueGuidanceKeyResult
		if objective.KeyResults != "" {
			if err := json.Unmarshal([]byte(objective.KeyResults), &results); err != nil {
				return content, fmt.Errorf("decode briefing key results: %w", err)
			}
		}
		for _, result := range results {
			if result.IsCompleted {
				continue
			}
			objectiveRows = append(objectiveRows, mailer.DigestRow{Label: result.Name, Text: keyResultEmailCopyFactText(result), URL: url, Icon: "calendar"})
			targets = append(targets, emailthread.TargetContext{Kind: "key_result", ID: result.ID, TeamID: objective.TeamID, ParentID: objective.ID, DisplayName: result.Name})
		}
	}
	if len(objectiveRows) > 0 {
		visible := min(len(objectiveRows), mailer.DigestDetailLimit)
		content.Sections = append(content.Sections, mailer.Digest{Intro: "Objectives and key results to review this week.", Rows: appendBriefingMore(objectiveRows[:visible:visible], len(objectiveRows), base+"/objectives", "objectives and key results")})
		content.Targets = append(content.Targets, targets[:visible]...)
	}
	if recipient.WeeklyEnabled {
		stats, err := s.Weekly.GetWeeklyDigestStats(ctx, notifications.WeeklyDigestStatsQuery{UserID: recipient.UserID, WorkspaceID: recipient.WorkspaceID, AsOf: asOf})
		if err != nil {
			return content, fmt.Errorf("load briefing weekly overview: %w", err)
		}
		// Notification details are supplied by the same unread/unsent batch as
		// activity mail; do not repeat them as mentions and comments here.
		var overview []string
		if stats.OverdueStories > 0 {
			overview = append(overview, fmt.Sprintf("%d %s overdue", stats.OverdueStories, pluralize(stats.OverdueStories, "story is", "stories are")))
		}
		if stats.DueThisWeekStories > 0 {
			overview = append(overview, fmt.Sprintf("%d %s due this week", stats.DueThisWeekStories, pluralize(stats.DueThisWeekStories, "story is", "stories are")))
		}
		if len(overview) > 0 {
			content.Sections = append([]mailer.Digest{{Intro: "Your week ahead", Rows: []mailer.DigestRow{{Text: strings.Join(overview, "; ") + ".", Label: "Review your assigned work", URL: base + "/my-work?tab=assigned"}}}}, content.Sections...)
		}
	}
	return content, nil
}

func appendBriefingMore(rows []mailer.DigestRow, total int, url, noun string) []mailer.DigestRow {
	if remaining := total - len(rows); remaining > 0 {
		rows = append(rows, mailer.DigestRow{More: true, Text: fmt.Sprintf("View %d more %s →", remaining, noun), URL: url})
	}
	return rows
}
