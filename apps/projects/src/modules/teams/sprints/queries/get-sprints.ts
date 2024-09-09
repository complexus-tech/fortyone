"use server";

import { get } from "@/lib/http";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import { Sprint } from "@/modules/sprints/types";
import qs from "qs";

export const getTeamSprints = async (teamId: string) => {
  const query = qs.stringify(
    { teamId },
    {
      skipNulls: true,
      addQueryPrefix: true,
      encodeValuesOnly: true,
    },
  );

  const sprints = await get<Sprint[]>(`sprints${query}`, {
    next: {
      revalidate: DURATION_FROM_SECONDS.MINUTE * 5,
      tags: [`team-sprints-${teamId}`],
    },
  });
  return sprints;
};
