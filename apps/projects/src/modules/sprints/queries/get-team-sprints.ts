import { stringify } from "qs";
import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { Sprint } from "@/modules/sprints/types";
import type { ApiResponse } from "@/types";

export const getTeamSprints = async (teamId: string, session: Session) => {
  if (!teamId) return [];
  const query = stringify(
    { teamId },
    {
      skipNulls: true,
      addQueryPrefix: true,
      encodeValuesOnly: true,
    },
  );

  const sprints = await get<ApiResponse<Sprint[]>>(`sprints${query}`, session);
  return sprints.data!;
};
