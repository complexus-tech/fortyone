"use server";

import { revalidateTag } from "next/cache";
import { put } from "@/lib/http";
import { storyTags } from "@/modules/stories/constants";
import type { DetailedStory } from "../types";

export const updateStoryAction = async (
  storyId: string,
  payload: Partial<DetailedStory>,
) => {
  const _ = await put(`stories/${storyId}`, payload);
  revalidateTag(storyTags.detail(storyId));
  revalidateTag(storyTags.mine());
  revalidateTag(storyTags.teams());
  revalidateTag(storyTags.objectives());
  revalidateTag(storyTags.sprints());

  return storyId;
};
