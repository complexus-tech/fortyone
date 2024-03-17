package issuesrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/issues"
	"github.com/google/uuid"
)

// dbIssue represents the database model for an dbIssue.
type dbIssue struct {
	ID              uuid.UUID  `db:"id"`
	SequenceID      int        `db:"sequence_id"`
	Title           string     `db:"title"`
	Description     string     `db:"description"`
	DescriptionHTML string     `db:"description_html"`
	Parent          *uuid.UUID `db:"parent_id"`
	Project         *uuid.UUID `db:"project_id"`
	Status          *uuid.UUID `db:"status_id"`
	Assignee        *uuid.UUID `db:"assignee_id"`
	BlockedBy       *uuid.UUID `db:"blocked_by_id"`
	Blocking        *uuid.UUID `db:"blocking_id"`
	Related         *uuid.UUID `db:"related_id"`
	Reporter        *uuid.UUID `db:"reporter_id"`
	Priority        *uuid.UUID `db:"priority_id"`
	Attachments     string     `db:"attachments"`
	Sprint          *uuid.UUID `db:"sprint_id"`
	Watchers        string     `db:"watchers"`
	Labels          string     `db:"labels"`
	Comments        string     `db:"comments"`
	StartDate       *time.Time `db:"start_date"`
	EndDate         *time.Time `db:"end_date"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
	DeletedAt       *time.Time `db:"deleted_at"`
}

// toCoreIssue converts a dbIssue to a CoreSingleIssue.
func toCoreIssue(i dbIssue) issues.CoreSingleIssue {
	return issues.CoreSingleIssue{
		ID:              i.ID,
		SequenceID:      i.SequenceID,
		Title:           i.Title,
		Description:     i.Description,
		Parent:          i.Parent,
		Project:         i.Project,
		Status:          i.Status,
		Assignee:        i.Assignee,
		BlockedBy:       i.BlockedBy,
		Blocking:        i.Blocking,
		Related:         i.Related,
		Reporter:        i.Reporter,
		DescriptionHTML: i.DescriptionHTML,
		Priority:        i.Priority,
		Sprint:          i.Sprint,
		StartDate:       i.StartDate,
		EndDate:         i.EndDate,
		CreatedAt:       i.CreatedAt,
		UpdatedAt:       i.UpdatedAt,
		DeletedAt:       i.DeletedAt,
	}
}

// toCoreIssues converts a slice of dbIssues to a slice of CoreIssueList.
func toCoreIssues(is []dbIssue) []issues.CoreIssueList {
	cl := make([]issues.CoreIssueList, len(is))
	for i, issue := range is {
		cl[i] = issues.CoreIssueList{
			ID:         issue.ID,
			SequenceID: issue.SequenceID,
			Title:      issue.Title,
			Parent:     issue.Parent,
			Project:    issue.Project,
			Status:     issue.Status,
			Assignee:   issue.Assignee,
			StartDate:  issue.StartDate,
			EndDate:    issue.EndDate,
			Priority:   issue.Priority,
			Sprint:     issue.Sprint,
		}
	}
	return cl
}
