import { z } from "zod";

const MAX_GENERATED_SUBSTORIES = 5;
const MAX_GENERATED_SUBSTORY_TITLE_LENGTH = 255;

export const substoryGenerationSchema = z.object({
  substories: z
    .array(
      z.object({
        title: z
          .string()
          .trim()
          .min(1)
          .max(MAX_GENERATED_SUBSTORY_TITLE_LENGTH)
          .describe("Clear, actionable substory title"),
      }),
    )
    .max(MAX_GENERATED_SUBSTORIES),
});
