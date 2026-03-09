import type { ApiResponse, Workspace } from "@/types";
import { get } from "api-client";

export const getWorkspaces = async (
  _token?: string,
  _cookieHeader?: string,
) => {
  const workspaces = await get<ApiResponse<Workspace[]>>("workspaces");

  return workspaces.data!;
};
