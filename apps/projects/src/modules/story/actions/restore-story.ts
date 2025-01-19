"use server";

import { revalidateTag } from "next/cache";
import { post } from "@/lib/http";
import { storyTags } from "@/modules/stories/constants";

export const restoreStoryAction = async (storyId: string) => {
  const _ = await post(`stories/${storyId}/restore`, {});
  revalidateTag(storyTags.detail(storyId));
  revalidateTag(storyTags.mine());
  revalidateTag(storyTags.teams());
  revalidateTag(storyTags.objectives());
  revalidateTag(storyTags.sprints());
  return storyId;
};
