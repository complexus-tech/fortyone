import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { bulkUpdateAction } from "@/modules/stories/actions/bulk-update-stories";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

export const bulkUpdateStories = tool({
  description:
    "Bulk update multiple stories at once. Only admins and members can perform bulk operations.",
  parameters: z.object({
    storyIds: z
      .array(z.string())
      .describe("Array of story IDs to update (required)"),
    updateData: z
      .object({
        statusId: z
          .string()
          .optional()
          .describe("Updated status ID for all stories"),
        assigneeId: z
          .string()
          .optional()
          .describe("Updated assignee ID for all stories"),
        priority: z
          .enum(["No Priority", "Low", "Medium", "High", "Urgent"])
          .optional()
          .describe("Updated priority for all stories"),
        sprintId: z
          .string()
          .optional()
          .describe("Updated sprint ID for all stories"),
        objectiveId: z
          .string()
          .optional()
          .describe("Updated objective ID for all stories"),
      })
      .describe("Update data to apply to all stories (required)"),
  }),

  execute: async ({ storyIds, updateData }) => {
    try {
      const session = await auth();
      if (!session) {
        return {
          success: false,
          error: "Authentication required to bulk update stories",
        };
      }

      const workspace = await getWorkspace(session);
      const userRole = workspace.userRole;

      // Only admins can perform bulk operations
      if (userRole === "guest") {
        return {
          success: false,
          error: "Only admins and members can perform bulk story operations",
        };
      }

      const result = await bulkUpdateAction({
        storyIds,
        updates: updateData,
      });

      if (result.error?.message) {
        return {
          success: false,
          error: result.error.message || "Failed to bulk update stories",
        };
      }

      return {
        success: true,
        message: `Successfully updated ${storyIds.length} stories.`,
      };
    } catch (error) {
      return {
        success: false,
        error: "Failed to bulk update stories",
      };
    }
  },
});
