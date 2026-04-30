export type IntegrationRequestStatus = "pending" | "accepted" | "declined";

export type IntegrationRequestProvider = "github" | "slack" | "intercom";

export type IntegrationRequest = {
  id: string;
  workspaceId: string;
  teamId: string;
  provider: IntegrationRequestProvider;
  sourceType: string;
  sourceExternalId: string;
  sourceNumber?: number;
  sourceUrl?: string;
  title: string;
  description?: string;
  status: IntegrationRequestStatus;
  metadata: Record<string, unknown>;
  acceptedStoryId?: string;
  createdAt: string;
  updatedAt: string;
};
