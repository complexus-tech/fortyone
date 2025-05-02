"use server";

import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { StoryAttachment } from "../types";

export const addAttachmentAction = async (storyId: string, file: File) => {
  try {
    const formData = new FormData();
    formData.append("file", file);
    const session = await auth();
    const res = await post<FormData, ApiResponse<StoryAttachment>>(
      `stories/${storyId}/attachments`,
      formData,
      session!,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
