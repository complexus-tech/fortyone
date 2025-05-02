"use server";

import { auth } from "@/auth";
import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const deleteStoryAttachmentAction = async (
  storyId: string,
  attachmentId: string,
) => {
  try {
    const session = await auth();
    const res = await remove<ApiResponse<null>>(
      `stories/${storyId}/attachments/${attachmentId}`,
      session!,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
