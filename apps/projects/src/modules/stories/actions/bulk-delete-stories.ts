"use server";

import { TAGS } from "@/constants/tags";
import { remove } from "@/lib/http";
import { revalidateTag } from "next/cache";

type Payload = {
  storyIds: string[];
};

export const bulkDeleteAction = async (storyIds: string[]) => {
  const stories = await remove<Payload, Payload>("/stories", {
    body: JSON.stringify({ storyIds }),
  });
  revalidateTag(TAGS.stories);
  storyIds.forEach((storyId) => revalidateTag(`${TAGS.stories}:${storyId}`));

  // trackEvent(LogEvents.storyCreated, {
  //   storyId: story.id,
  //   storyName: story.name,
  // });

  return stories;
};
