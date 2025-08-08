import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { deleteStoryAction } from "@/modules/story/actions/delete-story";
import { getStory } from "@/modules/story/queries/get-story";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

export const deleteStory = tool({
  description:
    "Delete a story. Only admins or story creators can delete stories.",
  inputSchema: z.object({
    storyId: z.string().describe("Story ID to delete (required)"),
  }),

  execute: async ({ storyId }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to delete stories",
        };
      }

      const workspace = await getWorkspace(session);
      const userRole = workspace.userRole;
      const userId = session.user!.id;

      // Check if user can delete this story
      const story = await getStory(storyId, session);
      if (!story) {
        return {
          success: false,
          error: "Story not found",
        };
      }

      const canDelete = userRole === "admin" || story.reporterId === userId;

      if (!canDelete) {
        return {
          success: false,
          error:
            "You can only delete stories you created or if you're an admin",
        };
      }

      const result = await deleteStoryAction(storyId);

      if (result.error?.message) {
        return {
          success: false,
          error: result.error.message,
        };
      }

      return {
        success: true,
        message: `Story "${story.title}" deleted successfully.`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error ? error.message : "Failed to delete story",
      };
    }
  },
});
