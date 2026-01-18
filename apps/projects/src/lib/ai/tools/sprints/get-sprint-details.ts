import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getSprint } from "@/modules/sprints/queries/get-sprint-details";
import { getSprintAnalytics } from "@/modules/sprints/queries/get-sprint-analytics";

export const getSprintDetailsTool = tool({
  description:
    "Get detailed information about a specific sprint including its stories, progress, and metadata.",
  inputSchema: z.object({
    sprintId: z.string().describe("Sprint ID to get details for (required)"),
  }),

  execute: async ({ sprintId }, { experimental_context }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access sprint details",
        };
      }

      const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;

      const ctx = { session, workspaceSlug };

      const sprint = await getSprint(sprintId, ctx);

      return {
        success: true,
        sprint,
        message: `Retrieved details for sprint "${sprint?.name}".`,
      };
    } catch (error) {
      return {
        success: false,
        error: "Failed to get sprint details",
      };
    }
  },
});
