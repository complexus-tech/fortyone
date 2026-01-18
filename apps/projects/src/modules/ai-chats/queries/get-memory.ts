import { get, type WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Memory } from "../types";

export const getMemories = async (ctx: WorkspaceCtx) => {
  const memories = await get<ApiResponse<Memory[]>>("users/memory", ctx);
  return memories.data!;
};
