package apiv1http

import (
	"context"

	keyresultsdomain "github.com/complexus-tech/projects-api/internal/modules/keyresults/domain"
	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	labels "github.com/complexus-tech/projects-api/internal/modules/labels/service"
	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	sprintdomain "github.com/complexus-tech/projects-api/internal/modules/sprints/domain"
	sprints "github.com/complexus-tech/projects-api/internal/modules/sprints/service"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

type labelReaderStub struct{}

func (labelReaderStub) GetLabels(context.Context, uuid.UUID, uuid.UUID, labels.LabelFilters) ([]labels.CoreLabel, error) {
	return []labels.CoreLabel{}, nil
}

type workflowStateReaderStub struct{}

func (workflowStateReaderStub) TeamListForMember(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) ([]states.CoreState, error) {
	return []states.CoreState{}, nil
}

type sprintReaderStub struct{}

func (sprintReaderStub) ListQuery(context.Context, sprintdomain.ListQuery) ([]sprints.CoreSprint, error) {
	return []sprints.CoreSprint{}, nil
}

type objectiveReaderStub struct{}

func (objectiveReaderStub) ListIntent(context.Context, objectivesdomain.ListQuery) ([]objectives.CoreObjective, error) {
	return []objectives.CoreObjective{}, nil
}

type keyResultReaderStub struct{}

func (keyResultReaderStub) ListPaginated(context.Context, keyresultsdomain.Filters) (keyresults.CoreKeyResultListResponse, error) {
	return keyresults.CoreKeyResultListResponse{KeyResults: []keyresults.CoreKeyResultWithObjective{}}, nil
}

type storyCommentReaderStub struct{}

func (storyCommentReaderStub) GetComments(context.Context, uuid.UUID, uuid.UUID, int, int) ([]stories.CoreComment, bool, error) {
	return []stories.CoreComment{}, false, nil
}

func (storyCommentReaderStub) GetComment(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (stories.CoreComment, error) {
	return stories.CoreComment{}, stories.ErrNotFound
}
