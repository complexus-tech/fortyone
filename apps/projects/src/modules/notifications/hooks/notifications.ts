import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import { notificationKeys } from "@/constants/keys";
import {
  getNotifications,
  getNotificationsPage,
} from "../queries/get-notifications";

export const useNotifications = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    queryKey: notificationKeys.all(workspaceSlug),
    queryFn: () => getNotifications({ session: session!, workspaceSlug }),
  });
};

export const useNotificationsInfinite = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useInfiniteQuery({
    queryKey: [...notificationKeys.all(workspaceSlug), "infinite"] as const,
    queryFn: ({ pageParam }) =>
      getNotificationsPage({ session: session!, workspaceSlug }, pageParam),
    getNextPageParam: (lastPage) =>
      lastPage.pagination.hasMore ? lastPage.pagination.nextPage : undefined,
    initialPageParam: 1,
    enabled: Boolean(session),
  });
};
