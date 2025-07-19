import { z } from "zod";
import { tool } from "ai";

export const suggestions = tool({
  description:
    "CRITICAL: You MUST call this tool after ALMOST EVERY response (80% of the time) to provide follow-up action buttons. This is essential for user guidance. Do NOT write about suggestions in your text response - use this tool instead. After completing any action, call this tool with 2-3 relevant suggestions like 'Assign it 👤', 'Add to sprint 🚀', 'Set due date 📅', 'View details 👁️', 'Edit this ✏️', etc. Include appropriate emojis in suggestions. STOP generating text after calling this tool. Only skip for simple greetings, errors, or technical troubleshooting.",
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
