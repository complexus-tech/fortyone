package integrationrequests

import (
	"context"
	"time"

	integrationrequestdomain "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/domain"
	"github.com/google/uuid"
)

const (
	ProviderGitHub   = integrationrequestdomain.ProviderGitHub
	ProviderSlack    = integrationrequestdomain.ProviderSlack
	ProviderIntercom = integrationrequestdomain.ProviderIntercom
	SourceTypeIssue  = integrationrequestdomain.SourceTypeIssue
	StatusPending    = integrationrequestdomain.StatusPending
	StatusAccepted   = integrationrequestdomain.StatusAccepted
	StatusDeclined   = integrationrequestdomain.StatusDeclined

	AcceptanceStateIdle      = integrationrequestdomain.AcceptanceStateIdle
	AcceptanceStateReserved  = integrationrequestdomain.AcceptanceStateReserved
	CommentDirectionInbound  = integrationrequestdomain.CommentDirectionInbound
	CommentDirectionOutbound = integrationrequestdomain.CommentDirectionOutbound
)

type CoreIntegrationRequest = integrationrequestdomain.IntegrationRequest
type CoreUpsertRequestInput = integrationrequestdomain.UpsertRequestInput
type CoreListRequestsFilter = integrationrequestdomain.ListRequestsFilter

type CoreBulkRequestResult struct {
	// Count and RequestIDs are retained for existing callers and contain only
	// successful mutations.
	Count          int
	RequestIDs     []uuid.UUID
	TotalCount     int
	SucceededCount int
	FailedCount    int
	Partial        bool
	Items          []CoreBulkRequestItemResult
}

type CoreBulkRequestItemResult struct {
	RequestID       uuid.UUID
	Success         bool
	Status          string
	AcceptedStoryID *uuid.UUID
	Error           string
}

type OptionalValue[T any] = integrationrequestdomain.OptionalValue[T]
type CoreUpdateRequestInput = integrationrequestdomain.UpdateRequestInput
type CoreProviderThread = integrationrequestdomain.ProviderThread
type CoreIntegrationRequestComment = integrationrequestdomain.Comment
type CoreThreadActivity = integrationrequestdomain.ThreadActivity
type CoreBindProviderThreadInput = integrationrequestdomain.BindProviderThreadInput
type CoreProviderThreadMatchInput = integrationrequestdomain.ProviderThreadMatchInput
type CoreProviderThreadLookupInput = integrationrequestdomain.ProviderThreadLookupInput
type CoreCreateCommentInput = integrationrequestdomain.CreateCommentInput
type CorePreparedProviderComment = integrationrequestdomain.PreparedProviderComment
type CoreInboundProviderCommentInput = integrationrequestdomain.InboundProviderCommentInput

// NewStory is the integration-request module's story-creation intent. It is
// mapped to the Stories module only at composition time, so provider and
// persistence contracts do not leak a sibling service model.
type NewStory struct {
	Title                    string
	Description              *string
	StatusID                 *uuid.UUID
	ReporterID               *uuid.UUID
	AssigneeID               *uuid.UUID
	TeamID                   uuid.UUID
	Priority                 string
	EstimateValue            *int16
	EstimatedDurationMinutes *int
	MinimumFocusBlockMinutes *int
	ObjectiveID              *uuid.UUID
	KeyResultID              *uuid.UUID
	SprintID                 *uuid.UUID
	StartDate                *time.Time
	EndDate                  *time.Time
	LabelIDs                 []uuid.UUID
	CreationKey              string
}

// Story is the provider-safe snapshot produced by an accepted request. It is
// deliberately limited to fields used by acceptance callbacks.
type Story struct {
	ID          uuid.UUID
	SequenceID  int
	TeamID      uuid.UUID
	TeamCode    string
	Title       string
	Description *string
	StatusID    *uuid.UUID
	Priority    string
	AssigneeID  *uuid.UUID
	ReporterID  *uuid.UUID
	EndDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type StoryCreator interface {
	CreateForIntegrationRequest(ctx context.Context, actorID, workspaceID uuid.UUID, input NewStory) (Story, error)
}

type ProviderAccepter interface {
	AcceptIntegrationRequest(ctx context.Context, request CoreIntegrationRequest, story Story) error
}

type ProviderCommenter interface {
	PrepareIntegrationRequestComment(ctx context.Context, request CoreIntegrationRequest, thread CoreProviderThread, input CoreCreateCommentInput) (CorePreparedProviderComment, error)
	DeliverIntegrationRequestComment(ctx context.Context, request CoreIntegrationRequest, thread CoreProviderThread, comment CoreIntegrationRequestComment, prepared CorePreparedProviderComment) error
}
