import type { ApiResponse } from "@/types";
import type { Team } from "@/modules/teams/types";
import ky from "ky";
import { auth } from "@/auth";
import { getApiUrl } from "@/lib/api-url";
import { buildAuthHeaders } from "@/lib/http/auth-headers";
import { getCookieHeader } from "@/lib/http/header";

const apiURL = getApiUrl();

export const getTeams = async (workspace: string): Promise<Team[]> => {
  if (!workspace) {
    return [];
  }
  const session = await auth();
  const cookieHeader = await getCookieHeader();
  const headers = buildAuthHeaders({ token: session?.token, cookieHeader });
  const res = await ky
    .get(`${apiURL}/workspaces/${workspace}/teams`, {
      credentials: "include",
      headers,
    })
    .json<ApiResponse<Team[]>>();

  if (res.error) {
    return [];
  }

  return res.data ?? [];
};
