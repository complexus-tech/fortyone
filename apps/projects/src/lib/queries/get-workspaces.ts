import type { ApiResponse, Workspace } from "@/types";
import ky from "ky";
import { buildAuthHeaders } from "@/lib/http/auth-headers";
import { getApiUrl } from "@/lib/api-url";

const apiURL = getApiUrl();

export async function getWorkspaces(
  token?: string,
  cookieHeader?: string,
): Promise<Workspace[]> {
  const headers = buildAuthHeaders({ token, cookieHeader });
  const res = await ky
    .get(`${apiURL}/workspaces`, {
      credentials: "include",
      headers,
    })
    .json<ApiResponse<Workspace[]>>();

  if (res.error) {
    throw new Error(res.error.message);
  }

  return res.data ?? [];
}
