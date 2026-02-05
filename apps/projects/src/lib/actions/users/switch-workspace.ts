"use server";

import type { ApiResponse, User } from "@/types";
import ky from "ky";
import { auth } from "@/auth";
import { getApiUrl } from "@/lib/api-url";
import { buildAuthHeaders } from "@/lib/http/auth-headers";
import { getCookieHeader } from "@/lib/http/header";

const apiUrl = getApiUrl();

export async function switchWorkspace(workspaceId: string) {
  const session = await auth();
  const cookieHeader = await getCookieHeader();
  const headers = buildAuthHeaders({ token: session?.token, cookieHeader });
  const res = await ky.post(`${apiUrl}/workspaces/switch`, {
    json: {
      workspaceId,
    },
    credentials: "include",
    headers,
  });
  const user = await res.json<ApiResponse<User>>();

  return user.data?.lastUsedWorkspaceId;
}
