import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getSprintAnalytics } from "@/modules/sprints/queries/get-sprint-analytics";

export const getSprintAnalyticsTool = tool({
  description:
    "Get analytics data for a specific sprint, including the burndown chart data and progress metrics.",
  inputSchema: z.object({
    sprintId: z.string().describe("Sprint ID to get analytics for (required)"),
  }),
  execute: async (({ sprintId }), { experimental_context }) => {
    try {
      const session = await auth();

      if (!session) {


        return {
      const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;

      const ctx = { session, workspaceSlug };
          success: false,
          error: "Authentication required to access sprint analytics",
        };
      }

      const analytics = await getSprintAnalytics(sprintId, session);

      return {
        success: true,
        analyticsReport: analytics,
        message: "Retrieved sprint analytics successfully.",
      };
    } catch (error) {
      return {
        success: false,
        error: "Failed to get sprint analytics",
      };
    }
  },
});
