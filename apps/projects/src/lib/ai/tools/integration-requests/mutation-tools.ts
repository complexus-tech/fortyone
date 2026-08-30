import { tool } from "ai";
import { z } from "zod";
import { normalizeTimeNeededPatch } from "@/lib/time-needed";
import { normalizeOptionalString } from "../normalize-input";
import { requireToolConfirmation } from "../tool-helpers";
import { updateIntegrationRequestInputSchema } from "./contracts";
import type {
  IntegrationRequestToolActionResult,
  IntegrationRequestToolBulkResult,
  IntegrationRequestToolDependencies,
} from "./contracts";
import { getIntegrationRequestGuestAccessError } from "./authorization";
import { getAuthenticatedIntegrationRequestContext } from "./context";

const requestConfirmationSchema = (description: string) =>
  z.object({
    requestId: z.string().describe("Integration request ID."),
    confirmed: z.boolean().optional().describe(description),
  });

const teamConfirmationSchema = (description: string) =>
  z.object({
    teamId: z.string().describe("Team ID."),
    confirmed: z.boolean().optional().describe(description),
  });

const toBulkMutationResult = (
  result: IntegrationRequestToolActionResult<IntegrationRequestToolBulkResult>,
  {
    missingDataError,
    pastTense,
  }: {
    missingDataError: string;
    pastTense: "Accepted" | "Declined";
  },
) => {
  if (result.error?.message) {
    return { success: false, error: result.error.message };
  }

  if (!result.data) {
    return {
      success: false,
      error: missingDataError,
    };
  }

  const completedSuccessfully = result.data.failedCount === 0;

  return {
    success: completedSuccessfully,
    partial: result.data.partial,
    result: result.data,
    message: completedSuccessfully
      ? `${pastTense} ${result.data.succeededCount} of ${result.data.totalCount} integration requests.`
      : `${pastTense} ${result.data.succeededCount} of ${result.data.totalCount} integration requests; ${result.data.failedCount} failed.`,
  };
};

export const createIntegrationRequestMutationTools = (
  dependencies: Pick<
    IntegrationRequestToolDependencies,
    | "acceptAllIntegrationRequests"
    | "acceptIntegrationRequest"
    | "declineAllIntegrationRequests"
    | "declineIntegrationRequest"
    | "postRequestGitHubComment"
    | "updateIntegrationRequest"
  >,
) => {
  const updateIntegrationRequestTool = tool({
    description:
      "Update fields on a pending integration request before accepting it as a story. Requires explicit confirmation.",
    inputSchema: updateIntegrationRequestInputSchema,
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
        estimatedDurationMinutes,
        minimumFocusBlockMinutes,
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

        const ctx =
          await getAuthenticatedIntegrationRequestContext(experimentalContext);
        if ("error" in ctx) return { success: false, error: ctx.error };

        const timeNeededPatch = normalizeTimeNeededPatch(
          estimatedDurationMinutes,
          minimumFocusBlockMinutes,
        );

        const result = await dependencies.updateIntegrationRequest(
          requestId,
          {
            title: normalizeOptionalString(title),
            description: normalizeOptionalString(description),
            statusId: normalizeOptionalString(statusId),
            priority,
            assigneeId: normalizeOptionalString(assigneeId),
            estimateValue,
            ...timeNeededPatch,
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

  const acceptIntegrationRequestTool = tool({
    description:
      "Accept a pending integration request and create the corresponding story. Requires explicit confirmation.",
    inputSchema: requestConfirmationSchema(
      "Must be true after the user explicitly confirms accepting.",
    ),
    execute: async (
      { requestId, confirmed },
      { experimental_context: experimentalContext },
    ) => {
      try {
        if (!confirmed) {
          return requireToolConfirmation("accept this integration request");
        }

        const ctx =
          await getAuthenticatedIntegrationRequestContext(experimentalContext);
        if ("error" in ctx) return { success: false, error: ctx.error };

        const guestAccessError = await getIntegrationRequestGuestAccessError(
          ctx,
          "Guests cannot accept requests",
        );
        if (guestAccessError) {
          return { success: false, error: guestAccessError };
        }

        const result = await dependencies.acceptIntegrationRequest(
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

  const declineIntegrationRequestTool = tool({
    description:
      "Decline a pending integration request. Requires explicit confirmation.",
    inputSchema: requestConfirmationSchema(
      "Must be true after the user explicitly confirms declining.",
    ),
    execute: async (
      { requestId, confirmed },
      { experimental_context: experimentalContext },
    ) => {
      try {
        if (!confirmed) {
          return requireToolConfirmation("decline this integration request");
        }

        const ctx =
          await getAuthenticatedIntegrationRequestContext(experimentalContext);
        if ("error" in ctx) return { success: false, error: ctx.error };

        const guestAccessError = await getIntegrationRequestGuestAccessError(
          ctx,
          "Guests cannot decline requests",
        );
        if (guestAccessError) {
          return { success: false, error: guestAccessError };
        }

        const result = await dependencies.declineIntegrationRequest(
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

  const acceptAllIntegrationRequestsTool = tool({
    description:
      "Accept every pending integration request in a team. Requires explicit confirmation.",
    inputSchema: teamConfirmationSchema(
      "Must be true after the user explicitly confirms accepting all pending requests.",
    ),
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

        const ctx =
          await getAuthenticatedIntegrationRequestContext(experimentalContext);
        if ("error" in ctx) return { success: false, error: ctx.error };

        const guestAccessError = await getIntegrationRequestGuestAccessError(
          ctx,
          "Guests cannot accept requests",
        );
        if (guestAccessError) {
          return { success: false, error: guestAccessError };
        }

        return toBulkMutationResult(
          await dependencies.acceptAllIntegrationRequests(
            teamId,
            ctx.workspaceSlug,
          ),
          {
            missingDataError:
              "The bulk accept completed without an itemized result",
            pastTense: "Accepted",
          },
        );
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

  const declineAllIntegrationRequestsTool = tool({
    description:
      "Decline every pending integration request in a team. Requires explicit confirmation.",
    inputSchema: teamConfirmationSchema(
      "Must be true after the user explicitly confirms declining all pending requests.",
    ),
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

        const ctx =
          await getAuthenticatedIntegrationRequestContext(experimentalContext);
        if ("error" in ctx) return { success: false, error: ctx.error };

        const guestAccessError = await getIntegrationRequestGuestAccessError(
          ctx,
          "Guests cannot decline requests",
        );
        if (guestAccessError) {
          return { success: false, error: guestAccessError };
        }

        return toBulkMutationResult(
          await dependencies.declineAllIntegrationRequests(
            teamId,
            ctx.workspaceSlug,
          ),
          {
            missingDataError:
              "The bulk decline completed without an itemized result",
            pastTense: "Declined",
          },
        );
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

  const postRequestGitHubCommentTool = tool({
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

        const ctx =
          await getAuthenticatedIntegrationRequestContext(experimentalContext);
        if ("error" in ctx) return { success: false, error: ctx.error };

        const result = await dependencies.postRequestGitHubComment(
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

  return {
    acceptAllIntegrationRequestsTool,
    acceptIntegrationRequestTool,
    declineAllIntegrationRequestsTool,
    declineIntegrationRequestTool,
    postRequestGitHubCommentTool,
    updateIntegrationRequestTool,
  };
};
