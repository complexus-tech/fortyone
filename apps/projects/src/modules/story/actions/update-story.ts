"use server";

import { revalidateTag } from "next/cache";
import { put } from "@/lib/http";
import { storyTags } from "@/modules/stories/constants";
import { getApiError } from "@/utils";
import type { DetailedStory } from "../types";

export const updateStoryAction = async (
  storyId: string,
  payload: Partial<DetailedStory>,
) => {
  try {
    const _ = await put(`stories/${storyId}`, payload);
    revalidateTag(storyTags.detail(storyId));
    revalidateTag(storyTags.mine());
    revalidateTag(storyTags.teams());
    revalidateTag(storyTags.objectives());
    revalidateTag(storyTags.sprints());

    return storyId;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to update story");
  }
};
