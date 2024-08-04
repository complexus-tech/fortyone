import { DURATION_FROM_SECONDS } from "@/constants/time";
import { get } from "@/lib/http";
import { storyTags } from "@/modules/stories/constants";
import { Story } from "@/modules/stories/types";

export const getMyStories = async () => {
  const stories = await get<Story[]>("my-stories", {
    next: {
      revalidate: DURATION_FROM_SECONDS.MINUTE * 5,
      tags: [storyTags.mine()],
    },
  });
  return stories;
};
