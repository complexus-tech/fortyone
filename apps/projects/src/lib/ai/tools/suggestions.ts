import { z } from "zod";
import { tool } from "ai";

export const suggestions = tool({
  description:
    "Return follow-up suggestions after a user action (e.g., after creating a story, updating a sprint, viewing a story, etc.)",
  parameters: z.object({
    suggestions: z
      .array(z.string())
      .describe("The suggestions to show the user max 3"),
  }),
  execute: async ({ suggestions }: { suggestions: string[] }) => {
    return {
      suggestions,
    };
  },
});
