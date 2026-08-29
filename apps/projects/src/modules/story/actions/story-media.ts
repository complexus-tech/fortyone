import { auth } from "@/auth";
import { getApiUrl } from "@/lib/api-url";
import { post, remove } from "@/lib/http";
import type { RichTextMedia } from "@/lib/tiptap/rich-text-media";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

const STORY_MEDIA_UPLOAD_TIMEOUT_MS = 5 * 60 * 1000;

const resolveStoryMediaUrl = (url: string) => {
  if (/^https?:\/\//i.test(url)) return url;
  return `${getApiUrl()}${url.startsWith("/") ? url : `/${url}`}`;
};

const storyMediaContext = async (workspaceSlug: string) => {
  const session = await auth();
  return { session: session!, workspaceSlug };
};

export const uploadStoryMediaAction = async (
  storyId: string,
  file: File,
  workspaceSlug: string,
) => {
  try {
    const formData = new FormData();
    formData.append("file", file);
    const response = await post<FormData, ApiResponse<RichTextMedia>>(
      `stories/${storyId}/media`,
      formData,
      await storyMediaContext(workspaceSlug),
      { timeout: STORY_MEDIA_UPLOAD_TIMEOUT_MS },
    );
    if (response.data) {
      response.data.url = resolveStoryMediaUrl(response.data.url);
    }
    return response;
  } catch (error) {
    return getApiError(error);
  }
};

export const deleteStoryMediaAction = async (
  storyId: string,
  attachmentId: string,
  workspaceSlug: string,
) => {
  try {
    return await remove<ApiResponse<null>>(
      `stories/${storyId}/media/${attachmentId}`,
      await storyMediaContext(workspaceSlug),
    );
  } catch (error) {
    return getApiError(error);
  }
};
