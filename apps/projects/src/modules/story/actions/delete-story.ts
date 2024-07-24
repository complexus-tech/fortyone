"use server";

import { remove } from "@/lib/http";
import { revalidateTag } from "next/cache";
import { TAGS } from "@/constants/tags";

export const deleteStoryAction = async (storyId: string) => {
  const _ = await remove(`/stories/${storyId}`);
  revalidateTag(TAGS.stories);
  revalidateTag(`${TAGS.stories}:${storyId}`);
  return { success: true };
};
