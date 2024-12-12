"use server";

import { remove } from "@/lib/http";
import { revalidateTag } from "next/cache";
import { storyTags } from "../constants";
import { ApiResponse } from "@/types";

type Payload = {
  storyIds: string[];
};

export const bulkDeleteAction = async (storyIds: string[]) => {
  const stories = await remove<ApiResponse<Payload>>("stories", {
    json: { storyIds },
  });
  storyIds.forEach((storyId) => {
    revalidateTag(storyTags.detail(storyId));
  });
  revalidateTag(storyTags.mine());
  revalidateTag(storyTags.teams());

  return stories.data;
};
