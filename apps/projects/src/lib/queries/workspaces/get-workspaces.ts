"use server";
import ky from "ky";
import { workspaceTags } from "@/constants/keys";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import type { ApiResponse, Workspace } from "@/types";
import { auth } from "@/auth";

export const getWorkspaces = async (): Promise<Workspace[]> => {
  // this is directly using ky because we are not injecting the workspace id.
  const session = await auth();
  const workspaces = await ky
    .get("workspaces", {
      next: {
        revalidate: DURATION_FROM_SECONDS.MINUTE * 30,
        tags: [workspaceTags.lists()],
      },
      headers: {
        Authorization: `Bearer ${session?.token}`,
      },
    })
    .json<ApiResponse<Workspace[]>>();
  return workspaces.data!;
};
