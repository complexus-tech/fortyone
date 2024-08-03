"use server";

import { post } from "@/lib/http";
import { DetailedStory, NewStory } from "../types";

export const createStoryAction = async (newStory: NewStory) => {
  const story = await post<NewStory, DetailedStory>(`/stories`, newStory);
  return story;
};
