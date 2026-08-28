package objectiveshttp

import (
	"context"
	"net/http"

	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
)

func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "objectiveshttp.handlers.Update")
	defer span.End()
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	objectiveID, ok := parseObjectivePathID(ctx, w, r, "id")
	if !ok {
		return nil
	}
	var request AppUpdateObjective
	if err := web.Decode(r, &request); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	comment := ""
	if request.Comment != nil {
		comment = *request.Comment
	}
	if err := h.objectives.UpdateIntentIfUnchanged(
		ctx, objectiveID, workspace.ID, userID, comment,
		request.ExpectedUpdatedAt, request.ObjectivePatch(),
	); err != nil {
		return respondObjectiveError(ctx, w, err)
	}
	h.invalidateObjectiveCache(ctx, workspace.ID, objectiveID)
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "objectiveshttp.handlers.Delete")
	defer span.End()
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	objectiveID, ok := parseObjectivePathID(ctx, w, r, "id")
	if !ok {
		return nil
	}
	if err := h.objectives.Delete(ctx, objectiveID, workspace.ID); err != nil {
		return respondObjectiveError(ctx, w, err)
	}
	h.invalidateObjectiveCache(ctx, workspace.ID, objectiveID)
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) CreateKeyResults(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	objectiveID, ok := parseObjectivePathID(ctx, w, r, "id")
	if !ok {
		return nil
	}
	if _, err := h.objectives.Get(ctx, objectiveID, workspace.ID); err != nil {
		return respondObjectiveError(ctx, w, err)
	}
	var request AppCreateKeyResultsRequest
	if err := web.Decode(r, &request); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	values := make([]keyresults.CoreNewKeyResult, len(request.KeyResults))
	for index, value := range request.KeyResults {
		values[index] = keyresults.CoreNewKeyResult{
			ObjectiveID: objectiveID, Name: value.Name, MeasurementType: value.MeasurementType,
			StartValue: value.StartValue, CurrentValue: value.CurrentValue, TargetValue: value.TargetValue,
			Lead: value.Lead, Contributors: value.Contributors, StartDate: value.StartDate.TimePtr(),
			EndDate: value.EndDate.TimePtr(), CreatedBy: userID,
		}
	}
	created, err := h.keyResults.CreateBatch(ctx, values, workspace.ID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	h.invalidateObjectiveCache(ctx, workspace.ID, objectiveID)
	return web.Respond(ctx, w, toAppKeyResults(created), http.StatusCreated)
}

func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "objectiveshttp.handlers.Create")
	defer span.End()
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	var request AppNewObjective
	if err := web.Decode(r, &request); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	keyResultDrafts := make([]objectives.CoreNewKeyResult, len(request.KeyResults))
	for index, value := range request.KeyResults {
		keyResultDrafts[index] = objectives.CoreNewKeyResult{
			Name: value.Name, MeasurementType: value.MeasurementType,
			StartValue: value.StartValue, CurrentValue: value.CurrentValue, TargetValue: value.TargetValue,
			Lead: value.Lead, Contributors: value.Contributors,
			StartDate: value.StartDate.TimePtr(), EndDate: value.EndDate.TimePtr(),
		}
	}
	objective, created, err := h.objectives.Create(
		ctx, toCoreNewObjective(request, userID), workspace.ID, keyResultDrafts,
	)
	if err != nil {
		return respondObjectiveError(ctx, w, err)
	}
	h.invalidateObjectiveCache(ctx, workspace.ID, objective.ID)
	return web.Respond(ctx, w, struct {
		Objective  AppObjectiveList `json:"objective"`
		KeyResults []AppKeyResult   `json:"keyResults,omitempty"`
	}{Objective: toAppObjective(objective), KeyResults: toAppObjectiveKeyResults(created)}, http.StatusCreated)
}
