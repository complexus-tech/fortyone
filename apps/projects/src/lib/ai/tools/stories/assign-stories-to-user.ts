import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import { bulkUpdateAction } from "@/modules/stories/actions/bulk-update-stories";

export const assignStoriesToUser = tool({
  description:
    "Assign multiple stories to a specific user. Only admins or members can assign stories.",
  inputSchema: z.object({
    storyIds: z
      .array(z.string())
      .describe("Array of story IDs to assign (required)"),
    assigneeId: z.string().describe("User ID to assign stories to (required)"),
  }),

  execute: async ({ storyIds, assigneeId }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to assign stories",
        };
      }

      const workspace = await getWorkspace(session);
      const userRole = workspace.userRole;
      if (userRole === "guest") {
        return {
          success: false,
          error: "Only admins or members can assign stories",
        };
      }

      const results = await bulkUpdateAction({
        storyIds,
        updates: { assigneeId },
      });

      if (results.error?.message) {
        return {
          success: false,
          error: results.error.message,
        };
      }

      return {
        success: true,
        message: `Successfully assigned ${storyIds.length} stories to user.`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error ? error.message : "Failed to assign stories",
      };
    }
  },
});
