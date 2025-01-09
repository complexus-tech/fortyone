"use server";

import { revalidateTag } from "next/cache";
import { post } from "@/lib/http";
import { storyTags } from "@/modules/stories/constants";
import type { ApiResponse } from "@/types";

type Payload = {
  storyIds: string[];
};

export const bulkRestoreAction = async (storyIds: string[]) => {
  const stories = await post<Payload, ApiResponse<Payload>>("stories/restore", {
    storyIds,
  });
  storyIds.forEach((storyId) => {
    revalidateTag(storyTags.detail(storyId));
  });
  revalidateTag(storyTags.teams());
  revalidateTag(storyTags.mine());

  return stories.data;
};
