import { z } from "zod";

const dateSchema = z.string().nullable();
const namedItemSchema = z.object({
  id: z.string().max(100),
  name: z.string().trim().min(1).max(500),
});

export const okrQualityAssessmentSchema = z.object({
  verdict: z.enum(["strong", "needs_attention", "duplicate"]),
  headline: z.string().trim().min(1).max(120),
  guidance: z.array(z.string().trim().min(1).max(180)).max(3),
  suggestedName: z.string().trim().max(500).nullable(),
  duplicateOf: z.string().trim().max(500).nullable(),
});

export const okrQualityRequestSchema = z.discriminatedUnion("kind", [
  z.object({
    kind: z.literal("objective"),
    draft: z.object({
      name: z.string().trim().min(1).max(500),
      summary: z.string().trim().max(500),
      startDate: dateSchema,
      endDate: dateSchema,
    }),
    existingObjectives: z.array(namedItemSchema).max(100),
  }),
  z.object({
    kind: z.literal("key_result"),
    draft: z.object({
      name: z.string().trim().min(1).max(500),
      measurementType: z.enum(["number", "percentage", "boolean"]),
      startValue: z.number().finite(),
      targetValue: z.number().finite(),
      startDate: dateSchema,
      endDate: dateSchema,
    }),
    objective: z.object({
      id: z.string().max(100),
      name: z.string().trim().min(1).max(500),
      startDate: dateSchema,
      endDate: dateSchema,
    }),
    existingKeyResults: z.array(namedItemSchema).max(100),
  }),
]);

export type OkrQualityAssessment = z.infer<typeof okrQualityAssessmentSchema>;
export type OkrQualityRequest = z.infer<typeof okrQualityRequestSchema>;
