import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getSprintDetails } from "@/modules/sprints/queries/get-sprint-details";
import { getSprintAnalytics } from "@/modules/sprints/queries/get-sprint-analytics";

export const getSprintDetailsTool = tool({
  description:
    "Get detailed information about a specific sprint including its stories, progress, and metadata.",
  inputSchema: z.object({
    sprintId: z.string().describe("Sprint ID to get details for (required)"),
  }),

  execute: async ({ sprintId }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access sprint details",
        };
      }

      const [sprint, analytics] = await Promise.all([
        getSprintDetails(sprintId, session),
        getSprintAnalytics(sprintId, session),
      ]);

      return {
        success: true,
        sprint,
        analyticsReport: analytics,
        message: `Retrieved details for sprint "${sprint.name}".`,
      };
    } catch (error) {
      return {
        success: false,
        error: "Failed to get sprint details",
      };
    }
  },
});
