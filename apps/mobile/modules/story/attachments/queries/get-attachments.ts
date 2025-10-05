import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { StoryAttachment } from "@/types/attachment";

export const getStoryAttachments = async (storyId: string) => {
  const response = await get<ApiResponse<StoryAttachment[]>>(
    `stories/${storyId}/attachments`
  );
  return response.data!;
};
