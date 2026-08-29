import type {
  components,
  Readable,
  Writable,
} from "./generated/schema.js";

// These aliases keep OpenAPI-generated schemas authoritative while presenting
// concise names to integrations. Readable/Writable apply the contract's
// readOnly and writeOnly rules without copying schema definitions.
export type Workspace = components["schemas"]["Workspace"];
export type WorkspaceRole = components["schemas"]["Workspace"]["role"];
export type WorkspaceResponse = components["schemas"]["WorkspaceResponse"];
export type Team = components["schemas"]["Team"];
export type TeamPage = components["schemas"]["TeamPageResponse"];
export type Story = components["schemas"]["Story"];
export type StoryPage = components["schemas"]["StoryPageResponse"];
export type StoryResponse = components["schemas"]["StoryResponse"];
export type CreateStoryRequest = Writable<
  components["schemas"]["CreateStoryRequest"]
>;
export type Comment = components["schemas"]["Comment"];
export type CommentResponse = components["schemas"]["CommentResponse"];
export type CommentPage = components["schemas"]["CommentPageResponse"];
export type Label = components["schemas"]["Label"];
export type LabelPage = components["schemas"]["LabelPageResponse"];
export type WorkflowState = components["schemas"]["WorkflowState"];
export type WorkflowStatePage =
  components["schemas"]["WorkflowStatePageResponse"];
export type Sprint = components["schemas"]["Sprint"];
export type SprintPage = components["schemas"]["SprintPageResponse"];
export type Objective = components["schemas"]["Objective"];
export type ObjectivePage = components["schemas"]["ObjectivePageResponse"];
export type KeyResult = components["schemas"]["KeyResult"];
export type KeyResultPage = components["schemas"]["KeyResultPageResponse"];
export type StoryCounts = Sprint["storyCounts"];
export type PageMeta = components["schemas"]["PageMeta"];
export type WebhookEventType = components["schemas"]["WebhookEventType"];
export type WebhookEndpoint = components["schemas"]["WebhookEndpoint"];
export type WebhookEndpointPage =
  components["schemas"]["WebhookEndpointPageResponse"];
export type WebhookEndpointResponse =
  components["schemas"]["WebhookEndpointResponse"];
export type CreatedWebhookEndpoint = Readable<
  components["schemas"]["CreatedWebhookEndpointResponse"]
>;
export type RotateWebhookSecretResponse = Readable<
  components["schemas"]["RotateWebhookSecretResponse"]
>;
export type CreateWebhookEndpointRequest = Writable<
  components["schemas"]["CreateWebhookEndpointRequest"]
>;
export type ReplaceWebhookSubscriptionsRequest = Writable<
  components["schemas"]["ReplaceWebhookSubscriptionsRequest"]
>;
export type DisableWebhookEndpointRequest = Writable<
  components["schemas"]["DisableWebhookEndpointRequest"]
>;
