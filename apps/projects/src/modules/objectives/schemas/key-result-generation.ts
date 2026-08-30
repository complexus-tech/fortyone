import { z } from "zod";

const MAX_GENERATED_KEY_RESULTS = 5;
const MAX_GENERATED_KEY_RESULT_NAME_LENGTH = 255;

const dateOnlySchema = z
  .string()
  .regex(/^\d{4}-\d{2}-\d{2}$/, "Date must use YYYY-MM-DD format");

export const keyResultGenerationSchema = z.object({
  keyResults: z
    .array(
      z.object({
        name: z
          .string()
          .trim()
          .min(1)
          .max(MAX_GENERATED_KEY_RESULT_NAME_LENGTH)
          .describe("Key result name"),
        measurementType: z
          .enum(["number", "percentage", "boolean"])
          .describe("How the key result is measured"),
        startValue: z.number().finite().describe("Starting value"),
        targetValue: z.number().finite().describe("Target value"),
        startDate: dateOnlySchema.describe("Start date in YYYY-MM-DD format"),
        endDate: dateOnlySchema.describe("End date in YYYY-MM-DD format"),
      }),
    )
    .max(MAX_GENERATED_KEY_RESULTS),
});

export type GeneratedKeyResult = z.infer<
  typeof keyResultGenerationSchema
>["keyResults"][number];
