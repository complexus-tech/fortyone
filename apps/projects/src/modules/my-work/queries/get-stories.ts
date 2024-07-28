import { get } from "@/lib/http";
import { Story } from "@/modules/stories/types";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import { TAGS } from "@/constants/tags";

export const getMyStories = async () => {
  const stories = await get<Story[]>(`/my-stories`, {
    next: {
      revalidate: DURATION_FROM_SECONDS.MINUTE * 30,
      tags: [TAGS.stories],
    },
  });
  return stories;
};
