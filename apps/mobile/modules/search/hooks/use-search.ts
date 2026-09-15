import { useInfiniteQuery } from "@tanstack/react-query";
import { searchKeys } from "@/constants/keys";
import { searchQuery } from "../queries/search";
import type { SearchQueryParams } from "../types";
import { getNextSearchPage } from "../pagination";

export const useSearch = (params: SearchQueryParams = {}) => {
  return useInfiniteQuery({
    queryKey: [...searchKeys.query(params), "infinite"],
    queryFn: ({ signal, pageParam }) =>
      searchQuery({ ...params, page: pageParam }, signal),
    initialPageParam: 1,
    getNextPageParam: getNextSearchPage,
    enabled: Boolean(params.query?.trim()),
    staleTime: 1000 * 60 * 3, // 3 minutes
  });
};
