import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { updateStoryAction } from "@/modules/story/actions/update-story";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

export const updateStory = tool({
  description:
    "Update an existing story. Only admins and members can update stories.",
  inputSchema: z.object({
    storyId: z.string().describe("Story ID to update (required)"),
    title: z.string().optional().describe("Updated title"),
    description: z.string().optional().describe("Updated description"),
    descriptionHTML: z.string().optional().describe("Updated description HTML"),
    statusId: z.string().optional().describe("Updated status ID"),
    assigneeId: z.string().optional().describe("Updated assignee ID"),
    priority: z
      .enum(["No Priority", "Low", "Medium", "High", "Urgent"])
      .optional()
      .describe("Updated priority"),
    sprintId: z.string().optional().describe("Updated sprint ID"),
    objectiveId: z.string().optional().describe("Updated objective ID"),
    startDate: z.string().optional().describe("Updated start date"),
    endDate: z.string().optional().describe("Updated end date"),
  }),

  execute: async ({
    storyId,
    title,
    description,
    descriptionHTML,
    statusId,
    assigneeId,
    priority,
    sprintId,
    objectiveId,
    startDate,
    endDate,
  }, { experimental_context }) => {
    try {
      const session = await auth();

      if (!session) {


        return {
          success: false,
          error: "Authentication required to update stories",
        };
      }

      const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;

      const ctx = { session, workspaceSlug };

      const workspace = await getWorkspace(ctx);
      const userRole = workspace.userRole;

      if (userRole === "guest") {
        return {
          success: false,
          error: "Only admins and members can update stories",
        };
      }

      const updateData = {
        title,
        description,
        descriptionHTML,
        statusId,
        assigneeId,
        priority,
        sprintId,
        objectiveId,
        startDate,
        endDate,
      };

      const result = await updateStoryAction(storyId, updateData, workspaceSlug);

      if (result.error?.message) {
        return {
          success: false,
          error: result.error.message,
        };
      }

      return {
        success: true,
        message: "Story updated successfully.",
      };
    } catch (error) {
      return {
        success: false,
        error: "Failed to update story",
      };
    }
  },
});
