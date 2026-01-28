import type { ApiResponse, Workspace } from "@/types";
import ky from "ky";
import { getApiUrl } from "@/lib/api-url";

const apiURL = getApiUrl();

export const getWorkspaces = async (token: string) => {
  const workspaces = await ky
    .get(`${apiURL}/workspaces`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    .json<ApiResponse<Workspace[]>>();

  return workspaces.data!;
};
