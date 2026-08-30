import { z } from "zod";
import { isEstimateValue } from "@/lib/estimate";
import { MAX_TIME_NEEDED_MINUTES } from "@/lib/time-needed";
import type { AuthenticatedIntegrationRequestContext } from "./context";

export const INTEGRATION_REQUEST_PRIORITY_SCHEMA = z.enum([
  "No Priority",
  "Low",
  "Medium",
  "High",
  "Urgent",
]);

export const INTEGRATION_REQUEST_STATUS_SCHEMA = z.enum([
  "pending",
  "accepted",
  "declined",
]);

export const INTEGRATION_REQUEST_PROVIDER_SCHEMA = z.enum([
  "github",
  "slack",
  "intercom",
]);

export const MAX_INTEGRATION_REQUEST_TEAMS = 20;

const INTEGRATION_REQUEST_UPDATE_FIELDS = [
  "title",
  "description",
  "statusId",
  "priority",
  "assigneeId",
  "estimateValue",
  "estimatedDurationMinutes",
  "minimumFocusBlockMinutes",
  "objectiveId",
  "keyResultId",
  "sprintId",
  "startDate",
  "endDate",
] as const;

export const listIntegrationRequestsInputSchema = z.object({
  teamIds: z
    .array(z.string().uuid("Each team ID must be a valid UUID."))
    .min(1)
    .max(
      MAX_INTEGRATION_REQUEST_TEAMS,
      `List requests for at most ${MAX_INTEGRATION_REQUEST_TEAMS} teams at once.`,
    )
    .refine(
      (teamIds) => new Set(teamIds).size === teamIds.length,
      "Team IDs must be unique.",
    )
    .describe(
      `Team IDs to list integration requests for (max ${MAX_INTEGRATION_REQUEST_TEAMS}).`,
    ),
  status: INTEGRATION_REQUEST_STATUS_SCHEMA.optional().describe(
    "Request status filter. Defaults to pending.",
  ),
  provider: INTEGRATION_REQUEST_PROVIDER_SCHEMA.optional().describe(
    "Provider filter, for example github or slack.",
  ),
  priority:
    INTEGRATION_REQUEST_PRIORITY_SCHEMA.optional().describe("Priority filter."),
  assigneeId: z.string().optional().describe("Assignee user ID filter."),
  createdAfter: z
    .string()
    .optional()
    .describe("Only include requests created on or after this ISO date."),
  createdBefore: z
    .string()
    .optional()
    .describe("Only include requests created on or before this ISO date."),
  page: z.number().min(1).optional().describe("Page number. Default 1."),
  pageSize: z
    .number()
    .min(1)
    .max(100)
    .optional()
    .describe("Requests per team per page. Default 20, max 100."),
});

export const getIntegrationRequestInputSchema = z.object({
  requestId: z.string().describe("Integration request ID."),
  includeGitHubComments: z
    .boolean()
    .optional()
    .describe("Include GitHub comments when the request came from GitHub."),
});

export const updateIntegrationRequestInputSchema = z
  .object({
    requestId: z.string().describe("Integration request ID."),
    confirmed: z
      .boolean()
      .optional()
      .describe("Must be true after the user explicitly confirms the update."),
    title: z.string().optional(),
    description: z.string().optional(),
    statusId: z.string().optional(),
    priority: INTEGRATION_REQUEST_PRIORITY_SCHEMA.optional(),
    assigneeId: z.string().optional(),
    estimateValue: z
      .number()
      .int()
      .refine(isEstimateValue, {
        message: "Complexity must be 1, 2, 3, 5, or 8.",
      })
      .nullable()
      .optional()
      .describe(
        "Relative complexity value using the team's scale. This is not a time duration; set null to clear it.",
      ),
    estimatedDurationMinutes: z
      .number()
      .int()
      .positive()
      .max(MAX_TIME_NEEDED_MINUTES)
      .nullable()
      .optional()
      .describe(
        "Total time needed in minutes for calendar scheduling. Set null to clear both the duration and its minimum focus block.",
      ),
    minimumFocusBlockMinutes: z
      .number()
      .int()
      .positive()
      .max(MAX_TIME_NEEDED_MINUTES)
      .nullable()
      .optional()
      .describe(
        "Optional smallest schedulable focus block in minutes. It cannot exceed estimatedDurationMinutes; set null to let Maya automatically fill available calendar time.",
      ),
    objectiveId: z.string().optional(),
    keyResultId: z.string().optional(),
    sprintId: z.string().optional(),
    startDate: z.string().optional(),
    endDate: z.string().optional(),
  })
  .refine(
    (input) =>
      INTEGRATION_REQUEST_UPDATE_FIELDS.some(
        (field) => input[field] !== undefined,
      ),
    "Provide at least one integration request field to update.",
  );

export type IntegrationRequestStatus = z.infer<
  typeof INTEGRATION_REQUEST_STATUS_SCHEMA
>;

export type IntegrationRequestProvider = z.infer<
  typeof INTEGRATION_REQUEST_PROVIDER_SCHEMA
>;

export type IntegrationRequestPriority = z.infer<
  typeof INTEGRATION_REQUEST_PRIORITY_SCHEMA
>;

export type UpdateIntegrationRequestToolInput = Omit<
  z.infer<typeof updateIntegrationRequestInputSchema>,
  "confirmed" | "requestId"
>;

export type IntegrationRequestToolRecord = {
  acceptedStoryId?: string;
  assigneeId?: string;
  createdAt: string;
  endDate?: string;
  estimateValue?: number;
  estimatedDurationMinutes?: number;
  id: string;
  keyResultId?: string;
  minimumFocusBlockMinutes?: number;
  objectiveId?: string;
  priority: IntegrationRequestPriority;
  provider: IntegrationRequestProvider;
  sourceNumber?: number;
  sourceType: string;
  sourceUrl?: string;
  sprintId?: string;
  startDate?: string;
  status: IntegrationRequestStatus;
  teamId: string;
  title: string;
  updatedAt: string;
};

export type IntegrationRequestToolPage = {
  pagination: {
    hasMore: boolean;
    nextPage: number;
    page: number;
    pageSize: number;
    totalCount: number;
  };
  requests: IntegrationRequestToolRecord[];
};

export type IntegrationRequestToolListFilters = {
  assigneeId?: string;
  createdAfter?: string;
  createdBefore?: string;
  priority?: IntegrationRequestPriority;
  provider?: IntegrationRequestProvider;
};

export type IntegrationRequestToolBulkResult = {
  count: number;
  failedCount: number;
  items: {
    acceptedStoryId?: string;
    error?: string;
    requestId: string;
    status: "accepted" | "declined" | "failed";
    success: boolean;
  }[];
  partial: boolean;
  requestIds: string[];
  succeededCount: number;
  totalCount: number;
};

export type IntegrationRequestToolActionResult<T> = {
  data?: T | null;
  error?: {
    message?: string;
  };
};

export type IntegrationRequestToolDependencies = {
  acceptAllIntegrationRequests: (
    teamId: string,
    workspaceSlug: string,
  ) => Promise<
    IntegrationRequestToolActionResult<IntegrationRequestToolBulkResult>
  >;
  acceptIntegrationRequest: (
    requestId: string,
    workspaceSlug: string,
  ) => Promise<
    IntegrationRequestToolActionResult<IntegrationRequestToolRecord>
  >;
  declineAllIntegrationRequests: (
    teamId: string,
    workspaceSlug: string,
  ) => Promise<
    IntegrationRequestToolActionResult<IntegrationRequestToolBulkResult>
  >;
  declineIntegrationRequest: (
    requestId: string,
    workspaceSlug: string,
  ) => Promise<
    IntegrationRequestToolActionResult<IntegrationRequestToolRecord>
  >;
  getIntegrationRequest: (
    requestId: string,
    ctx: AuthenticatedIntegrationRequestContext,
  ) => Promise<IntegrationRequestToolRecord>;
  getRequestGitHubComments: (
    requestId: string,
    ctx: AuthenticatedIntegrationRequestContext,
  ) => Promise<unknown[]>;
  getTeamIntegrationRequestsPage: (
    teamId: string,
    ctx: AuthenticatedIntegrationRequestContext,
    status: IntegrationRequestStatus,
    page: number,
    pageSize: number,
    filters: IntegrationRequestToolListFilters,
  ) => Promise<IntegrationRequestToolPage>;
  postRequestGitHubComment: (
    requestId: string,
    payload: { body: string },
    workspaceSlug: string,
  ) => Promise<IntegrationRequestToolActionResult<null>>;
  updateIntegrationRequest: (
    requestId: string,
    payload: UpdateIntegrationRequestToolInput,
    workspaceSlug: string,
  ) => Promise<
    IntegrationRequestToolActionResult<IntegrationRequestToolRecord>
  >;
};

export const toIntegrationRequestSummary = (
  request: IntegrationRequestToolRecord,
) => ({
  id: request.id,
  teamId: request.teamId,
  provider: request.provider,
  sourceType: request.sourceType,
  sourceNumber: request.sourceNumber,
  sourceUrl: request.sourceUrl,
  title: request.title,
  status: request.status,
  priority: request.priority,
  assigneeId: request.assigneeId,
  estimateValue: request.estimateValue,
  estimatedDurationMinutes: request.estimatedDurationMinutes,
  minimumFocusBlockMinutes: request.minimumFocusBlockMinutes,
  objectiveId: request.objectiveId,
  keyResultId: request.keyResultId,
  sprintId: request.sprintId,
  startDate: request.startDate,
  endDate: request.endDate,
  acceptedStoryId: request.acceptedStoryId,
  createdAt: request.createdAt,
  updatedAt: request.updatedAt,
});
