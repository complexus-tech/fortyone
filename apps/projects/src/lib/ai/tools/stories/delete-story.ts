import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { deleteStoryAction } from "@/modules/story/actions/delete-story";
import { isStoryDeletionOutcomeUncertainError } from "@/modules/story/actions/story-deletion-error";
import { requireToolConfirmation } from "../tool-helpers";

export const deleteStory = tool({
  description:
    "Delete one exact story. Resolve its current ID and human-readable title first; both are required so the approval screen can show a verifiable target. Only admins or story creators can delete stories.",
  inputSchema: z.object({
    storyId: z
      .string()
      .uuid("Story ID must be a valid UUID.")
      .describe("Exact story ID to delete."),
    storyTitle: z
      .string()
      .trim()
      .min(1, "Story title is required for approval.")
      .describe(
        "Current human-readable title of storyId. This is shown to the user before deletion alongside the exact ID.",
      ),
    confirmed: z
      .boolean()
      .optional()
      .describe("Must be true after the user explicitly confirms deletion."),
  }),

  execute: async (
    { storyId, storyTitle, confirmed },
    { experimental_context: experimentalContext },
  ) => {
    try {
      if (!confirmed) {
        return requireToolConfirmation("delete this story");
      }

      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to delete stories",
        };
      }

      const workspaceSlug = (experimentalContext as { workspaceSlug: string })
        .workspaceSlug;
      const result = await deleteStoryAction(storyId, workspaceSlug);

      if (result.error?.message) {
        return {
          success: false,
          error: result.error.message,
        };
      }

      return {
        success: true,
        message: `Story "${storyTitle}" deleted successfully.`,
      };
    } catch (error) {
      if (isStoryDeletionOutcomeUncertainError(error)) throw error;

      return {
        success: false,
        error:
          error instanceof Error ? error.message : "Failed to delete story",
      };
    }
  },
});
