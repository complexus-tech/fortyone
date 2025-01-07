"use server";
import { linkTags } from "@/constants/keys";
import { get } from "@/lib/http";
import { ApiResponse, Link } from "@/types";
import { DURATION_FROM_SECONDS } from "@/constants/time";

export const getLinks = async (storyId: string) => {
  const links = await get<ApiResponse<Link[]>>(`stories/${storyId}/links`, {
    next: {
      revalidate: DURATION_FROM_SECONDS.MINUTE * 5,
      tags: [linkTags.story(storyId)],
    },
  });
  return links.data;
};
