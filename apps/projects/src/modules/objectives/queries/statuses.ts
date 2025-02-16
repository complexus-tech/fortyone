"use server";

import { DURATION_FROM_SECONDS } from "@/constants/time";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { State } from "@/types/states";
import { objectiveTags } from "../constants";

export const getObjectiveStatuses = async () => {
  const statuses = await get<ApiResponse<State[]>>("objective-statuses", {
    next: {
      revalidate: DURATION_FROM_SECONDS.MINUTE * 5,
      tags: [objectiveTags.statuses()],
    },
  });
  return statuses.data!;
};
