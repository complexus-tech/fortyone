import { get } from "@/lib/http";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import { TAGS } from "@/constants/tags";
import { auth } from "@/auth";
import { DetailedStory } from "../types";

export const getStory = async (id: string) => {
  const session = await auth();

  const story = await get<DetailedStory>(`/stories/${id}`, {
    next: {
      revalidate: DURATION_FROM_SECONDS.MINUTE * 10,
      tags: [`${TAGS.stories}:${id}`],
    },
  });
  // track story view
  return story;
};
