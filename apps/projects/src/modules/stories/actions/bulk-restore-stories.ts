"use server";

import { post } from "@/lib/http";
import { revalidateTag } from "next/cache";
import { storyTags } from "@/modules/stories/constants";

type Payload = {
  storyIds: string[];
};

export const bulkRestoreAction = async (storyIds: string[]) => {
  const stories = await post<Payload, Payload>("stories/restore", {
    storyIds,
  });
  storyIds.forEach((storyId) => {
    revalidateTag(storyTags.detail(storyId));
  });
  revalidateTag(storyTags.teams());

  return stories;
};
