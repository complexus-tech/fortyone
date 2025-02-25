"use server";

import { DURATION_FROM_SECONDS } from "@/constants/time";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { objectiveTags } from "../constants";
import type { ObjectiveStatus } from "../types";

export const getObjectiveStatuses = async () => {
  const statuses = await get<ApiResponse<ObjectiveStatus[]>>(
    "objective-statuses",
    {
      next: {
        revalidate: DURATION_FROM_SECONDS.MINUTE * 5,
        tags: [objectiveTags.statuses()],
      },
    },
  );
  return statuses.data!;
};
