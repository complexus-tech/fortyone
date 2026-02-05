import type { ApiResponse, Workspace } from "@/types";
import ky from "ky";
import { buildAuthHeaders } from "@/lib/http/auth-headers";
import { getApiUrl } from "@/lib/api-url";

const apiURL = getApiUrl();

export const getWorkspaces = async (token?: string, cookieHeader?: string) => {
  const headers = buildAuthHeaders({ token, cookieHeader });
  const workspaces = await ky
    .get(`${apiURL}/workspaces`, {
      credentials: "include",
      headers,
    })
    .json<ApiResponse<Workspace[]>>();

  return workspaces.data!;
};
