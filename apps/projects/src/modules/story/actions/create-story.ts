"use server";

import { post } from "@/lib/http";
import { DetailedStory, NewStory } from "../types";
import { revalidateTag } from "next/cache";
import { TAGS } from "@/constants/tags";

export const createStoryAction = async (newStory: NewStory) => {
  const story = await post<NewStory, DetailedStory>(`/stories`, newStory);
  revalidateTag(TAGS.stories);
  return story;
};
