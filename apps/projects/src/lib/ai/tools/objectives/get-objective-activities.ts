import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getObjectiveActivities } from "@/modules/objectives/queries/get-objective-activities";

export const getObjectiveActivitiesTool = tool({
  description:
    "View objective activity timeline and changes: track who made what changes, when they happened, and provide detailed history of objective modifications.",
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

  execute: async ({ objectiveId, page = 1, pageSize = 20 }, { experimental_context }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access objective activities",
        };
      }

      const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;

      const ctx = { session, workspaceSlug };

      const response = await getObjectiveActivities(
        objectiveId,
        ctx,
        page,
        pageSize,
      );
      const activities = response.activities;

      return {
        success: true,
        activities,
        count: activities.length,
        pagination: response.pagination,
        message: `Found ${activities.length} activity${activities.length !== 1 ? "s" : ""} for this objective.`,
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : "An error occurred",
      };
    }
  },
});
