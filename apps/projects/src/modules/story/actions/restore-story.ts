"use server";

import { post } from "@/lib/http";
import { revalidateTag } from "next/cache";
import { storyTags } from "@/modules/stories/constants";

export const restoreStoryAction = async (storyId: string) => {
  const _ = await post(`stories/${storyId}/restore`, {});
  revalidateTag(storyTags.detail(storyId));
  revalidateTag(storyTags.mine());
  revalidateTag(storyTags.teams());
  return storyId;
};
