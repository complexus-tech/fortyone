"use server";

import { statusTags } from "@/constants/keys";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { State } from "@/types/states";

export const getStatuses = async () => {
  const statuses = await get<ApiResponse<State[]>>("states", {
    next: {
      revalidate: DURATION_FROM_SECONDS.MINUTE * 5,
      tags: [statusTags.lists()],
    },
  });
  return statuses.data;
};
