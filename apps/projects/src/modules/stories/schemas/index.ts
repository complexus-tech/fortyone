import { z } from "zod";

export const substoryGenerationSchema = z.object({
  substories: z.array(
    z.object({
      title: z.string().describe("Clear, actionable substory title"),
    }),
  ),
});
