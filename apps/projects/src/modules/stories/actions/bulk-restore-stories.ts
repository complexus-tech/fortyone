"use server";

import { TAGS } from "@/constants/tags";
import { post } from "@/lib/http";
import { revalidateTag } from "next/cache";

type Payload = {
  storyIds: string[];
};

export const bulkRestoreAction = async (storyIds: string[]) => {
  const payload = { storyIds };
  const stories = await post<Payload, Payload>("/stories/restore", payload);
  revalidateTag(TAGS.stories);
  storyIds.forEach((storyId) => revalidateTag(`${TAGS.stories}:${storyId}`));

  // trackEvent(LogEvents.storyCreated, {
  //   storyId: story.id,
  //   storyName: story.name,
  // });

  return stories;
};
