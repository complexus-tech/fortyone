import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { DetailedStory, StoryPriority } from "@/modules/stories/types";

export type CreateStoryPayload = {
  title: string;
  description?: string;
  descriptionHTML?: string;
  teamId: string;
  statusId?: string | null;
  assigneeId?: string | null;
  priority: StoryPriority;
  labelIds?: string[];
};

export const createStory = async (payload: CreateStoryPayload) => {
  const response = await post<CreateStoryPayload, ApiResponse<DetailedStory>>(
    "stories",
    payload
  );

  return response;
};
