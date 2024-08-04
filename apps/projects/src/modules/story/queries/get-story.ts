import { get } from "@/lib/http";
import { DetailedStory } from "../types";
import { storyTags } from "@/modules/stories/constants";
import { DURATION_FROM_SECONDS } from "@/constants/time";

export const getStory = async (id: string) => {
  const story = await get<DetailedStory>(`stories/${id}`, {
    next: {
      revalidate: DURATION_FROM_SECONDS.MINUTE * 5,
      tags: [storyTags.detail(id)],
    },
  });
  return story;
};
