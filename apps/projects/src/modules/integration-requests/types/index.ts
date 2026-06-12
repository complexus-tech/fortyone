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
  status: IntegrationRequestStatus;
  metadata: Record<string, unknown>;
  acceptedStoryId?: string;
  createdAt: string;
  updatedAt: string;
};

export type UpdateIntegrationRequestInput = Partial<
  Pick<
    IntegrationRequest,
    | "title"
    | "description"
    | "statusId"
    | "priority"
    | "assigneeId"
    | "estimateValue"
    | "objectiveId"
    | "keyResultId"
    | "sprintId"
    | "startDate"
    | "endDate"
  >
>;

export type BulkIntegrationRequestResult = {
  count: number;
  requestIds: string[];
};
