"use server";

import { revalidateTag } from "next/cache";
import { patch } from "@/lib/http";
import { storyTags } from "@/modules/stories/constants";
import type { DetailedStory } from "../types";

export const updateStoryAction = async (
  storyId: string,
  payload: Partial<DetailedStory>,
) => {
  const _ = await patch(`stories/${storyId}`, payload);
  revalidateTag(storyTags.detail(storyId));
  revalidateTag(storyTags.mine());
  revalidateTag(storyTags.teams());
  revalidateTag(storyTags.objectives());
  revalidateTag(storyTags.sprints());

  return storyId;
};
