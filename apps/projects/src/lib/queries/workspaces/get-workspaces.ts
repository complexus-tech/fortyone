import ky from "ky";
import type { ApiResponse, Workspace } from "@/types";

const apiURL = process.env.NEXT_PUBLIC_API_URL;

export const getWorkspaces = async (token: string) => {
  const workspaces = await ky
    .get(`${apiURL}/workspaces`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    .json<ApiResponse<Workspace[]>>();

  return workspaces.data!;
};
