import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { searchQuery } from "../queries/search";
import type { SearchQueryParams } from "../types";
import { useWorkspacePath } from "@/hooks";

const searchKeys = {
  all: (workspaceSlug: string) => ["search", workspaceSlug] as const,
  query: (workspaceSlug: string, params: SearchQueryParams) => [...searchKeys.all(workspaceSlug), params] as const,
};

export const useSearch = (params: SearchQueryParams = {}) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    queryKey: searchKeys.query(workspaceSlug, params),
    queryFn: () => searchQuery({ session: session!, workspaceSlug }, params),
    enabled: Boolean(params.query),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 3,
  });
};
