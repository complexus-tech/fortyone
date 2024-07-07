"use server";
import "server-only";

import { get } from "@/lib/http";
import { Sprint } from "@/modules/sprints/types";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import { TAGS } from "@/constants/tags";

export const getRunningSprints = async () => {
  const sprints = await get<Sprint[]>("/sprints", {
    next: {
      revalidate: DURATION_FROM_SECONDS.HOUR,
      tags: [TAGS.sprints],
    },
  });
  return sprints;
};
