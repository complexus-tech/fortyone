import { useQuery } from "@tanstack/react-query";
import { searchKeys } from "@/constants/keys";
import { searchQuery } from "../queries/search";
import type { SearchQueryParams } from "../types";

export const useSearch = (params: SearchQueryParams = {}) => {
  return useQuery({
    queryKey: searchKeys.query(params),
    queryFn: () => searchQuery(params),
    enabled: Boolean(params.query),
    staleTime: 1000 * 60 * 3, // 3 minutes
  });
};
