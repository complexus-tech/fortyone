"use server";

import { post } from "@/lib/http";
import { DetailedStory, NewStory } from "../types";
import { revalidateTag } from "next/cache";
import { storyTags } from "@/modules/stories/constants";

export const createStoryAction = async (newStory: NewStory) => {
  const story = await post<NewStory, DetailedStory>("stories", newStory);
  revalidateTag(storyTags.mine());
  revalidateTag(storyTags.teams());
  if (newStory.parentId) {
    revalidateTag(storyTags.detail(newStory.parentId));
  }
  return story;
};
