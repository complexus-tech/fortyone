import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getStory, getStoryRef } from "@/modules/story/queries/get-story";
import { DetailedStory } from "@/modules/story/types";

export const getStoryDetails = tool({
  description:
    "Get detailed information about a specific story including its metadata, assignments, and relationships.",
  inputSchema: z.object({
    storyId: z
      .string()
      .optional()
      .describe(
        "Story ID to get details for (required if there is no storyRef).",
      ),
    storyRef: z
      .string()
      .optional()
      .describe(
        "Story reference to get details for (required if there is no storyId) format team code-sequence number. examples: WEB-123, WEB123, WEB 123. If storyId is provided, this field is ignored.",
      ),
  }),

  execute: async ({ storyId, storyRef }, { experimental_context }) => {
    if (!storyId && !storyRef) {
      return {
        success: false,
        error: "Either storyId or storyRef is required",
      };
    }
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access story details",
        };
      }

      const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;

      const ctx = { session, workspaceSlug };
      let story: DetailedStory | null | undefined = null;
      if (storyId) {
        story = await getStory(storyId, ctx);
      } else if (storyRef) {
        story = await getStoryRef(storyRef, ctx);
      }

      if (!story) {
        return {
          success: false,
          error: "Story not found",
        };
      }

      return {
        success: true,
        story,
        message: `Retrieved details for story "${story.title}".`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to get story details",
      };
    }
  },
});
