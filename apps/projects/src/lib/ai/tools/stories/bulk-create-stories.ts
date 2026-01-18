import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { createStoryAction } from "@/modules/story/actions/create-story";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

export const bulkCreateStories = tool({
  description:
    "Bulk create multiple stories at once. Only admins and members can perform bulk operations. always use this tool when creating multiple stories at once but create in batches of 10.",
  inputSchema: z.object({
    storiesData: z
      .array(
        z.object({
          title: z.string().describe("Story title (required)"),
          description: z.string().optional().describe("Story description"),
          descriptionHTML: z
            .string()
            .optional()
            .describe("Story description HTML"),
          teamId: z.string().describe("Team ID where story belongs (required)"),
          statusId: z.string().describe("Initial status ID (required)"),
          assigneeId: z.string().optional().describe("Assignee user ID"),
          priority: z
            .enum(["No Priority", "Low", "Medium", "High", "Urgent"])
            .describe("Story priority (required)"),
          sprintId: z.string().optional().describe("Sprint ID to assign story"),
          objectiveId: z
            .string()
            .optional()
            .describe("Objective ID to assign story"),
          parentId: z
            .string()
            .optional()
            .describe("Parent story ID for sub-stories"),
          startDate: z
            .string()
            .optional()
            .describe("Story start date (ISO  date string e.g 2005-06-13)"),
          endDate: z
            .string()
            .optional()
            .describe("Story end date (ISO  date string e.g 2005-06-13)"),
        }),
      )
      .describe("Array of story data for bulk creation (required)"),
  }),

  execute: async ({ storiesData }, { experimental_context }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to create stories",
        };
      }

      const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;

      const ctx = { session, workspaceSlug };

      const workspace = await getWorkspace(ctx);
      const userRole = workspace.userRole;

      // Only admins can perform bulk operations
      if (userRole === "guest") {
        return {
          success: false,
          error: "Only admins and members can perform bulk story operations",
        };
      }

      const results = await Promise.all(
        storiesData.map((storyData) => createStoryAction(storyData, workspaceSlug)),
      );

      const successCount = results.filter((r) => !r.error).length;
      const errorCount = results.filter((r) => r.error).length;
      const createdStories = results.filter((r) => !r.error).map((r) => r.data);

      return {
        success: true,
        createdCount: successCount,
        errorCount,
        stories: createdStories,
        message: `Successfully created ${successCount} stories.${errorCount > 0 ? ` ${errorCount} stories failed to create.` : ""}`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to bulk create stories",
      };
    }
  },
});
