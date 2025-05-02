import { stringify } from "qs";
import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { Sprint } from "@/modules/sprints/types";
import type { ApiResponse } from "@/types";

export const getObjectiveSprints = async (
  objectiveId: string,
  session: Session,
) => {
  if (!objectiveId) return [];
  const query = stringify(
    { objectiveId },
    {
      skipNulls: true,
      addQueryPrefix: true,
      encodeValuesOnly: true,
    },
  );

  const sprints = await get<ApiResponse<Sprint[]>>(`sprints${query}`, session);
  return sprints.data!;
};
