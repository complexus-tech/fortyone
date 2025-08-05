import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { createStoryAction } from "@/modules/story/actions/create-story";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

export const createStory = tool({
  description:
    "Create a new story. Guests cannot create stories. Members and admins can create stories for teams they belong to.",
  parameters: z.object({
    title: z.string().describe("Story title (required)"),
    description: z.string().optional().describe("Story description"),
    descriptionHTML: z.string().optional().describe("Story description HTML"),
    teamId: z.string().describe("Team ID where story belongs (required)"),
    statusId: z.string().describe("Initial status ID (required)"),
    assigneeId: z.string().optional().describe("Assignee user ID"),
    priority: z
      .enum(["No Priority", "Low", "Medium", "High", "Urgent"])
      .describe("Story priority (required)"),
    sprintId: z.string().optional().describe("Sprint ID to assign story"),
    objectiveId: z.string().optional().describe("Objective ID to assign story"),
    parentId: z.string().optional().describe("Parent story ID for sub-stories"),
    startDate: z.string().optional().describe("Story start date (ISO string)"),
    endDate: z.string().optional().describe("Story end date (ISO string)"),
  }),

  execute: async ({
    title,
    description,
    descriptionHTML,
    teamId,
    statusId,
    assigneeId,
    priority,
    sprintId,
    objectiveId,
    parentId,
    startDate,
    endDate,
  }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to create stories",
        };
      }

      const workspace = await getWorkspace(session);
      const userRole = workspace.userRole;

      // Check permissions for guests
      if (userRole === "guest") {
        return {
          success: false,
          error: "Guests can only create stories for teams they belong to",
        };
      }

      const storyData = {
        title,
        description,
        descriptionHTML,
        teamId,
        statusId,
        assigneeId,
        priority,
        sprintId,
        objectiveId,
        parentId,
        startDate,
        endDate,
      };

      const result = await createStoryAction(storyData);

      if (result.error?.message) {
        return {
          success: false,
          error: result.error.message || "Failed to create story",
        };
      }

      return {
        success: true,
        story: result.data,
        message: `Story "${title}" created successfully.`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error ? error.message : "Failed to create story",
      };
    }
  },
});
