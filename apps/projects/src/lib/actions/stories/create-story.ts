"use server";

import { TAGS } from "@/constants/tags";
import { post } from "@/lib/http";
import { trackEvent } from "@/lib/tracking";
import { LogEvents } from "@/lib/tracking/events";
import { revalidateTag } from "next/cache";

export const createStoryAction = async (story: any) => {
  const s = await post("/stories", story);
  revalidateTag(TAGS.stories);

  trackEvent(LogEvents.storyCreated, {
    storyId: story.id,
    storyName: story.name,
  });

  return story;
};
