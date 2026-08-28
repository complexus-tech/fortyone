package adminhttp

import (
	"time"

	admindomain "github.com/complexus-tech/projects-api/internal/modules/admin/domain"
	platformpatch "github.com/complexus-tech/projects-api/internal/platform/patch"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type updateWorkspaceTrialRequest struct {
	TrialEndsOn time.Time `json:"trialEndsOn"`
	Reason      string    `json:"reason"`
}

type updateWorkspaceDeletedRequest struct {
	Deleted bool   `json:"deleted"`
	Reason  string `json:"reason"`
}

type reasonRequest struct {
	Reason string `json:"reason"`
}

type updateUserStateRequest struct {
	IsActive   web.PatchField[bool] `json:"isActive"`
	IsInternal web.PatchField[bool] `json:"isInternal"`
	Reason     string               `json:"reason"`
}

type createAdminNoteRequest struct {
	TargetType  string     `json:"targetType"`
	TargetID    uuid.UUID  `json:"targetId"`
	WorkspaceID *uuid.UUID `json:"workspaceId"`
	Body        string     `json:"body"`
}

func userStatePatch(request updateUserStateRequest) admindomain.UserStatePatch {
	return admindomain.UserStatePatch{
		IsActive:   booleanPatchField(request.IsActive),
		IsInternal: booleanPatchField(request.IsInternal),
	}
}

func booleanPatchField(field web.PatchField[bool]) platformpatch.Field[bool] {
	value, specified := field.Value()
	if !specified {
		return platformpatch.Field[bool]{}
	}
	if value == nil {
		return platformpatch.Clear[bool]()
	}
	return platformpatch.Set(*value)
}
