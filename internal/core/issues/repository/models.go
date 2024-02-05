package issues

import "time"

// Issue represents a project issue.
type Issue struct {
	IssueID         int       `db:"issue_id"`
	Title           string    `db:"title"`
	Description     string    `db:"description"`
	Parent          int       `db:"parent_id"`
	Project         int       `db:"project_id"`
	Status          int       `db:"status_id"`
	Assignee        int       `db:"assignee_id"`
	Reporter        int       `db:"reporter_id"`
	DescriptionHTML string    `db:"description_html"`
	Type            string    `db:"type"`
	Priority        string    `db:"priority_id"`
	Attachments     string    `db:"attachments"`
	Sprint          string    `db:"sprint_id"`
	Team            string    `db:"team_id"`
	Watchers        string    `db:"watchers"`
	Labels          string    `db:"labels"`
	Comments        string    `db:"comments"`
	StartDate       time.Time `db:"start_date"`
	EndDate         time.Time `db:"end_date"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
	DeletedAt       time.Time `db:"deleted_at"`
}

// IssueStatus represents the status of an issue.
type IssueStatus struct {
	StatusID int
	Name     string
}

type IssueType struct {
	TypeID int
	Name   string
}

type IssuePriority struct {
	PriorityID int
	Name       string
	Order      int
}

type IssueComment struct {
	CommentID   int
	IssueID     int
	CommenterID int
	Comment     string
	CreatedDate time.Time
	UpdatedDate time.Time
	Commenter   string
}

type IssueAttachment struct {
	AttachmentID int
	IssueID      int
	FileName     string
	FileType     string
	FileSize     int
	FilePath     string
	CreatedDate  time.Time
	UploaderID   int
	Uploader     string
}

type IssueWatcher struct {
	WatcherID int
	IssueID   int
	UserID    int
	Watcher   string
}

type IssueAssignee struct {
	AssigneeID int
	IssueID    int
	UserID     int
	Assignee   string
}

type IssueSprint struct {
	SprintID   int
	IssueID    int
	SprintName string
}

type IssueTeam struct {
	TeamID   int
	IssueID  int
	TeamName string
}

type IssueProject struct {
	ProjectID   int
	ProjectName string
}

type IssueUser struct {
	UserID   int
	UserName string
}
