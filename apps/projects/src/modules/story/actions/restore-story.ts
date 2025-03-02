"use server";

import { revalidateTag } from "next/cache";
import { post } from "@/lib/http";
import { storyTags } from "@/modules/stories/constants";
import { getApiError } from "@/utils";

export const restoreStoryAction = async (storyId: string) => {
  try {
    const _ = await post(`stories/${storyId}/restore`, {});
    revalidateTag(storyTags.detail(storyId));
    revalidateTag(storyTags.mine());
    revalidateTag(storyTags.teams());
    revalidateTag(storyTags.objectives());
    revalidateTag(storyTags.sprints());
    return storyId;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to restore story");
  }
};
