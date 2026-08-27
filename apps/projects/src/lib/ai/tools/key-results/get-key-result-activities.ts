import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getKeyResultActivities } from "@/modules/objectives/queries/get-key-result-activities";
import { describeBackendPage } from "../tool-helpers";

export const getKeyResultActivitiesTool = tool({
  description:
    "View a paginated key result activity timeline, including metadata that says when more history is available.",
  inputSchema: z.object({
    keyResultId: z.string().describe("Key Result ID for activity operations"),
    page: z.number().min(1).optional().describe("Page number for pagination"),
    pageSize: z
      .number()
      .min(1)
      .max(100)
      .optional()
      .describe("Number of activities per page"),
  }),

  execute: async (
    { keyResultId, page = 1, pageSize = 20 },
    { experimental_context: experimentalContext },
  ) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access key result activities",
        };
      }

      const workspaceSlug = (experimentalContext as { workspaceSlug: string })
        .workspaceSlug;

      const ctx = { session, workspaceSlug };

      const response = await getKeyResultActivities(
        keyResultId,
        page,
        pageSize,
        ctx,
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
        message: `Returned ${activities.length} key result activit${activities.length === 1 ? "y" : "ies"} from page ${pagination.page}${pagination.hasMore ? "; more pages are available" : ""}.`,
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : "An error occurred",
      };
    }
  },
});
