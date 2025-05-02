import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { searchQuery } from "../queries/search";
import type { SearchQueryParams } from "../types";

const searchKeys = {
  all: ["search"] as const,
  query: (params: SearchQueryParams) => [...searchKeys.all, params] as const,
};

export const useSearch = (params: SearchQueryParams = {}) => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: searchKeys.query(params),
    queryFn: () => searchQuery(session!, params),
    enabled: Boolean(params.query),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 3,
  });
};
