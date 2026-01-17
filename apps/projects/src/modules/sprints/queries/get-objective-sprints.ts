import { stringify } from "qs";
import { get, type WorkspaceCtx } from "@/lib/http";
import type { Sprint } from "@/modules/sprints/types";
import type { ApiResponse } from "@/types";

export const getObjectiveSprints = async (
  objectiveId: string,
  ctx: WorkspaceCtx,
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

  const sprints = await get<ApiResponse<Sprint[]>>(`sprints${query}`, ctx);
  return sprints.data!;
};
