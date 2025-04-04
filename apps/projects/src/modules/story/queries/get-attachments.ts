import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { StoryAttachment } from "../types";

export const getStoryAttachments = async (id: string) => {
  const story = await get<ApiResponse<StoryAttachment[]>>(
    `stories/${id}/attachments`,
  );
  return story.data!;
};
