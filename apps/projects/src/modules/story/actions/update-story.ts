"use server";

import { patch } from "@/lib/http";
import { revalidateTag } from "next/cache";
import { TAGS } from "@/constants/tags";
import { DetailedStory } from "../types";

export const updateStoryAction = async (
  storyId: string,
  payload: Partial<DetailedStory>,
) => {
  const _ = await patch(`/stories/${storyId}`, payload);
  revalidateTag(TAGS.stories);
  revalidateTag(`${TAGS.stories}:${storyId}`);
  return { success: true };
};
