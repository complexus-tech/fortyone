import { z } from "zod";

export const figmaDescriptionRequestSchema = z.object({
  fileName: z.string().trim().min(1).max(500),
  nodeName: z.string().trim().min(1).max(500).nullable(),
  nodeType: z.string().trim().max(100).nullable(),
  textContent: z.array(z.string().trim().min(1).max(500)).min(1).max(24),
});

export const figmaDescriptionSchema = z.object({
  overview: z.string().trim().min(1).max(1_000),
  requirements: z.array(z.string().trim().min(1).max(300)).max(8),
  acceptanceCriteria: z.array(z.string().trim().min(1).max(300)).max(8),
  implementationNotes: z.array(z.string().trim().min(1).max(300)).max(5),
});

export type FigmaDescription = z.infer<typeof figmaDescriptionSchema>;
export type FigmaDescriptionRequest = z.infer<
  typeof figmaDescriptionRequestSchema
>;
