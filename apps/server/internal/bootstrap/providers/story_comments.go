package providers

import (
	"context"
	"errors"

	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

// StoryCommentCreator adapts the concrete comments service to the narrow port
// owned by the stories service. Cross-module command and result mapping belongs
// in bootstrap, not in either feature module.
type StoryCommentCreator struct {
	comments *comments.Service
}

func NewStoryCommentCreator(service *comments.Service) (*StoryCommentCreator, error) {
	if service == nil {
		return nil, errors.New("comment service is required")
	}
	return &StoryCommentCreator{comments: service}, nil
}

func (adapter *StoryCommentCreator) CreateComment(
	ctx context.Context,
	command stories.CreateCommentCommand,
) (stories.CoreComment, error) {
	created, err := adapter.comments.CreateComment(ctx, comments.CreateCommentCommand{
		WorkspaceID: command.WorkspaceID,
		StoryID:     command.StoryID,
		ParentID:    copyCommentParentID(command.ParentID),
		Actor:       command.Actor,
		Content:     command.Content,
		MentionedUserIDs: append(
			[]uuid.UUID(nil),
			command.MentionedUserIDs...,
		),
	})
	if err != nil {
		return stories.CoreComment{}, err
	}
	return storyCommentToStories(created), nil
}

func copyCommentParentID(parentID *uuid.UUID) *uuid.UUID {
	if parentID == nil {
		return nil
	}
	copy := *parentID
	return &copy
}

func storyCommentToStories(comment comments.CoreComment) stories.CoreComment {
	children := make([]stories.CoreComment, len(comment.SubComments))
	for index, child := range comment.SubComments {
		children[index] = storyCommentToStories(child)
	}
	return stories.CoreComment{
		ID: comment.ID, StoryID: comment.StoryID, Parent: copyCommentParentID(comment.Parent),
		UserID: comment.UserID, Comment: comment.Comment,
		CreatedAt: comment.CreatedAt, UpdatedAt: comment.UpdatedAt, SubComments: children,
	}
}
