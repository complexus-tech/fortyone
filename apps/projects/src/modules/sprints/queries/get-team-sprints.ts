"use server";

import { stringify } from "qs";
import { get } from "@/lib/http";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import type { Sprint } from "@/modules/sprints/types";
import type { ApiResponse } from "@/types";
import { sprintTags } from "@/constants/keys";

export const getTeamSprints = async (teamId: string) => {
  if (!teamId) return [];
  const query = stringify(
    { teamId },
    {
      skipNulls: true,
      addQueryPrefix: true,
      encodeValuesOnly: true,
    },
  );

  const sprints = await get<ApiResponse<Sprint[]>>(`sprints${query}`, {
    next: {
      revalidate: DURATION_FROM_SECONDS.MINUTE * 30,
      tags: [sprintTags.lists(), sprintTags.team(teamId)],
    },
  });
  return sprints.data!;
};
