"use server";

import qs from "qs";
import { get } from "@/lib/http";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import type { Sprint } from "@/modules/sprints/types";
import type { ApiResponse } from "@/types";

export const getTeamSprints = async (teamId: string) => {
  const query = qs.stringify(
    { teamId },
    {
      skipNulls: true,
      addQueryPrefix: true,
      encodeValuesOnly: true,
    },
  );

  const sprints = await get<ApiResponse<Sprint[]>>(`sprints${query}`, {
    next: {
      revalidate: DURATION_FROM_SECONDS.MINUTE * 5,
      tags: [`team-sprints-${teamId}`],
    },
  });
  return sprints;
};
