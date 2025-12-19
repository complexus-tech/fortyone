"use server";

import { post } from "@/lib/http/fetch";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { StoryAssociationType } from "../types";

export const addAssociationAction = async (
  fromStoryId: string,
  payload: { toStoryId: string; type: StoryAssociationType },
) => {
  try {
    const session = await auth();
    const res = await post<
      { toStoryId: string; type: StoryAssociationType },
      ApiResponse<null>
    >(`stories/${fromStoryId}/associations`, payload, session!);
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
