import { useQuery } from "@tanstack/react-query";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { searchQuery } from "../queries/search";
import type { SearchQueryParams } from "../types";

const searchKeys = {
  all: ["search"] as const,
  query: (params: SearchQueryParams) => [...searchKeys.all, params] as const,
};

export const useSearch = (params: SearchQueryParams = {}) => {
  return useQuery({
    queryKey: searchKeys.query(params),
    queryFn: () => searchQuery(params),
    enabled: Boolean(params.query),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 3,
  });
};
