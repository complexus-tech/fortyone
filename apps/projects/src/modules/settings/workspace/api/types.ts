export type DeveloperScope =
  | "workspaces:read"
  | "teams:read"
  | "stories:read"
  | "stories:write"
  | "comments:read"
  | "comments:write"
  | "labels:read"
  | "labels:write"
  | "sprints:read"
  | "objectives:read"
  | "objectives:write"
  | "webhooks:manage";

export type Credential = {
  id: string;
  principalId: string;
  kind: "personal_access_token" | "service_account_key";
  name: string;
  prefix: string;
  scopes: DeveloperScope[];
  teamIds: string[];
  expiresAt: string;
  lastUsedAt?: string;
  rotatedFromId?: string;
  rotatedAt?: string;
  revokedAt?: string;
  createdAt: string;
};

export type IssuedCredential = {
  credential: Credential;
  token: string;
};

export type CreateCredentialInput = {
  name: string;
  scopes: DeveloperScope[];
  teamIds?: string[];
  expiresAt: string;
};

export type ServiceAccount = {
  id: string;
  name: string;
  workspaceRole: "guest" | "member";
  status: "active" | "disabled";
  createdAt: string;
  updatedAt: string;
  disabledAt?: string;
  disabledReason?: string;
};

export type OAuthApplication = {
  id: string;
  clientId: string;
  name: string;
  registrationKind: string;
  redirectUris: string[];
  expiresAt: string;
  ownerWorkspaceId: string;
  ownerUserId: string;
  status: string;
  createdAt: string;
  updatedAt: string;
  revokedAt?: string;
};

export type OAuthClientSecret = {
  id: string;
  applicationId: string;
  prefix: string;
  expiresAt: string;
  lastUsedAt?: string;
  overlapExpiresAt?: string;
  revokedAt?: string;
  createdAt: string;
};

export type IssuedOAuthApplication = {
  application: OAuthApplication;
  clientSecret: OAuthClientSecret;
  secret: string;
};

export type IssuedOAuthSecret = {
  clientSecret: OAuthClientSecret;
  secret: string;
  previousSecretOverlapExpiresAt?: string;
};

export type WebhookEventType =
  | "story.created"
  | "story.updated"
  | "story.deleted"
  | "comment.created"
  | "comment.updated"
  | "comment.deleted";

export type WebhookEndpoint = {
  id: string;
  workspaceId: string;
  name: string;
  url: string;
  status: "active" | "disabled";
  secretGeneration: number;
  subscriptionGeneration: number;
  subscriptions: WebhookEventType[];
  consecutiveFailures: number;
  lastSuccessAt?: string;
  disabledAt?: string;
  disabledReason?: string;
  createdAt: string;
  updatedAt: string;
};

export type WebhookEndpointPage = {
  items: WebhookEndpoint[];
  nextCursor?: string;
};

export type CreatedWebhookEndpoint = {
  endpoint: WebhookEndpoint;
  signingSecret: string;
};

export type RotatedWebhookSecret = {
  signingSecret: string;
  generation: number;
  previousSecretExpiresAt: string;
};
