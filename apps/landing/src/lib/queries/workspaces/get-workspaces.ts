import ky from "ky";
import type { ApiResponse, Workspace } from "@/types";

const apiURL = process.env.NEXT_PUBLIC_API_URL;

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
