import ky from "ky";
import type { ApiResponse } from "@/types";
import type { Team } from "@/modules/teams/types";
import { auth } from "@/auth";

const apiURL = process.env.NEXT_PUBLIC_API_URL;

export const getTeams = async (workspace: string): Promise<Team[]> => {
  if (!workspace) {
    return [];
  }
  const session = await auth();
  const res = await ky
    .get(`${apiURL}/workspaces/${workspace}/teams`, {
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
