import { z } from "zod";
import { tool } from "ai";

export const suggestions = tool({
  description:
    "Provide follow-up action buttons after completing user requests. Use this tool with 2-3 relevant suggestions like 'Assign it', 'Add to sprint', 'View details'. Stop generating text after calling this tool.",
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
