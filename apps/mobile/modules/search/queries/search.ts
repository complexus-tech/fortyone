import { stringify } from "qs";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { SearchQueryParams, SearchResponse } from "../types";

export const searchQuery = async (params: SearchQueryParams = {}) => {
  const query = stringify(params, {
    skipNulls: true,
    addQueryPrefix: true,
    encodeValuesOnly: true,
  });
  const response = await get<ApiResponse<SearchResponse>>(`search${query}`);
  return response.data;
};
