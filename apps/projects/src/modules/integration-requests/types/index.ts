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
  statusId?: string;
  priority: "No Priority" | "Low" | "Medium" | "High" | "Urgent";
  assigneeId?: string;
  estimateValue?: number;
  objectiveId?: string;
  keyResultId?: string;
  sprintId?: string;
  startDate?: string;
  endDate?: string;
  labelIds: string[];
  status: IntegrationRequestStatus;
  metadata: Record<string, unknown>;
  acceptedStoryId?: string;
  createdAt: string;
  updatedAt: string;
};

export type IntegrationRequestsPage = {
  requests: IntegrationRequest[];
  pagination: {
    page: number;
    pageSize: number;
    totalCount: number;
    hasMore: boolean;
    nextPage: number;
  };
};

export type UpdateIntegrationRequestInput = {
  title?: string;
  description?: string | null;
  statusId?: string | null;
  priority?: IntegrationRequest["priority"];
  assigneeId?: string | null;
  estimateValue?: number | null;
  objectiveId?: string | null;
  keyResultId?: string | null;
  sprintId?: string | null;
  startDate?: string | null;
  endDate?: string | null;
  labelIds?: string[];
};

export type IntegrationRequestProviderThread = {
  id: string;
  integrationRequestId: string;
  teamId: string;
  acceptedStoryId?: string;
  provider: "slack";
  externalChannelId: string;
  externalThreadId: string;
  externalSourceMessageId?: string;
  sourceUrl?: string;
  requestTitle: string;
  createdAt: string;
  updatedAt: string;
};

export type IntegrationRequestComment = {
  id: string;
  threadId: string;
  direction: "inbound" | "outbound";
  authorUserId?: string;
  authorName: string;
  authorAvatar?: string;
  externalAuthorId?: string;
  externalMessageId?: string;
  deliveryStatus?: "sent" | "sending" | "retrying" | "failed" | "not-sent";
  body: string;
  createdAt: string;
  updatedAt: string;
};

export type IntegrationRequestThreadActivity = {
  thread: IntegrationRequestProviderThread;
  comments: IntegrationRequestComment[];
};

export type BulkIntegrationRequestResult = {
  count: number;
  requestIds: string[];
};
