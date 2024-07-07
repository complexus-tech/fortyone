"use server";
import "server-only";

import { auth } from "@/auth";
import { get } from "@/lib/http";
import { TAGS } from "@/constants/tags";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import { Story } from "@/modules/stories/types";

export const getStories = async (
  {
    tags = [],
  }: {
    createdById?: string;
    teamId?: string;
    sprintId?: string;
    objectId?: string;
    epicId?: string;
    assigneeId?: string;
    tags?: string[];
  },
  config?: RequestInit,
) => {
  const session = await auth();
  const stories = await get<Story[]>("/my-stories", {
    next: {
      revalidate: DURATION_FROM_SECONDS.MINUTE * 5,
      tags: [TAGS.stories, ...tags],
    },
    ...config,
  });
  return stories;
};
