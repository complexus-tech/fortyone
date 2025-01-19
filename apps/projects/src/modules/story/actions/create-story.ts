"use server";

import { revalidateTag } from "next/cache";
import { post } from "@/lib/http";
import { storyTags } from "@/modules/stories/constants";
import type { ApiResponse } from "@/types";
import type { DetailedStory, NewStory } from "../types";

export const createStoryAction = async (newStory: NewStory) => {
  const story = await post<NewStory, ApiResponse<DetailedStory>>(
    "stories",
    newStory,
  );
  revalidateTag(storyTags.mine());
  revalidateTag(storyTags.teams());
  revalidateTag(storyTags.objectives());
  revalidateTag(storyTags.sprints());
  if (newStory.parentId) {
    revalidateTag(storyTags.detail(newStory.parentId));
  }
  return story.data!;
};
