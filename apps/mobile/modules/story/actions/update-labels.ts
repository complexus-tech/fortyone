import { put } from "@/lib/http/fetch";
import type { ApiResponse } from "@/types";

export const updateLabelsAction = async (storyId: string, labels: string[]) => {
  const response = await put<{ labels: string[] }, ApiResponse<null>>(
    `stories/${storyId}/labels`,
    { labels }
  );
  return response;
};
