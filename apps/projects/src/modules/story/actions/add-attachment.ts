"use server";

import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type { StoryAttachment } from "../types";

export const addAttachmentAction = async (storyId: string, file: File) => {
  try {
    const formData = new FormData();
    formData.append("file", file);
    const res = await post<FormData, ApiResponse<StoryAttachment>>(
      `stories/${storyId}/attachments`,
      formData,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
