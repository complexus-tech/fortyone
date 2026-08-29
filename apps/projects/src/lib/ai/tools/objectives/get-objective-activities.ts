import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getObjectiveActivities } from "@/modules/objectives/queries/get-objective-activities";
import { describeBackendPage } from "../tool-helpers";

export const getObjectiveActivitiesTool = tool({
  description:
    "View a paginated objective activity timeline, including metadata that says when more history is available.",
  inputSchema: z.object({
    objectiveId: z.string().describe("Objective ID for activity operations"),
    page: z.number().min(1).optional().describe("Page number for pagination"),
    pageSize: z
      .number()
      .min(1)
      .max(100)
      .optional()
      .describe("Number of activities per page"),
  }),

  execute: async (
    { objectiveId, page = 1, pageSize = 20 },
    { experimental_context: experimentalContext },
  ) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access objective activities",
        };
      }

      const workspaceSlug = (experimentalContext as { workspaceSlug: string })
        .workspaceSlug;

      const ctx = { session, workspaceSlug };

      const response = await getObjectiveActivities(
        objectiveId,
        ctx,
        page,
        pageSize,
      );
      const activities = response.activities;
      const pagination = describeBackendPage(
        response.pagination,
        activities.length,
      );

      return {
        success: true,
        activities,
        count: activities.length,
        pagination,
        message: `Returned ${activities.length} objective activit${activities.length === 1 ? "y" : "ies"} from page ${pagination.page}${pagination.hasMore ? "; more pages are available" : ""}.`,
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : "An error occurred",
      };
    }
  },
});
