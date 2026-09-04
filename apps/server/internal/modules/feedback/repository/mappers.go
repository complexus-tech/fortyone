package feedbackrepository

import (
	"fmt"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/domain"
	feedbacksql "github.com/complexus-tech/projects-api/internal/modules/feedback/repository/sqlc"
	"github.com/google/uuid"
)

type portalProjection struct {
	ID                  uuid.UUID
	WorkspaceID         uuid.UUID
	Name                string
	Slug                string
	IsPublic            bool
	ParticipationMode   string
	GuestIdentityPolicy string
	HasPublishedUpdates bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (row portalProjection) core() feedback.CorePortal {
	return feedback.CorePortal{
		ID:                  row.ID,
		WorkspaceID:         row.WorkspaceID,
		Name:                row.Name,
		Slug:                row.Slug,
		IsPublic:            row.IsPublic,
		ParticipationMode:   row.ParticipationMode,
		GuestIdentityPolicy: row.GuestIdentityPolicy,
		HasPublishedUpdates: row.HasPublishedUpdates,
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
	}
}

func toCoreBoard(row feedbacksql.FeedbackBoard) feedback.CoreBoard {
	return feedback.CoreBoard{
		ID: row.ID, WorkspaceID: row.WorkspaceID, PortalID: row.PortalID, TeamID: row.TeamID,
		Name: row.Name, Slug: row.Slug, Color: row.Color, OrderIndex: int(row.OrderIndex),
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

type itemProjection struct {
	ID, WorkspaceID, PortalID, BoardID, ContributorID   uuid.UUID
	AuthorID                                            *uuid.UUID
	AuthorName, AuthorEmail                             string
	AuthorAvatar                                        *string
	ParticipantKind                                     string
	AuthorMasked, Following                             bool
	MergedIntoItemID                                    *uuid.UUID
	MergedAt                                            *time.Time
	MergedByUserID                                      *uuid.UUID
	Title, Description, DescriptionHTML, Slug, Status   string
	VoteCount, UpvoteCount, DownvoteCount, CommentCount int32
	RoadmapSummary                                      *string
	BoardTeamID                                         uuid.UUID
	BoardName, BoardSlug, BoardColor                    string
	BoardOrderIndex                                     int32
	BoardCreatedAt, BoardUpdatedAt                      time.Time
	PrimaryLinkID, PrimaryStoryID                       *uuid.UUID
	PrimaryStoryTitle, PrimaryRelationship              *string
	PrimaryCreatedByUserID                              *uuid.UUID
	PrimaryCreatedAt                                    *time.Time
	ReadAt, DeletedAt                                   *time.Time
	CreatedAt, UpdatedAt                                time.Time
}

func (row itemProjection) core() feedback.CoreItem {
	authorID := uuid.Nil
	if row.AuthorID != nil {
		authorID = *row.AuthorID
	}
	item := feedback.CoreItem{
		ID: row.ID, WorkspaceID: row.WorkspaceID, PortalID: row.PortalID, BoardID: row.BoardID,
		ContributorID: row.ContributorID, AuthorID: authorID, AuthorName: row.AuthorName,
		AuthorEmail: row.AuthorEmail, AuthorAvatar: row.AuthorAvatar, ParticipantKind: row.ParticipantKind,
		AuthorMasked: row.AuthorMasked, MergedIntoItemID: row.MergedIntoItemID, MergedAt: row.MergedAt,
		MergedByUserID: row.MergedByUserID, Following: row.Following, Title: row.Title,
		Description: row.Description, DescriptionHTML: row.DescriptionHTML, Slug: row.Slug, Status: row.Status, VoteCount: int(row.VoteCount),
		UpvoteCount: int(row.UpvoteCount), DownvoteCount: int(row.DownvoteCount), CommentCount: int(row.CommentCount),
		RoadmapSummary: row.RoadmapSummary, ReadAt: row.ReadAt, DeletedAt: row.DeletedAt,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		Board: feedback.CoreBoard{ID: row.BoardID, WorkspaceID: row.WorkspaceID, PortalID: row.PortalID,
			TeamID: row.BoardTeamID, Name: row.BoardName, Slug: row.BoardSlug, Color: row.BoardColor,
			OrderIndex: int(row.BoardOrderIndex), CreatedAt: row.BoardCreatedAt, UpdatedAt: row.BoardUpdatedAt},
	}
	if row.PrimaryLinkID != nil && row.PrimaryStoryID != nil {
		link := feedback.CoreStoryLink{ID: *row.PrimaryLinkID, WorkspaceID: row.WorkspaceID, ItemID: row.ID,
			StoryID: *row.PrimaryStoryID, IsPrimary: true}
		if row.PrimaryStoryTitle != nil {
			link.StoryTitle = *row.PrimaryStoryTitle
		}
		if row.PrimaryRelationship != nil {
			link.Relationship = *row.PrimaryRelationship
		}
		if row.PrimaryCreatedByUserID != nil {
			link.CreatedByUserID = *row.PrimaryCreatedByUserID
		}
		if row.PrimaryCreatedAt != nil {
			link.CreatedAt = *row.PrimaryCreatedAt
		}
		item.StoryLinks = []feedback.CoreStoryLink{link}
	}
	return item
}

type commentProjection struct {
	ID, WorkspaceID, ItemID, ContributorID uuid.UUID
	AuthorID                               *uuid.UUID
	ParentID                               *uuid.UUID
	AuthorName                             string
	AuthorAvatar                           *string
	ParticipantKind                        string
	AuthorMasked                           bool
	Body                                   string
	CreatedAt, UpdatedAt                   time.Time
}

func (row commentProjection) core() feedback.CoreComment {
	authorID := uuid.Nil
	if row.AuthorID != nil {
		authorID = *row.AuthorID
	}
	return feedback.CoreComment{ID: row.ID, WorkspaceID: row.WorkspaceID, ItemID: row.ItemID,
		AuthorID: authorID, ContributorID: row.ContributorID, ParentID: row.ParentID,
		AuthorName: row.AuthorName, AuthorAvatar: row.AuthorAvatar, ParticipantKind: row.ParticipantKind,
		AuthorMasked: row.AuthorMasked, Body: row.Body, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
}

func stringValue(value any) (string, error) {
	switch typed := value.(type) {
	case string:
		return typed, nil
	case []byte:
		return string(typed), nil
	case nil:
		return "", nil
	default:
		return "", fmt.Errorf("unexpected feedback text projection %T", value)
	}
}

func uuidPointer(value uuid.UUID) *uuid.UUID {
	if value == uuid.Nil {
		return nil
	}
	return &value
}

func timePointer(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}

func nonEmptyStringPointer(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
