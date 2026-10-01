package workautomationsadapter

import (
	"context"
	"errors"
	"html"
	"strings"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	automation "github.com/complexus-tech/projects-api/internal/modules/workautomations/domain"
	workautomations "github.com/complexus-tech/projects-api/internal/modules/workautomations/service"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

type StoryBackend interface {
	Get(context.Context, uuid.UUID, uuid.UUID) (stories.CoreSingleStory, error)
	UpdateTeamAutomationIfUnchanged(context.Context, uuid.UUID, uuid.UUID, time.Time, map[string]any, string) error
	CreateExternal(context.Context, uuid.UUID, stories.CoreNewStory, uuid.UUID) (stories.CoreSingleStory, error)
}

type Adapter struct{ backend StoryBackend }

func New(backend StoryBackend) workautomations.StoryActions { return &Adapter{backend: backend} }

func ownerContext(ctx context.Context, owner, workspace uuid.UUID) (context.Context, error) {
	// This composition port is used only by the worker. Binding a human identity
	// ensures the story transaction rechecks that owner's current live authority.
	return auth.BindWorkspace(auth.SetUserID(ctx, owner), workspace)
}
func (a *Adapter) Get(ctx context.Context, owner, workspace, id uuid.UUID) (automation.Story, error) {
	ctx, err := ownerContext(ctx, owner, workspace)
	if err != nil {
		return automation.Story{}, err
	}
	story, err := a.backend.Get(ctx, id, workspace)
	if err != nil {
		return automation.Story{}, err
	}
	if story.DeletedAt != nil || story.ArchivedAt != nil {
		return automation.Story{}, errors.New("task is unavailable")
	}
	return automation.Story{ID: story.ID, TeamID: story.Team, StatusID: story.Status, AssigneeID: story.Assignee, Priority: story.Priority, UpdatedAt: story.UpdatedAt}, nil
}
func (a *Adapter) Update(ctx context.Context, owner, workspace uuid.UUID, story automation.Story, actions automation.Actions, reason string) error {
	ctx, err := ownerContext(ctx, owner, workspace)
	if err != nil {
		return err
	}
	updates := make(map[string]any)
	if actions.StatusID != nil {
		updates["status_id"] = *actions.StatusID
	}
	if actions.Priority != nil {
		updates["priority"] = *actions.Priority
	}
	if actions.AssigneeID != nil {
		updates["assignee_id"] = *actions.AssigneeID
	} else if actions.ClearAssignee {
		updates["assignee_id"] = nil
	}
	return a.backend.UpdateTeamAutomationIfUnchanged(ctx, story.ID, workspace, story.UpdatedAt, updates, reason)
}
func (a *Adapter) Create(ctx context.Context, owner, workspace, team uuid.UUID, draft automation.Draft, key string) (uuid.UUID, error) {
	ctx, err := ownerContext(ctx, owner, workspace)
	if err != nil {
		return uuid.Nil, err
	}
	description, descriptionHTML := draft.Description, resetChecklist(draft.DescriptionHTML, draft.Checklist)
	fields := make([]storydomain.CustomFieldValue, len(draft.CustomFieldValues))
	for i, field := range draft.CustomFieldValues {
		fields[i] = storydomain.CustomFieldValue{FieldID: field.FieldID, Value: field.Value}
	}
	story, err := a.backend.CreateExternal(ctx, owner, stories.CoreNewStory{Title: draft.Title, Description: &description, DescriptionHTML: &descriptionHTML, Team: team, Status: draft.StatusID, Assignee: draft.AssigneeID, Reporter: &owner, Priority: draft.Priority, LabelIDs: append([]uuid.UUID(nil), draft.LabelIDs...), EstimateValue: draft.EstimateValue, EstimatedDurationMinutes: draft.EstimatedDurationMinutes, MinimumFocusBlockMinutes: draft.MinimumFocusBlockMinutes, CustomFieldValues: fields, CreationKey: &key, ExternalDelivery: storydomain.ExternalStoryDeliveryInternalOnly}, workspace)
	return story.ID, err
}
func resetChecklist(content string, checklist []string) string {
	content = strings.ReplaceAll(content, `data-checked="true"`, `data-checked="false"`)
	if len(checklist) == 0 {
		return content
	}
	var builder strings.Builder
	builder.WriteString(content)
	builder.WriteString(`<ul data-type="taskList">`)
	for _, item := range checklist {
		builder.WriteString(`<li data-type="taskItem" data-checked="false"><label><input type="checkbox"><span></span></label><div><p>`)
		builder.WriteString(html.EscapeString(item))
		builder.WriteString(`</p></div></li>`)
	}
	builder.WriteString(`</ul>`)
	return builder.String()
}
