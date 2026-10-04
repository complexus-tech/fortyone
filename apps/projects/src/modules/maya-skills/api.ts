import type { z } from "zod";
import type { WorkspaceCtx } from "@/lib/http";
import { get, post, put, remove } from "@/lib/http";
import { mayaSkillInputSchema, mayaSkillSchema } from "./types";
import type { MayaSkillInput, MayaSkillUpdate } from "./types";

const authenticatedContext = (ctx: WorkspaceCtx) => {
  if (!ctx.session || !ctx.workspaceSlug) {
    throw new Error("Sign in to manage Maya skills.");
  }
  return ctx;
};

const decodeData =
  <T>(schema: z.ZodType<T>) =>
  (response: unknown): T => {
    if (!response || typeof response !== "object" || !("data" in response)) {
      throw new Error("Missing Maya skill data");
    }
    return schema.parse(response.data);
  };

export const listMayaSkills = (ctx: WorkspaceCtx) =>
  get(
    "maya/skills",
    authenticatedContext(ctx),
    undefined,
    decodeData(mayaSkillSchema.array()),
  );

export const createMayaSkill = (input: MayaSkillInput, ctx: WorkspaceCtx) =>
  post(
    "maya/skills",
    mayaSkillInputSchema.parse(input),
    authenticatedContext(ctx),
    undefined,
    decodeData(mayaSkillSchema),
  );

export const updateMayaSkill = (
  id: string,
  input: MayaSkillUpdate,
  ctx: WorkspaceCtx,
) =>
  put(
    `maya/skills/${encodeURIComponent(id)}`,
    { ...mayaSkillInputSchema.parse(input), updatedAt: input.updatedAt },
    authenticatedContext(ctx),
    undefined,
    decodeData(mayaSkillSchema),
  );

export const deleteMayaSkill = (id: string, ctx: WorkspaceCtx) =>
  remove(`maya/skills/${encodeURIComponent(id)}`, authenticatedContext(ctx));
