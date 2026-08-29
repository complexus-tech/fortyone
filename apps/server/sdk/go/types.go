package fortyone

// These aliases give the externally referenced OpenAPI schemas stable,
// idiomatic names while keeping the generated contract types authoritative.
type (
	Workspace                     = ComponentsResourcesWorkspace
	WorkspaceResponse             = ComponentsResourcesWorkspaceResponse
	WorkspaceRole                 = ComponentsResourcesWorkspaceRole
	Team                          = ComponentsResourcesTeam
	TeamPage                      = ComponentsResourcesTeamPageResponse
	Story                         = ComponentsResourcesStory
	StoryResponse                 = ComponentsResourcesStoryResponse
	StoryPage                     = ComponentsResourcesStoryPageResponse
	Comment                       = ComponentsResourcesComment
	CommentResponse               = ComponentsResourcesCommentResponse
	CommentPage                   = ComponentsResourcesCommentPageResponse
	CreateStoryRequest            = ComponentsResourcesCreateStoryRequest
	Label                         = ComponentsResourcesLabel
	LabelPage                     = ComponentsResourcesLabelPageResponse
	WorkflowState                 = ComponentsResourcesWorkflowState
	WorkflowStateCategory         = ComponentsResourcesWorkflowStateCategory
	WorkflowStatePage             = ComponentsResourcesWorkflowStatePageResponse
	Sprint                        = ComponentsResourcesSprint
	SprintPage                    = ComponentsResourcesSprintPageResponse
	Objective                     = ComponentsResourcesObjective
	ObjectiveHealth               = ComponentsResourcesObjectiveHealth
	ObjectiveScheduleStatus       = ComponentsResourcesObjectiveScheduleStatus
	ObjectivePage                 = ComponentsResourcesObjectivePageResponse
	KeyResult                     = ComponentsResourcesKeyResult
	KeyResultMeasurementType      = ComponentsResourcesKeyResultMeasurementType
	KeyResultPage                 = ComponentsResourcesKeyResultPageResponse
	StoryCounts                   = ComponentsResourcesStoryCounts
	PageMeta                      = ComponentsCommonPageMeta
	Error                         = ComponentsCommonError
	ErrorField                    = ComponentsCommonErrorField
	ErrorResponse                 = ComponentsCommonErrorResponse
	WebhookEndpoint               = ComponentsWebhooksWebhookEndpoint
	WebhookEndpointResponse       = ComponentsWebhooksWebhookEndpointResponse
	WebhookEndpointPage           = ComponentsWebhooksWebhookEndpointPageResponse
	WebhookEndpointStatus         = ComponentsWebhooksWebhookEndpointStatus
	WebhookEventType              = ComponentsWebhooksWebhookEventType
	CreateWebhookEndpointRequest  = ComponentsWebhooksCreateWebhookEndpointRequest
	CreatedWebhookEndpoint        = ComponentsWebhooksCreatedWebhookEndpointResponse
	ReplaceWebhookSubscriptions   = ComponentsWebhooksReplaceWebhookSubscriptionsRequest
	RotatedWebhookSecret          = ComponentsWebhooksRotateWebhookSecretResponse
	DisableWebhookEndpointRequest = ComponentsWebhooksDisableWebhookEndpointRequest
	WorkspaceID                   = ComponentsCommonWorkspaceId
	StoryID                       = ComponentsCommonStoryId
	CommentID                     = ComponentsCommonCommentId
	EndpointID                    = ComponentsCommonEndpointId
	Cursor                        = ComponentsCommonCursor
	PageLimit                     = ComponentsCommonPageLimit
)

const (
	WorkspaceRoleAdmin  = ComponentsResourcesWorkspaceRoleAdmin
	WorkspaceRoleGuest  = ComponentsResourcesWorkspaceRoleGuest
	WorkspaceRoleMember = ComponentsResourcesWorkspaceRoleMember

	WorkflowStateBacklog   = ComponentsResourcesWorkflowStateCategoryBacklog
	WorkflowStateUnstarted = ComponentsResourcesWorkflowStateCategoryUnstarted
	WorkflowStateStarted   = ComponentsResourcesWorkflowStateCategoryStarted
	WorkflowStatePaused    = ComponentsResourcesWorkflowStateCategoryPaused
	WorkflowStateCompleted = ComponentsResourcesWorkflowStateCategoryCompleted
	WorkflowStateCancelled = ComponentsResourcesWorkflowStateCategoryCancelled

	KeyResultPercentage = ComponentsResourcesKeyResultMeasurementTypePercentage
	KeyResultNumber     = ComponentsResourcesKeyResultMeasurementTypeNumber
	KeyResultBoolean    = ComponentsResourcesKeyResultMeasurementTypeBoolean

	WebhookEndpointActive   = ComponentsWebhooksWebhookEndpointStatusActive
	WebhookEndpointDisabled = ComponentsWebhooksWebhookEndpointStatusDisabled

	WebhookEventCommentCreated = ComponentsWebhooksWebhookEventTypeCommentCreated
	WebhookEventCommentDeleted = ComponentsWebhooksWebhookEventTypeCommentDeleted
	WebhookEventCommentUpdated = ComponentsWebhooksWebhookEventTypeCommentUpdated
	WebhookEventStoryCreated   = ComponentsWebhooksWebhookEventTypeStoryCreated
	WebhookEventStoryDeleted   = ComponentsWebhooksWebhookEventTypeStoryDeleted
	WebhookEventStoryUpdated   = ComponentsWebhooksWebhookEventTypeStoryUpdated
)
