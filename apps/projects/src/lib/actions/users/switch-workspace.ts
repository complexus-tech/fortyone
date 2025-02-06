"use server";
import ky from "ky";
import type { ApiResponse, User } from "@/types";
import { auth } from "@/auth";

const apiURL = process.env.NEXT_PUBLIC_API_URL;

export async function switchWorkspace(workspaceId: string) {
  const session = await auth();
  if (!session) {
    throw new Error("Unauthorized");
  }
  const res = await ky
    .post(`${apiURL}/users/switch-workspace`, {
      json: {
        workspaceId,
      },
      headers: {
        Authorization: `Bearer ${session.token}`,
      },
    })
    .json<ApiResponse<User>>();

  return res.data?.lastUsedWorkspaceId;
}
