"use server";
import "server-only";

import { auth } from "@/auth";
import { get } from "@/lib/http";
import { TAGS } from "@/constants/tags";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import { Sprint } from "../types";

export const getTeamSprints = async () => {
  const session = await auth();
  const sprints = await get<Sprint[]>("/sprints", {
    next: {
      revalidate: DURATION_FROM_SECONDS.HOUR * 1,
      tags: [TAGS.sprints],
    },
  });
  return sprints;
};
