import { z } from "zod";
import { tool } from "ai";

export const navigation = tool({
  description: "Navigate to different pages of the application",
  parameters: z.object({
    destination: z
      .enum([
        "my-work",
        "summary",
        "analytics",
        "sprints",
        "notifications",
        "settings",
        "roadmaps",
      ])
      .describe("The destination to navigate to"),
  }),
  execute: async ({ destination }: { destination: string }) => {
    return {
      route: destination,
    };
  },
});
