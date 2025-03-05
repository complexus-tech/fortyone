"use server";

import { stringify } from "qs";
import { get } from "@/lib/http";
import type { Sprint } from "@/modules/sprints/types";
import type { ApiResponse } from "@/types";

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

  const sprints = await get<ApiResponse<Sprint[]>>(`sprints${query}`);
  return sprints.data!;
};
