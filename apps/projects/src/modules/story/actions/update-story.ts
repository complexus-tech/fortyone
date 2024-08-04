"use server";

import { patch } from "@/lib/http";
import { revalidateTag } from "next/cache";
import { DetailedStory } from "../types";
import { storyTags } from "@/modules/stories/constants";

export const updateStoryAction = async (
  storyId: string,
  payload: Partial<DetailedStory>,
) => {
  const _ = await patch(`stories/${storyId}`, payload);
  revalidateTag(storyTags.detail(storyId));
  revalidateTag(storyTags.mine());
  revalidateTag(storyTags.teams());
  return storyId;
};
