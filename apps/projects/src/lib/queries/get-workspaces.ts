import type { ApiResponse, Workspace } from "@/types";
import { get } from "api-client";

export async function getWorkspaces(
  _token?: string,
  _cookieHeader?: string,
): Promise<Workspace[]> {
  const res = await get<ApiResponse<Workspace[]>>("workspaces");

  if (res.error) {
    throw new Error(res.error.message);
  }

  return res.data ?? [];
}
