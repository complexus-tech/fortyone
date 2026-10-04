import { z } from "zod";

export const mayaSkillInputSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, "Give this skill a name.")
    .refine(
      (value) => Array.from(value).length <= 80,
      "Keep the name to 80 characters or fewer.",
    ),
  description: z
    .string()
    .trim()
    .refine(
      (value) => Array.from(value).length <= 300,
      "Keep the description to 300 characters or fewer.",
    ),
  instructions: z
    .string()
    .trim()
    .min(1, "Add instructions for Maya.")
    .refine(
      (value) => Array.from(value).length <= 12_000,
      "Keep instructions to 12,000 characters or fewer.",
    ),
});

export const mayaSkillSchema = mayaSkillInputSchema.extend({
  id: z.string().uuid(),
  createdAt: z.string(),
  updatedAt: z.string(),
});

export type MayaSkill = z.infer<typeof mayaSkillSchema>;
export type MayaSkillInput = z.infer<typeof mayaSkillInputSchema>;
export type MayaSkillUpdate = MayaSkillInput & { updatedAt: string };
