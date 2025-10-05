import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Workspace } from "@/types/workspace";

export const getWorkspaces = async () => {
  const response = await get<ApiResponse<Workspace[]>>("workspaces", {
    useWorkspace: false,
  });
  return response.data!;
};
