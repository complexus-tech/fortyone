import ky from "ky";
import type { ApiResponse, Team } from "@/types";
import { auth } from "@/auth";

const apiURL = process.env.NEXT_PUBLIC_API_URL;

export const getTeams = async (workspaceId: string): Promise<Team[]> => {
  if (!workspaceId) {
    return [];
  }
  const session = await auth();
  const res = await ky
    .get(`${apiURL}/workspaces/${workspaceId}/teams`, {
      headers: {
        Authorization: `Bearer ${session?.token}`,
      },
    })
    .json<ApiResponse<Team[]>>();

  if (res.error) {
    return [];
  }

  return res.data ?? [];
};
