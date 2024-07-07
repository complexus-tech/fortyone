"use server";

import { get } from "@/lib/http";
import "server-only";
import { Objective } from "../types";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import { TAGS } from "@/constants/tags";

export const getObjectives = async () => {
  const objectives = await get<Objective[]>("/objectives", {
    next: {
      revalidate: DURATION_FROM_SECONDS.HOUR * 5,
      tags: [TAGS.objectives],
    },
  });
  return objectives;
};
