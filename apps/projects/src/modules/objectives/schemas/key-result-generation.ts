import { z } from "zod";

export const keyResultGenerationSchema = z.object({
  keyResults: z.array(
    z.object({
      name: z.string().describe("Key result name"),
      measurementType: z
        .enum(["number", "percentage", "boolean"])
        .describe("How the key result is measured"),
      startValue: z.number().describe("Starting value"),
      targetValue: z.number().describe("Target value"),
    }),
  ),
});
