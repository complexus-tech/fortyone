import type { WorkspaceCtx } from "@/lib/http";
import { get, post, put, remove } from "@/lib/http";
import { presetSchema, presetPageSchema } from "./schemas";
import type { Preset, PresetInput, PresetKind, PresetPage } from "./types";

const decodeData = (response: unknown): Preset => {
  if (!response || typeof response !== "object" || !("data" in response))
    throw new Error("Missing preset data");
  return presetSchema.parse(response.data);
};
const decodePage = (response: unknown): PresetPage => {
  if (!response || typeof response !== "object" || !("data" in response))
    throw new Error("Missing preset data");
  return presetPageSchema.parse(response.data);
};

export const listPresets = (
  teamId: string,
  kind: PresetKind,
  cursor: string,
  ctx: WorkspaceCtx,
) => {
  const query = new URLSearchParams({
    teamId,
    kind,
    limit: "50",
    ...(cursor ? { cursor } : {}),
  });
  return get(`work-presets?${query.toString()}`, ctx, undefined, decodePage);
};
export const createPreset = (input: PresetInput, ctx: WorkspaceCtx) =>
  post("work-presets", input, ctx, undefined, decodeData);
export const renamePreset = (id: string, name: string, ctx: WorkspaceCtx) =>
  put(`work-presets/${id}`, { name }, ctx, undefined, decodeData);
export const archivePreset = (id: string, ctx: WorkspaceCtx) =>
  remove(`work-presets/${id}`, ctx);
