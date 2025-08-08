import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { restoreStoryAction } from "@/modules/story/actions/restore-story";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

export const restoreStory = tool({
  description:
    "Restore a deleted story. Only admins and members can restore deleted stories.",
  inputSchema: z.object({
    storyId: z.string().describe("Story ID to restore (required)"),
  }),

  execute: async ({ storyId }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to restore stories",
        };
      }

      const workspace = await getWorkspace(session);
      const userRole = workspace.userRole;

      // Only admins can restore stories
      if (userRole === "guest") {
        return {
          success: false,
          error: "Only admins and members can restore deleted stories",
        };
      }

      const result = await restoreStoryAction(storyId);

      if (result.error) {
        return {
          success: false,
          error: result.error.message || "Failed to restore story",
        };
      }

      return {
        success: true,
        message: "Story restored successfully.",
      };
    } catch (error) {
      return {
        success: false,
        error: "Failed to restore story",
      };
    }
  },
});
