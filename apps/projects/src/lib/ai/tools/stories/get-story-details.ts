import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getStory } from "@/modules/story/queries/get-story";

export const getStoryDetails = tool({
  description:
    "Get detailed information about a specific story including its metadata, assignments, and relationships.",
  inputSchema: z.object({
    storyId: z.string().describe("Story ID to get details for (required)"),
  }),

  execute: async ({ storyId }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access story details",
        };
      }

      const story = await getStory(storyId, session);

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
