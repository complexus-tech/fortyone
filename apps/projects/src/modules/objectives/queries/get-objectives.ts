import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Objective } from "../types";

export const getObjectives = async (ctx: WorkspaceCtx, search = "") => {
  const params = new URLSearchParams();
  if (search.trim()) {
    params.set("search", search.trim());
  }
  const query = params.toString();
  const objectives = await get<ApiResponse<Objective[]>>(
    `objectives${query ? `?${query}` : ""}`,
    ctx,
  );
  return objectives.data ?? [];
};
