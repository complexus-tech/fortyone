package issuesrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/issues"
	"github.com/google/uuid"
)

// Issue represents a project issue.
type issue struct {
	ID              int        `db:"id"`
	Title           string     `db:"title"`
	Description     string     `db:"description"`
	Parent          int        `db:"parent_id"`
	Project         int        `db:"project_id"`
	Status          int        `db:"status_id"`
	Assignee        int        `db:"assignee_id"`
	Reporter        uuid.UUID  `db:"reporter_id"`
	DescriptionHTML string     `db:"description_html"`
	Type            string     `db:"type"`
	Priority        string     `db:"priority_id"`
	Attachments     string     `db:"attachments"`
	Sprint          string     `db:"sprint_id"`
	Team            string     `db:"team_id"`
	Watchers        string     `db:"watchers"`
	Labels          string     `db:"labels"`
	Comments        string     `db:"comments"`
	StartDate       time.Time  `db:"start_date"`
	EndDate         time.Time  `db:"end_date"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
	DeletedAt       *time.Time `db:"deleted_at"`
}

func toCoreIssue(i issue) issues.Issue {
	return issues.Issue{
		ID:          i.ID,
		Title:       i.Title,
		Description: i.Description,
	}
}

func toCoreIssues(is []issue) []issues.Issue {
	issues := make([]issues.Issue, len(is))
	for i, issue := range is {
		issues[i] = toCoreIssue(issue)
	}
	return issues
}
