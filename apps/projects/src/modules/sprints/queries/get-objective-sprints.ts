import { stringify } from "qs";
import { get } from "@/lib/http";
import type { Sprint } from "@/modules/sprints/types";
import type { ApiResponse } from "@/types";

export const getObjectiveSprints = async (objectiveId: string) => {
  if (!objectiveId) return [];
  const query = stringify(
    { objectiveId },
    {
      skipNulls: true,
      addQueryPrefix: true,
      encodeValuesOnly: true,
    },
  );

  const sprints = await get<ApiResponse<Sprint[]>>(`sprints${query}`);
  return sprints.data!;
};
