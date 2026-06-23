import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { createStoryAction } from "@/modules/story/actions/create-story";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import { requireToolConfirmation } from "../tool-helpers";
import { normalizeStoryInput } from "./normalize-story-input";

export const bulkCreateStories = tool({
  description:
    "Bulk create multiple stories at once. Only admins and members can perform bulk operations. always use this tool when creating multiple stories at once but create in batches of 10.",
  inputSchema: z.object({
    confirmed: z
      .boolean()
      .optional()
      .describe(
        "Must be true after the user explicitly confirms the bulk creation.",
      ),
    storiesData: z
      .array(
        z.object({
          title: z.string().describe("Story title (required)"),
          description: z
            .string()
            .nullable()
            .optional()
            .describe("Story description"),
          descriptionHTML: z
            .string()
            .nullable()
            .optional()
            .describe("Story description HTML"),
          teamId: z.string().describe("Team ID where story belongs (required)"),
          statusId: z.string().describe("Initial status ID (required)"),
          assigneeId: z
            .string()
            .nullable()
            .optional()
            .describe("Assignee user ID"),
          priority: z
            .enum(["No Priority", "Low", "Medium", "High", "Urgent"])
            .default("No Priority")
            .describe("Story priority (required)"),
          estimateValue: z
            .number()
            .int()
            .nullable()
            .optional()
            .describe(
              "Canonical estimate value for the team's estimation scheme. Use 1, 2, 3, 5, or 8. Use 0, null, or omit for unestimated work.",
            ),
          labelIds: z
            .array(z.string())
            .optional()
            .describe("Label IDs to attach to the story."),
          sprintId: z
            .string()
            .nullable()
            .optional()
            .describe("Sprint ID to assign story"),
          objectiveId: z
            .string()
            .nullable()
            .optional()
            .describe("Objective ID to assign story"),
          parentId: z
            .string()
            .nullable()
            .optional()
            .describe("Parent story ID for sub-stories"),
          startDate: z
            .string()
            .nullable()
            .optional()
            .describe("Story start date (ISO  date string e.g 2005-06-13)"),
          endDate: z
            .string()
            .nullable()
            .optional()
            .describe("Story end date (ISO  date string e.g 2005-06-13)"),
        }),
      )
      .describe("Array of story data for bulk creation (required)"),
  }),

  execute: async (
    { confirmed, storiesData },
    { experimental_context: experimentalContext },
  ) => {
    try {
      if (!confirmed) {
        return requireToolConfirmation("create these stories");
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

      // Only admins can perform bulk operations
      if (userRole === "guest") {
        return {
          success: false,
          error: "Only admins and members can perform bulk story operations",
        };
      }

      const normalizedStoriesData = storiesData.map((storyData) =>
        normalizeStoryInput(storyData),
      );

      const results = await Promise.all(
        normalizedStoriesData.map((storyData) =>
          createStoryAction(storyData, workspaceSlug),
        ),
      );

      const successCount = results.filter((r) => !r.error).length;
      const errorCount = results.filter((r) => r.error).length;
      const createdStories = results.filter((r) => !r.error).map((r) => r.data);
      const failedStories = results
        .map((result, index) => ({
          title: normalizedStoriesData[index].title,
          error: result.error?.message,
        }))
        .filter(
          (failure): failure is { title: string; error: string } =>
            typeof failure.error === "string" && failure.error.length > 0,
        );

      if (errorCount > 0) {
        return {
          success: false,
          createdCount: successCount,
          errorCount,
          stories: createdStories,
          failedStories,
          error: failedStories
            .map((failure) => `${failure.title}: ${failure.error}`)
            .join("; "),
          message: `Created ${successCount} stories. ${errorCount} stories failed to create.`,
        };
      }

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
