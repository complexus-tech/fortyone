import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import { labelKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getLabels, getLabelsPage } from "../queries/labels/get-labels";

export const LABEL_MENU_PAGE_SIZE = 15;

export const useLabels = (
  params: {
    search?: string;
    teamId?: string;
  } = {},
) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const normalizedSearch = params.search?.trim() ?? "";

  return useQuery({
    queryKey: [
      ...labelKeys.lists(workspaceSlug),
      params.teamId ?? "",
      normalizedSearch,
    ],
    queryFn: () =>
      getLabels(
        { session: session!, workspaceSlug },
        { ...params, search: normalizedSearch || undefined },
      ),
    enabled: Boolean(session),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};

export const useLabelsInfinite = (
  params: {
    search?: string;
    teamId?: string;
  } = {},
  pageSize = LABEL_MENU_PAGE_SIZE,
  enabled = true,
) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const normalizedSearch = params.search?.trim() ?? "";

  return useInfiniteQuery({
    queryKey: [
      ...labelKeys.lists(workspaceSlug),
      "infinite",
      params.teamId ?? "",
      normalizedSearch,
      pageSize,
    ] as const,
    queryFn: ({ pageParam }) =>
      getLabelsPage(
        { session: session!, workspaceSlug },
        {
          page: pageParam,
          pageSize,
          search: normalizedSearch || undefined,
          teamId: params.teamId,
        },
      ),
    getNextPageParam: (lastPage) =>
      lastPage.pagination.hasMore ? lastPage.pagination.nextPage : undefined,
    initialPageParam: 1,
    enabled: Boolean(session && enabled),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
