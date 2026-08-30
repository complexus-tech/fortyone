import { get } from "api-client";
import type { ApiResponse, Workspace } from "@/types";

export const getWorkspaces = async (
  _token?: string,
  _cookieHeader?: string,
) => {
  const workspaces = await get<ApiResponse<Workspace[]>>("workspaces");

  return workspaces.data!;
};
