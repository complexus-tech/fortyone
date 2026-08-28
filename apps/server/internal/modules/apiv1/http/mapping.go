package apiv1http

import (
	"fmt"
	"time"

	openapiv1 "github.com/complexus-tech/projects-api/internal/generated/openapi/v1"
	outboundwebhooksdomain "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/domain"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	"github.com/google/uuid"
)

func workspaceModel(workspace workspaces.CoreWorkspace) openapiv1.ComponentsResourcesWorkspace {
	return openapiv1.ComponentsResourcesWorkspace{
		Id: workspace.ID, Slug: workspace.Slug, Name: workspace.Name, Color: workspace.Color,
		AvatarUrl: workspace.AvatarURL, Active: workspace.IsActive,
		Role:      openapiv1.ComponentsResourcesWorkspaceRole(workspace.UserRole),
		CreatedAt: workspace.CreatedAt.UTC(), UpdatedAt: workspace.UpdatedAt.UTC(),
	}
}

func teamModel(team teams.CoreTeam) openapiv1.ComponentsResourcesTeam {
	return openapiv1.ComponentsResourcesTeam{
		Id: team.ID, WorkspaceId: team.Workspace, Name: team.Name, Code: team.Code,
		Color: team.Color, Private: team.IsPrivate, MemberCount: team.MemberCount,
		SprintsEnabled: team.SprintsEnabled, CreatedAt: team.CreatedAt.UTC(), UpdatedAt: team.UpdatedAt.UTC(),
	}
}

func storyListModel(story stories.CoreStoryList) openapiv1.ComponentsResourcesStory {
	teamCode := "STORY"
	if story.TeamSummary != nil && story.TeamSummary.Code != "" {
		teamCode = story.TeamSummary.Code
	}
	return openapiv1.ComponentsResourcesStory{
		Id: story.ID, WorkspaceId: story.Workspace, TeamId: story.Team, SequenceId: story.SequenceID,
		Reference: fmt.Sprintf("%s-%d", teamCode, story.SequenceID), Title: story.Title,
		EstimateLabel: story.EstimateLabel, EstimateValue: estimateValue(story.EstimateValue),
		EstimatedDurationMinutes: story.EstimatedDurationMinutes, MinimumFocusBlockMinutes: story.MinimumFocusBlockMinutes,
		AutoSchedulingEnabled: story.AutoSchedulingEnabled, AutoSchedulingLocked: story.AutoSchedulingLocked,
		AutoSchedulingStatus: story.AutoSchedulingStatus, ParentId: story.Parent, ObjectiveId: story.Objective,
		StatusId: story.Status, AssigneeId: story.Assignee, ReporterId: story.Reporter, Priority: story.Priority,
		SprintId: story.Sprint, KeyResultId: story.KeyResult, StartDate: utcTimePointer(story.StartDate),
		EndDate: utcTimePointer(story.EndDate), CreatedAt: story.CreatedAt.UTC(), UpdatedAt: story.UpdatedAt.UTC(),
		CompletedAt: utcTimePointer(story.CompletedAt), ArchivedAt: utcTimePointer(story.ArchivedAt),
		Labels: copyUUIDs(story.Labels),
	}
}

func storyDetailModel(story stories.CoreSingleStory) openapiv1.ComponentsResourcesStory {
	teamCode := story.TeamCode
	if teamCode == "" {
		teamCode = "STORY"
	}
	return openapiv1.ComponentsResourcesStory{
		Id: story.ID, WorkspaceId: story.Workspace, TeamId: story.Team, SequenceId: story.SequenceID,
		Reference: fmt.Sprintf("%s-%d", teamCode, story.SequenceID), Title: story.Title, Description: story.Description,
		EstimateLabel: story.EstimateLabel, EstimateValue: estimateValue(story.EstimateValue),
		EstimatedDurationMinutes: story.EstimatedDurationMinutes, MinimumFocusBlockMinutes: story.MinimumFocusBlockMinutes,
		AutoSchedulingEnabled: story.AutoSchedulingEnabled, AutoSchedulingLocked: story.AutoSchedulingLocked,
		AutoSchedulingStatus: story.AutoSchedulingStatus, ParentId: story.Parent, ObjectiveId: story.Objective,
		StatusId: story.Status, AssigneeId: story.Assignee, ReporterId: story.Reporter, Priority: story.Priority,
		SprintId: story.Sprint, KeyResultId: story.KeyResult, StartDate: utcTimePointer(story.StartDate),
		EndDate: utcTimePointer(story.EndDate), CreatedAt: story.CreatedAt.UTC(), UpdatedAt: story.UpdatedAt.UTC(),
		CompletedAt: utcTimePointer(story.CompletedAt), ArchivedAt: utcTimePointer(story.ArchivedAt),
		Labels: copyUUIDs(story.Labels),
	}
}

func estimateValue(value *int16) *int32 {
	if value == nil {
		return nil
	}
	converted := int32(*value)
	return &converted
}

func copyUUIDs(values []uuid.UUID) []uuid.UUID {
	result := make([]uuid.UUID, len(values))
	copy(result, values)
	return result
}

func utcTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	utc := value.UTC()
	return &utc
}

func webhookModel(endpoint outboundwebhooksdomain.Endpoint) openapiv1.ComponentsWebhooksWebhookEndpoint {
	subscriptions := make([]openapiv1.ComponentsWebhooksWebhookEventType, len(endpoint.Subscriptions))
	for index, eventType := range endpoint.Subscriptions {
		subscriptions[index] = openapiv1.ComponentsWebhooksWebhookEventType(eventType)
	}
	return openapiv1.ComponentsWebhooksWebhookEndpoint{
		Id: endpoint.ID, WorkspaceId: endpoint.WorkspaceID, Name: endpoint.Name, Url: endpoint.URL,
		Status:           openapiv1.ComponentsWebhooksWebhookEndpointStatus(endpoint.Status),
		SecretGeneration: endpoint.SecretGeneration, SubscriptionGeneration: endpoint.SubscriptionGeneration,
		Subscriptions: subscriptions, ConsecutiveFailures: endpoint.ConsecutiveFailures,
		LastSuccessAt: utcTimePointer(endpoint.LastSuccessAt), DisabledAt: utcTimePointer(endpoint.DisabledAt),
		DisabledReason: endpoint.DisabledReason, CreatedAt: endpoint.CreatedAt.UTC(), UpdatedAt: endpoint.UpdatedAt.UTC(),
	}
}
