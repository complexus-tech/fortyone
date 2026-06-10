import { stringify } from "qs";
import { get, type WorkspaceCtx } from "@/lib/http";
import type { Sprint } from "@/modules/sprints/types";
import type { ApiResponse } from "@/types";

export const getTeamSprints = async (
  teamId: string,
  ctx: WorkspaceCtx,
  search = "",
) => {
  if (!teamId) return [];
  const query = stringify(
    { search: search.trim() || undefined, teamId },
    {
      skipNulls: true,
      addQueryPrefix: true,
      encodeValuesOnly: true,
    },
  );

  const sprints = await get<ApiResponse<Sprint[]>>(`sprints${query}`, ctx);
  return sprints.data!;
};
