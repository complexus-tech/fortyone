import { tool } from "ai";
import { z } from "zod";
import { auth } from "@/auth";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import { acceptIntegrationRequestAction } from "@/modules/integration-requests/actions/accept";
import { acceptAllIntegrationRequestsAction } from "@/modules/integration-requests/actions/accept-all";
import { declineIntegrationRequestAction } from "@/modules/integration-requests/actions/decline";
import { declineAllIntegrationRequestsAction } from "@/modules/integration-requests/actions/decline-all";
import { postRequestGitHubCommentAction } from "@/modules/integration-requests/actions/post-github-comment";
import { updateIntegrationRequestAction } from "@/modules/integration-requests/actions/update";
import { getIntegrationRequest } from "@/modules/integration-requests/queries/get-request";
import { getRequestGitHubComments } from "@/modules/integration-requests/queries/get-request-github-comments";
import { getTeamIntegrationRequestsPage } from "@/modules/integration-requests/queries/get-team-requests";
import type { IntegrationRequestStatus } from "@/modules/integration-requests/types";
import { normalizeOptionalString } from "./normalize-input";
import {
  requireToolConfirmation,
  resolvePaginationInput,
} from "./tool-helpers";

const prioritySchema = z.enum([
  "No Priority",
  "Low",
  "Medium",
  "High",
  "Urgent",
]);

const requestStatusSchema = z.enum(["pending", "accepted", "declined"]);
const requestProviderSchema = z.enum(["github", "slack", "intercom"]);

const getAuthenticatedContext = async (experimentalContext: unknown) => {
  const session = await auth();

  if (!session) {
    return { error: "Authentication required to access integration requests" };
  }

  const workspaceSlug = (experimentalContext as { workspaceSlug?: string })
    .workspaceSlug;

  if (!workspaceSlug) {
    return { error: "Workspace context is required" };
  }

  return { session, workspaceSlug };
};

const toRequestSummary = (
  request: Awaited<ReturnType<typeof getIntegrationRequest>>,
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
  objectiveId: request.objectiveId,
  keyResultId: request.keyResultId,
  sprintId: request.sprintId,
  startDate: request.startDate,
  endDate: request.endDate,
  acceptedStoryId: request.acceptedStoryId,
  createdAt: request.createdAt,
  updatedAt: request.updatedAt,
});

export const listIntegrationRequestsTool = tool({
  description:
    "List integration requests for one or more teams. Use this for GitHub, Slack, or Intercom request triage before accepting or declining requests.",
  inputSchema: z.object({
    teamIds: z
      .array(z.string())
      .min(1)
      .describe("Team IDs to list integration requests for."),
    status: requestStatusSchema
      .optional()
      .describe("Request status filter. Defaults to pending."),
    provider: requestProviderSchema
      .optional()
      .describe("Provider filter, for example github or slack."),
    priority: prioritySchema.optional().describe("Priority filter."),
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
  }),
  execute: async (
    {
      teamIds,
      status = "pending",
      provider,
      priority,
      assigneeId,
      createdAfter,
      createdBefore,
      page,
      pageSize,
    },
    { experimental_context: experimentalContext },
  ) => {
    try {
      const ctx = await getAuthenticatedContext(experimentalContext);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const pagination = resolvePaginationInput({ page, pageSize });
      const pages = await Promise.all(
        teamIds.map(async (teamId) => {
          const response = await getTeamIntegrationRequestsPage(
            teamId,
            ctx,
            status as IntegrationRequestStatus,
            pagination.page,
            pagination.pageSize,
            {
              provider,
              priority,
              assigneeId: normalizeOptionalString(assigneeId),
              createdAfter: normalizeOptionalString(createdAfter),
              createdBefore: normalizeOptionalString(createdBefore),
            },
          );

          return {
            teamId,
            requests: response.requests.map(toRequestSummary),
            pagination: response.pagination,
          };
        }),
      );

      const totalReturned = pages.reduce(
        (total, pageResult) => total + pageResult.requests.length,
        0,
      );

      return {
        success: true,
        teams: pages,
        count: totalReturned,
        filters: {
          status,
          provider,
          priority,
          assigneeId,
          createdAfter,
          createdBefore,
        },
        message: `Found ${totalReturned} integration request${totalReturned === 1 ? "" : "s"}.`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to list integration requests",
      };
    }
  },
});

export const getIntegrationRequestTool = tool({
  description:
    "Get a single integration request with source metadata before triage, editing, accepting, or declining.",
  inputSchema: z.object({
    requestId: z.string().describe("Integration request ID."),
    includeGitHubComments: z
      .boolean()
      .optional()
      .describe("Include GitHub comments when the request came from GitHub."),
  }),
  execute: async (
    { requestId, includeGitHubComments },
    { experimental_context: experimentalContext },
  ) => {
    try {
      const ctx = await getAuthenticatedContext(experimentalContext);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const request = await getIntegrationRequest(requestId, ctx);
      const comments =
        includeGitHubComments && request.provider === "github"
          ? await getRequestGitHubComments(requestId, ctx)
          : undefined;

      return {
        success: true,
        request,
        githubComments: comments,
        message: `Retrieved request "${request.title}".`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to get integration request",
      };
    }
  },
});

export const updateIntegrationRequestTool = tool({
  description:
    "Update fields on a pending integration request before accepting it as a story. Requires explicit confirmation.",
  inputSchema: z.object({
    requestId: z.string().describe("Integration request ID."),
    confirmed: z
      .boolean()
      .optional()
      .describe("Must be true after the user explicitly confirms the update."),
    title: z.string().optional(),
    description: z.string().optional(),
    statusId: z.string().optional(),
    priority: prioritySchema.optional(),
    assigneeId: z.string().optional(),
    estimateValue: z
      .number()
      .int()
      .optional()
      .describe("Canonical estimate value for the team's estimation scheme."),
    objectiveId: z.string().optional(),
    keyResultId: z.string().optional(),
    sprintId: z.string().optional(),
    startDate: z.string().optional(),
    endDate: z.string().optional(),
  }),
  execute: async (
    {
      requestId,
      confirmed,
      title,
      description,
      statusId,
      priority,
      assigneeId,
      estimateValue,
      objectiveId,
      keyResultId,
      sprintId,
      startDate,
      endDate,
    },
    { experimental_context: experimentalContext },
  ) => {
    try {
      if (!confirmed) {
        return requireToolConfirmation("update this integration request");
      }

      const ctx = await getAuthenticatedContext(experimentalContext);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const result = await updateIntegrationRequestAction(
        requestId,
        {
          title: normalizeOptionalString(title),
          description: normalizeOptionalString(description),
          statusId: normalizeOptionalString(statusId),
          priority,
          assigneeId: normalizeOptionalString(assigneeId),
          estimateValue,
          objectiveId: normalizeOptionalString(objectiveId),
          keyResultId: normalizeOptionalString(keyResultId),
          sprintId: normalizeOptionalString(sprintId),
          startDate: normalizeOptionalString(startDate),
          endDate: normalizeOptionalString(endDate),
        },
        ctx.workspaceSlug,
      );

      if (result.error?.message) {
        return { success: false, error: result.error.message };
      }

      return {
        success: true,
        request: result.data,
        message: "Integration request updated.",
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to update integration request",
      };
    }
  },
});

export const acceptIntegrationRequestTool = tool({
  description:
    "Accept a pending integration request and create the corresponding story. Requires explicit confirmation.",
  inputSchema: z.object({
    requestId: z.string().describe("Integration request ID."),
    confirmed: z
      .boolean()
      .optional()
      .describe("Must be true after the user explicitly confirms accepting."),
  }),
  execute: async (
    { requestId, confirmed },
    { experimental_context: experimentalContext },
  ) => {
    try {
      if (!confirmed) {
        return requireToolConfirmation("accept this integration request");
      }

      const ctx = await getAuthenticatedContext(experimentalContext);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const workspace = await getWorkspace(ctx);
      if (workspace.userRole === "guest") {
        return { success: false, error: "Guests cannot accept requests" };
      }

      const result = await acceptIntegrationRequestAction(
        requestId,
        ctx.workspaceSlug,
      );

      if (result.error?.message) {
        return { success: false, error: result.error.message };
      }

      return {
        success: true,
        request: result.data,
        message: "Integration request accepted.",
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to accept integration request",
      };
    }
  },
});

export const declineIntegrationRequestTool = tool({
  description:
    "Decline a pending integration request. Requires explicit confirmation.",
  inputSchema: z.object({
    requestId: z.string().describe("Integration request ID."),
    confirmed: z
      .boolean()
      .optional()
      .describe("Must be true after the user explicitly confirms declining."),
  }),
  execute: async (
    { requestId, confirmed },
    { experimental_context: experimentalContext },
  ) => {
    try {
      if (!confirmed) {
        return requireToolConfirmation("decline this integration request");
      }

      const ctx = await getAuthenticatedContext(experimentalContext);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const workspace = await getWorkspace(ctx);
      if (workspace.userRole === "guest") {
        return { success: false, error: "Guests cannot decline requests" };
      }

      const result = await declineIntegrationRequestAction(
        requestId,
        ctx.workspaceSlug,
      );

      if (result.error?.message) {
        return { success: false, error: result.error.message };
      }

      return {
        success: true,
        request: result.data,
        message: "Integration request declined.",
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to decline integration request",
      };
    }
  },
});

export const acceptAllIntegrationRequestsTool = tool({
  description:
    "Accept every pending integration request in a team. Requires explicit confirmation.",
  inputSchema: z.object({
    teamId: z.string().describe("Team ID."),
    confirmed: z
      .boolean()
      .optional()
      .describe(
        "Must be true after the user explicitly confirms accepting all pending requests.",
      ),
  }),
  execute: async (
    { teamId, confirmed },
    { experimental_context: experimentalContext },
  ) => {
    try {
      if (!confirmed) {
        return requireToolConfirmation(
          "accept all pending integration requests in this team",
        );
      }

      const ctx = await getAuthenticatedContext(experimentalContext);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const workspace = await getWorkspace(ctx);
      if (workspace.userRole === "guest") {
        return { success: false, error: "Guests cannot accept requests" };
      }

      const result = await acceptAllIntegrationRequestsAction(
        teamId,
        ctx.workspaceSlug,
      );

      if (result.error?.message) {
        return { success: false, error: result.error.message };
      }

      return {
        success: true,
        result: result.data,
        message: `Accepted ${result.data?.count ?? 0} integration request${result.data?.count === 1 ? "" : "s"}.`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to accept all integration requests",
      };
    }
  },
});

export const declineAllIntegrationRequestsTool = tool({
  description:
    "Decline every pending integration request in a team. Requires explicit confirmation.",
  inputSchema: z.object({
    teamId: z.string().describe("Team ID."),
    confirmed: z
      .boolean()
      .optional()
      .describe(
        "Must be true after the user explicitly confirms declining all pending requests.",
      ),
  }),
  execute: async (
    { teamId, confirmed },
    { experimental_context: experimentalContext },
  ) => {
    try {
      if (!confirmed) {
        return requireToolConfirmation(
          "decline all pending integration requests in this team",
        );
      }

      const ctx = await getAuthenticatedContext(experimentalContext);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const workspace = await getWorkspace(ctx);
      if (workspace.userRole === "guest") {
        return { success: false, error: "Guests cannot decline requests" };
      }

      const result = await declineAllIntegrationRequestsAction(
        teamId,
        ctx.workspaceSlug,
      );

      if (result.error?.message) {
        return { success: false, error: result.error.message };
      }

      return {
        success: true,
        result: result.data,
        message: `Declined ${result.data?.count ?? 0} integration request${result.data?.count === 1 ? "" : "s"}.`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to decline all integration requests",
      };
    }
  },
});

export const getRequestGitHubCommentsTool = tool({
  description:
    "Get GitHub comments attached to a GitHub integration request before triage.",
  inputSchema: z.object({
    requestId: z.string().describe("Integration request ID."),
  }),
  execute: async (
    { requestId },
    { experimental_context: experimentalContext },
  ) => {
    try {
      const ctx = await getAuthenticatedContext(experimentalContext);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const comments = await getRequestGitHubComments(requestId, ctx);

      return {
        success: true,
        comments,
        count: comments.length,
        message: `Found ${comments.length} GitHub comment${comments.length === 1 ? "" : "s"}.`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to get request GitHub comments",
      };
    }
  },
});

export const postRequestGitHubCommentTool = tool({
  description:
    "Post a comment to the GitHub issue linked to an integration request. Requires explicit confirmation.",
  inputSchema: z.object({
    requestId: z.string().describe("Integration request ID."),
    body: z.string().describe("GitHub comment body."),
    confirmed: z
      .boolean()
      .optional()
      .describe("Must be true after the user explicitly confirms posting."),
  }),
  execute: async (
    { requestId, body, confirmed },
    { experimental_context: experimentalContext },
  ) => {
    try {
      if (!confirmed) {
        return requireToolConfirmation(
          "post this comment to the request's GitHub issue",
        );
      }

      const ctx = await getAuthenticatedContext(experimentalContext);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const result = await postRequestGitHubCommentAction(
        requestId,
        { body },
        ctx.workspaceSlug,
      );

      if (result.error?.message) {
        return { success: false, error: result.error.message };
      }

      return {
        success: true,
        message: "GitHub comment posted.",
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to post request GitHub comment",
      };
    }
  },
});
