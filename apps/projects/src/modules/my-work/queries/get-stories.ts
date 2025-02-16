"use server";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import { get } from "@/lib/http";
import { storyTags } from "@/modules/stories/constants";
import type { Story } from "@/modules/stories/types";
import type { ApiResponse } from "@/types";

export const getMyStories = async () => {
  const stories = await get<ApiResponse<Story[]>>("my-stories", {
    next: {
      revalidate: DURATION_FROM_SECONDS.MINUTE * 5,
      tags: [storyTags.mine()],
    },
  });
  return stories.data!;
};
