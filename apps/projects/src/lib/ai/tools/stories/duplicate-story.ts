import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { duplicateStoryAction } from "@/modules/story/actions/duplicate-story";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

export const duplicateStory = tool({
  description:
    "Duplicate an existing story. Only admins and members can duplicate stories.",
  inputSchema: z.object({
    storyId: z.string().describe("Story ID to duplicate (required)"),
  }),

  execute: async (({ storyId }), { experimental_context }) => {
    try {
      const session = await auth();

      if (!session) {


        return {
      const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;

      const ctx = { session, workspaceSlug };
          success: false,
          error: "Authentication required to duplicate stories",
        };
      }

      const workspace = await getWorkspace(session);
      const userRole = workspace.userRole;

      if (userRole === "guest") {
        return {
          success: false,
          error:
            "You can only duplicate stories you created or if you're an admin",
        };
      }

      const result = await duplicateStoryAction(storyId);

      if (result.error) {
        return {
          success: false,
          error: result.error.message || "Failed to duplicate story",
        };
      }

      return {
        success: true,
        story: result.data,
        message: `Story "${result.data?.title}" duplicated successfully.`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error ? error.message : "Failed to duplicate story",
      };
    }
  },
});
