import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { bulkDeleteAction } from "@/modules/stories/actions/bulk-delete-stories";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

export const bulkDeleteStories = tool({
  description:
    "Bulk delete multiple stories at once. Only admins and members can perform bulk operations.",
  parameters: z.object({
    storyIds: z
      .array(z.string())
      .describe("Array of story IDs to delete (required)"),
  }),

  execute: async ({ storyIds }) => {
    try {
      const session = await auth();
      if (!session) {
        return {
          success: false,
          error: "Authentication required to bulk delete stories",
        };
      }

      const workspace = await getWorkspace(session);
      const userRole = workspace.userRole;

      if (userRole === "guest") {
        return {
          success: false,
          error: "Only admins and members can perform bulk story operations",
        };
      }

      const result = await bulkDeleteAction(storyIds);

      if (result.error?.message) {
        return {
          success: false,
          error: result.error.message || "Failed to bulk delete stories",
        };
      }

      return {
        success: true,
        message: `Successfully deleted ${storyIds.length} stories.`,
      };
    } catch (error) {
      return {
        success: false,
        error: "Failed to bulk delete stories",
      };
    }
  },
});
