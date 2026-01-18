import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getObjectiveAnalytics } from "@/modules/objectives/queries/get-objective-analytics";

export const objectiveAnalyticsTool = tool({
  description:
    "Get analytics and progress data for a specific objective, progress chart, team allocation, priority breakdown, progress breakdown, and health metrics.",
  inputSchema: z.object({
    objectiveId: z
      .string()
      .describe("Objective ID to get analytics for (required)"),
  }),

  execute: async ({ objectiveId }, { experimental_context }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access objective analytics",
        };
      }

      const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;

      const ctx = { session, workspaceSlug };

      const analytics = await getObjectiveAnalytics(objectiveId, ctx);

      return {
        success: true,
        analytics: {
          objectiveId: analytics.objectiveId,
          priorityBreakdown: analytics.priorityBreakdown,
          progressBreakdown: analytics.progressBreakdown,
          teamAllocation: analytics.teamAllocation,
          progressChart: analytics.progressChart,
        },
        message: `Retrieved analytics for objective.`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to get objective analytics",
      };
    }
  },
});
