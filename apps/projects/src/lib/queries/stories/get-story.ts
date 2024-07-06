"use server";
import "server-only";

import { get } from "@/lib/http";

export const getStory = async (storyId: string) => {
  const story = await get(`/stories/${storyId}`, {
    next: {
      revalidate: 3600,
    },
  });
  return story;
};
