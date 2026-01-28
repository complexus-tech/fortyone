"use server";

import type { ApiResponse, User } from "@/types";
import ky from "ky";
import { getApiUrl } from "@/lib/api-url";
import { auth } from "@/auth";

const apiUrl = getApiUrl();

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
