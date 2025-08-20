"use server";

import { auth } from "@/auth";
import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

type Payload = {
  storyIds: string[];
  hardDelete?: boolean;
};

export const bulkDeleteAction = async ({ storyIds, hardDelete }: Payload) => {
  try {
    const session = await auth();
    const stories = await remove<ApiResponse<Payload>>("stories", session!, {
      json: { storyIds, hardDelete },
    });
    return stories;
  } catch (error) {
    return getApiError(error);
  }
};
