import { tool } from "ai";
import { normalizeOptionalString } from "../normalize-input";
import { resolvePaginationInput } from "../tool-helpers";
import {
  getIntegrationRequestInputSchema,
  listIntegrationRequestsInputSchema,
  toIntegrationRequestSummary,
} from "./contracts";
import type { IntegrationRequestToolDependencies } from "./contracts";
import { getAuthenticatedIntegrationRequestContext } from "./context";

export const createIntegrationRequestReadTools = (
  dependencies: Pick<
    IntegrationRequestToolDependencies,
    | "getIntegrationRequest"
    | "getRequestGitHubComments"
    | "getTeamIntegrationRequestsPage"
  >,
) => {
  const listIntegrationRequestsTool = tool({
    description:
      "List integration requests for one or more teams. Use this for GitHub, Slack, or Intercom request triage before accepting or declining requests.",
    inputSchema: listIntegrationRequestsInputSchema,
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
        const ctx =
          await getAuthenticatedIntegrationRequestContext(experimentalContext);
        if ("error" in ctx) return { success: false, error: ctx.error };

        const pagination = resolvePaginationInput({ page, pageSize });
        const pages = await Promise.all(
          teamIds.map(async (teamId) => {
            const response = await dependencies.getTeamIntegrationRequestsPage(
              teamId,
              ctx,
              status,
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
              requests: response.requests.map(toIntegrationRequestSummary),
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

  const getIntegrationRequestTool = tool({
    description:
      "Get a single integration request with source metadata before triage, editing, accepting, or declining.",
    inputSchema: getIntegrationRequestInputSchema,
    execute: async (
      { requestId, includeGitHubComments },
      { experimental_context: experimentalContext },
    ) => {
      try {
        const ctx =
          await getAuthenticatedIntegrationRequestContext(experimentalContext);
        if ("error" in ctx) return { success: false, error: ctx.error };

        const request = await dependencies.getIntegrationRequest(
          requestId,
          ctx,
        );
        const comments =
          includeGitHubComments && request.provider === "github"
            ? await dependencies.getRequestGitHubComments(requestId, ctx)
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

  const getRequestGitHubCommentsTool = tool({
    description:
      "Get GitHub comments attached to a GitHub integration request before triage.",
    inputSchema: getIntegrationRequestInputSchema.pick({ requestId: true }),
    execute: async (
      { requestId },
      { experimental_context: experimentalContext },
    ) => {
      try {
        const ctx =
          await getAuthenticatedIntegrationRequestContext(experimentalContext);
        if ("error" in ctx) return { success: false, error: ctx.error };

        const comments = await dependencies.getRequestGitHubComments(
          requestId,
          ctx,
        );

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

  return {
    getIntegrationRequestTool,
    getRequestGitHubCommentsTool,
    listIntegrationRequestsTool,
  };
};
