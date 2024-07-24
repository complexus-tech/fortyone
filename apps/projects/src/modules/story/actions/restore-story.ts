"use server";

import { post } from "@/lib/http";
import { revalidateTag } from "next/cache";
import { TAGS } from "@/constants/tags";

export const restoreStoryAction = async (storyId: string) => {
  const id = await post(`/stories/${storyId}/restore`);
  revalidateTag(TAGS.stories);
  revalidateTag(`${TAGS.stories}:${storyId}`);
  return id;
};
