"use server";
import { get } from "@/lib/http";
import { workspaceTags } from "@/constants/keys";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import type { ApiResponse, Workspace } from "@/types";

export const getWorkspace = async (id: string): Promise<Workspace> => {
  const workspace = await get<ApiResponse<Workspace>>(id, {
    next: {
      revalidate: DURATION_FROM_SECONDS.MINUTE * 30,
      tags: [workspaceTags.detail(id)],
    },
  });
  return workspace.data!;
};
