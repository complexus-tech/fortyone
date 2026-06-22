import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { createStoryAction } from "@/modules/story/actions/create-story";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import { requireToolConfirmation } from "../tool-helpers";
import { normalizeStoryInput } from "./normalize-story-input";

export const createStory = tool({
  description:
    "Create a new story. Guests cannot create stories. Members and admins can create stories for teams they belong to.",
  inputSchema: z.object({
    title: z.string().describe("Story title (required)"),
    confirmed: z
      .boolean()
      .optional()
      .describe(
        "Must be true after the user explicitly confirms creating the story.",
      ),
    description: z.string().optional().describe("Story description"),
    descriptionHTML: z
      .string()
      .optional()
      .describe(
        "Story description HTML (Always provided and properly formatted if description is provided)",
      ),
    teamId: z
      .string()
      .describe("Team ID where story belongs (required) (UUID)"),
    statusId: z
      .string()
      .describe(
        "Initial status ID (required) (UUID) always use statuses tool to get the statuses",
      ),
    assigneeId: z.string().optional().describe("Assignee user ID (UUID)"),
    priority: z
      .enum(["No Priority", "Low", "Medium", "High", "Urgent"])
      .default("No Priority")
      .describe("Story priority (required)"),
    estimateValue: z
      .number()
      .int()
      .optional()
      .describe(
        "Canonical estimate value for the team's estimation scheme. Use the team's configured estimate scale.",
      ),
    labelIds: z
      .array(z.string())
      .optional()
      .describe("Label IDs to attach to the story."),
    sprintId: z
      .string()
      .optional()
      .describe("Sprint ID to assign story (UUID)"),
    objectiveId: z
      .string()
      .optional()
      .describe("Objective ID to assign story (UUID)"),
    parentId: z
      .string()
      .optional()
      .describe("Parent story ID for sub-stories (UUID)"),
    startDate: z
      .string()
      .optional()
      .describe("Story start date (ISO date string e.g 2005-06-13)"),
    endDate: z
      .string()
      .optional()
      .describe("Story end date (ISO date string e.g 2005-06-13)"),
  }),

  execute: async (
    {
      title,
      confirmed,
      description,
      descriptionHTML,
      teamId,
      statusId,
      assigneeId,
      priority,
      estimateValue,
      labelIds,
      sprintId,
      objectiveId,
      parentId,
      startDate,
      endDate,
    },
    { experimental_context: experimentalContext },
  ) => {
    try {
      if (!confirmed) {
        return requireToolConfirmation("create this story");
      }

      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to create stories",
        };
      }

      const workspaceSlug = (experimentalContext as { workspaceSlug: string })
        .workspaceSlug;

      const ctx = { session, workspaceSlug };

      const workspace = await getWorkspace(ctx);
      const userRole = workspace.userRole;

      // Check permissions for guests
      if (userRole === "guest") {
        return {
          success: false,
          error: "Guests can only create stories for teams they belong to",
        };
      }

      const storyData = normalizeStoryInput({
        title,
        description,
        descriptionHTML,
        teamId,
        statusId,
        assigneeId,
        priority,
        estimateValue,
        labelIds,
        sprintId,
        objectiveId,
        parentId,
        startDate,
        endDate,
      });

      const result = await createStoryAction(storyData, workspaceSlug);

      if (result.error?.message) {
        return {
          success: false,
          error: result.error.message || "Failed to create story",
        };
      }

      if (!result.data?.id) {
        return {
          success: false,
          error: "Story creation did not return a created story.",
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
