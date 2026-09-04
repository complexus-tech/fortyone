package integrations

const (
	CapabilityWebhookVerification CapabilityKey = "control.webhook_verification"
	CapabilityTokenRefresh        CapabilityKey = "control.token_refresh"
	CapabilityExternalIdentity    CapabilityKey = "control.external_identity"

	CapabilityCodeHostRepositoryCatalog CapabilityKey = "codehost.repository_catalog"
	CapabilityCodeHostWorkItemWriter    CapabilityKey = "codehost.work_item_writer"
	CapabilityCodeHostCommentWriter     CapabilityKey = "codehost.comment_writer"

	CapabilityMessagingDelivery CapabilityKey = "messaging.delivery"
	CapabilityMessagingCommands CapabilityKey = "messaging.commands"
	CapabilityMessagingThreads  CapabilityKey = "messaging.threads"

	CapabilityCalendarEvents       CapabilityKey = "calendar.events"
	CapabilityCalendarAvailability CapabilityKey = "calendar.availability"
	CapabilityCalendarWatch        CapabilityKey = "calendar.watch"

	CapabilityDesignFileContext CapabilityKey = "design.file_context"

	CapabilityCloudFileContext CapabilityKey = "cloud_content.file_context"
	CapabilityCloudFileCreate  CapabilityKey = "cloud_content.file_create"
	CapabilityCloudFileImport  CapabilityKey = "cloud_content.file_import"
)
