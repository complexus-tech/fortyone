"use server";

import ky from "ky";
import type { ApiResponse, Workspace } from "@/types";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import { workspaceTags } from "@/constants/keys";

const apiURL = process.env.NEXT_PUBLIC_API_URL;

export const getWorkspaces = async (token: string) => {
  const workspaces = await ky
    .get(`${apiURL}/workspaces`, {
      headers: { Authorization: `Bearer ${token}` },
      next: {
        revalidate: DURATION_FROM_SECONDS.MINUTE * 10,
        tags: [workspaceTags.lists()],
      },
    })
    .json<ApiResponse<Workspace[]>>();
  return workspaces.data!;
};
