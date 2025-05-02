import { stringify } from "qs";
import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { SearchQueryParams, SearchResponse } from "@/modules/search/types";

export const searchQuery = async (
  session: Session,
  params: SearchQueryParams = {},
) => {
  const query = stringify(params, {
    skipNulls: true,
    addQueryPrefix: true,
    encodeValuesOnly: true,
  });
  const response = await get<ApiResponse<SearchResponse>>(
    `search${query}`,
    session,
  );
  return response.data!;
};
