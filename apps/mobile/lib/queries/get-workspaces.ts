import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Workspace } from "@/types/workspace";

export const getWorkspaces = async (signal?: AbortSignal) => {
  const response = await get<ApiResponse<Workspace[]>>("workspaces", {
    useWorkspace: false,
    signal,
  });
  return response.data!;
};
