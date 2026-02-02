import type { ApiResponse, Workspace } from "@/types";
import ky from "ky";
import { getApiUrl } from "@/lib/api-url";

const apiURL = getApiUrl();

export async function getWorkspaces(token: string): Promise<Workspace[]> {
  const res = await ky
    .get(`${apiURL}/workspaces`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .json<ApiResponse<Workspace[]>>();

  if (res.error) {
    throw new Error(res.error.message);
  }

  return res.data ?? [];
}
