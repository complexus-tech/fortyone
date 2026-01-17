"use server";

import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { StoryAssociationType } from "../types";

export const addAssociationAction = async (
  fromStoryId: string,
  payload: { toStoryId: string; type: StoryAssociationType },
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const res = await post<
      { toStoryId: string; type: StoryAssociationType },
      ApiResponse<null>
    >(`stories/${fromStoryId}/associations`, payload, ctx);
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
