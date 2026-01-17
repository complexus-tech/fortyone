import { stringify } from "qs";
import { get, type WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { KeyResultListResponse, KeyResultFilters } from "../types";

export const getWorkspaceKeyResults = async (
  ctx: WorkspaceCtx,
  filters?: KeyResultFilters,
) => {
  const query = stringify(filters, {
    skipNulls: true,
    addQueryPrefix: true,
    encodeValuesOnly: true,
    arrayFormat: "comma",
  });

  const keyResults = await get<ApiResponse<KeyResultListResponse>>(
    `key-results${query}`,
    ctx,
  );

  return keyResults.data!;
};
