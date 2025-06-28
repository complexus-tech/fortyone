import { z } from "zod";
import { tool } from "ai";

export const theme = tool({
  description: "Change the application theme or appearance settings",
  parameters: z.object({
    theme: z
      .enum(["light", "dark", "system", "toggle"])
      .describe(
        "The theme to set: light, dark, system, or toggle between light/dark",
      ),
  }),
  execute: async ({ theme }: { theme: string }) => {
    return {
      theme,
    };
  },
});
