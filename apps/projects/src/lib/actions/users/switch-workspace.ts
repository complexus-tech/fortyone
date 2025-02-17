"use server";
import ky from "ky";
import { auth } from "@/auth";
import type { ApiResponse, User } from "@/types";

const apiUrl = process.env.NEXT_PUBLIC_API_URL;

export async function switchWorkspace(workspaceId: string) {
  const session = await auth();
  const res = await ky.post(`${apiUrl}/workspaces/switch`, {
    json: {
      workspaceId,
    },
    headers: {
      Authorization: `Bearer ${session?.token}`,
    },
  });
  const user = await res.json<ApiResponse<User>>();

  return user.data?.lastUsedWorkspaceId;
}
