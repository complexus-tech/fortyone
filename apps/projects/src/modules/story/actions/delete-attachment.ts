"use server";

import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const deleteStoryAttachmentAction = async (
  storyId: string,
  attachmentId: string,
) => {
  try {
    const res = await remove<ApiResponse<null>>(
      `stories/${storyId}/attachments/${attachmentId}`,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
