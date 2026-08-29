import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { StoryAttachment } from "../types";

const ATTACHMENT_UPLOAD_TIMEOUT_MS = 5 * 60 * 1000;

export const addAttachmentAction = async (
  storyId: string,
  file: File,
  workspaceSlug: string,
) => {
  try {
    const formData = new FormData();
    formData.append("file", file);
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const res = await post<FormData, ApiResponse<StoryAttachment>>(
      `stories/${storyId}/attachments`,
      formData,
      ctx,
      { timeout: ATTACHMENT_UPLOAD_TIMEOUT_MS },
    );
    return res;
  } catch (error) {
    const response = getApiError(error);
    if (
      response.error?.message === "An error occurred" &&
      error instanceof Error &&
      error.message.trim()
    ) {
      return {
        data: null,
        error: {
          message: error.message,
        },
      };
    }
    return response;
  }
};
