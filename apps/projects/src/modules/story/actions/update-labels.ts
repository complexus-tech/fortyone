"use server";

import { revalidateTag } from "next/cache";
import { put } from "@/lib/http";
import { storyTags } from "@/modules/stories/constants";
import { labelTags } from "@/constants/keys";

export const updateLabelsAction = async (storyId: string, labels: string[]) => {
  const _ = await put(`stories/${storyId}/labels`, { labels });
  revalidateTag(storyTags.detail(storyId));
  revalidateTag(storyTags.mine());
  revalidateTag(storyTags.teams());
  revalidateTag(labelTags.lists());
  revalidateTag(storyTags.objectives());
  revalidateTag(storyTags.sprints());

  return storyId;
};
