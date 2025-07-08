import { z } from "zod";
import { tool } from "ai";

export const quickCreate = tool({
  description: "Open dialogs to create new items in the application",
  parameters: z.object({
    action: z
      .enum(["story", "objective", "sprint"])
      .describe("The type of item to create: story, objective, or sprint"),
  }),
  execute: async ({ action }: { action: string }) => {
    return {
      action,
    };
  },
});
