package desktophttp

import (
	"fmt"
	"time"

	documents "github.com/complexus-tech/projects-api/internal/modules/documents/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

type Story struct {
	ID              uuid.UUID  `json:"id"`
	Title           string     `json:"title"`
	Reference       string     `json:"reference"`
	TeamID          uuid.UUID  `json:"teamId"`
	WorkspaceID     uuid.UUID  `json:"workspaceId"`
	StatusID        *uuid.UUID `json:"statusId"`
	AssigneeID      *uuid.UUID `json:"assigneeId"`
	Priority        string     `json:"priority"`
	Description     *string    `json:"description"`
	DescriptionHTML *string    `json:"descriptionHTML"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	CreatedAt       time.Time  `json:"createdAt"`
}

func mapStory(story stories.CoreSingleStory) Story {
	return Story{ID: story.ID, Title: story.Title, Reference: fmt.Sprintf("%s-%d", story.TeamCode, story.SequenceID), TeamID: story.Team, WorkspaceID: story.Workspace, StatusID: story.Status, AssigneeID: story.Assignee, Priority: story.Priority, Description: story.Description, DescriptionHTML: story.DescriptionHTML, UpdatedAt: story.UpdatedAt, CreatedAt: story.CreatedAt}
}

type Document struct {
	ID                 uuid.UUID `json:"id"`
	WorkspaceID        uuid.UUID `json:"workspaceId"`
	Title              string    `json:"title"`
	ContentHTML        string    `json:"contentHtml"`
	ContentText        string    `json:"contentText"`
	Visibility         string    `json:"visibility"`
	Revision           int64     `json:"revision"`
	CollaborationState []byte    `json:"collaborationState,omitempty"`
	CollaborationEpoch int64     `json:"collaborationEpoch"`
	Collaborative      bool      `json:"collaborative"`
	CanEdit            bool      `json:"canEdit"`
	UpdatedAt          time.Time `json:"updatedAt"`
	CreatedAt          time.Time `json:"createdAt"`
}

func mapDocument(document documents.CoreDocument) Document {
	return Document{ID: document.ID, WorkspaceID: document.WorkspaceID, Title: document.Title, ContentHTML: document.ContentHTML, ContentText: document.ContentText, Visibility: string(document.Visibility), Revision: document.Revision, CollaborationEpoch: document.CollaborationEpoch, CollaborationState: document.CollaborationState, Collaborative: document.Collaborative, CanEdit: document.CanEdit, UpdatedAt: document.UpdatedAt, CreatedAt: document.CreatedAt}
}

type Comment struct {
	ID          uuid.UUID  `json:"id"`
	StoryID     uuid.UUID  `json:"storyId"`
	UserID      uuid.UUID  `json:"userId"`
	ParentID    *uuid.UUID `json:"parentId"`
	Comment     string     `json:"comment"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	SubComments []Comment  `json:"subComments"`
}

func mapComment(comment stories.CoreComment) Comment {
	children := make([]Comment, len(comment.SubComments))
	for i, child := range comment.SubComments {
		children[i] = mapComment(child)
	}
	return Comment{ID: comment.ID, StoryID: comment.StoryID, UserID: comment.UserID, ParentID: comment.Parent, Comment: comment.Comment, CreatedAt: comment.CreatedAt, UpdatedAt: comment.UpdatedAt, SubComments: children}
}

type Team struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Code  string    `json:"code"`
	Color string    `json:"color"`
}
type Status struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	TeamID     uuid.UUID `json:"teamId"`
	Category   string    `json:"category"`
	Color      string    `json:"color"`
	OrderIndex int       `json:"orderIndex"`
	IsDefault  bool      `json:"isDefault"`
}
