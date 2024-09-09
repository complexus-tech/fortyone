"use server";

import { remove } from "@/lib/http";
import { revalidateTag } from "next/cache";
import { storyTags } from "@/modules/stories/constants";

export const deleteStoryAction = async (storyId: string) => {
  const _ = await remove(`stories/${storyId}`);
  revalidateTag(storyTags.detail(storyId));
  revalidateTag(storyTags.mine());
  revalidateTag(storyTags.teams());
  return storyId;
};
