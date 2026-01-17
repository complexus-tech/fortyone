import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { updateLabelsAction } from "@/modules/story/actions/update-labels";
import { getStory } from "@/modules/story/queries/get-story";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

export const storyLabelsTool = tool({
  description: "Manage story labels: get, add, remove, set labels on stories.",
  inputSchema: z.object({
    action: z.enum([
      "get-story-labels",
      "set-story-labels",
      "add-labels-to-story",
      "remove-labels-from-story",
    ]),
    storyId: z.string(),
    labelIds: z.array(z.string()).optional(),
  }),
  execute: async (({ action, storyId, labelIds = [] }), { experimental_context }) => {
    try {
      const session = await auth();
      if (!session) return { success: false, error: "Authentication required" };
      const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;

      const ctx = { session, workspaceSlug };


      const workspace = await getWorkspace(session);
      const userRole = workspace.userRole;
      const isGuest = userRole === "guest";

      const story = await getStory(storyId, session);
      if (!story) return { success: false, error: "Story not found" };
      const currentLabelIds: string[] = story.labels ?? [];

      switch (action) {
        case "get-story-labels": {
          return {
            success: true,
            labels: currentLabelIds,
          };
        }
        case "set-story-labels": {
          if (isGuest)
            return {
              success: false,
              error: "Guests cannot modify story labels",
            };
          const res = await updateLabelsAction(storyId, labelIds);
          if (res.error) return { success: false, error: res.error.message };
          return { success: true, labelIds };
        }
        case "add-labels-to-story": {
          if (isGuest)
            return {
              success: false,
              error: "Guests cannot modify story labels",
            };

          const newLabelIds = Array.from(
            new Set([...currentLabelIds, ...labelIds]),
          );
          const res = await updateLabelsAction(storyId, newLabelIds);
          if (res.error) return { success: false, error: res.error.message };
          return { success: true, labelIds: newLabelIds };
        }
        case "remove-labels-from-story": {
          if (isGuest)
            return {
              success: false,
              error: "Guests cannot modify story labels",
            };

          const filtered = currentLabelIds.filter(
            (id) => !labelIds.includes(id),
          );
          const res = await updateLabelsAction(storyId, filtered);
          if (res.error) return { success: false, error: res.error.message };
          return { success: true, labelIds: filtered };
        }
        default:
          return { success: false, error: "Invalid action" };
      }
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : "An error occurred",
      };
    }
  },
});
