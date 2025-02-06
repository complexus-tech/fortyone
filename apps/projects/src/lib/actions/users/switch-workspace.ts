"use server";
import type { ApiResponse, User } from "@/types";
import { post } from "@/lib/http";

export async function switchWorkspace(workspaceId: string) {
  const res = await post<
    {
      workspaceId: string;
    },
    ApiResponse<User>
  >("switch", {
    workspaceId,
  });

  return res.data?.lastUsedWorkspaceId;
}
