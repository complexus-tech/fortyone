import { stringify } from "qs";
import { get, type WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { SearchQueryParams, SearchResponse } from "@/modules/search/types";

export const searchQuery = async (
  ctx: WorkspaceCtx,
  params: SearchQueryParams = {},
) => {
  const query = stringify(params, {
    skipNulls: true,
    addQueryPrefix: true,
    encodeValuesOnly: true,
  });
  const response = await get<ApiResponse<SearchResponse>>(`search${query}`, ctx);
  return response.data!;
};
