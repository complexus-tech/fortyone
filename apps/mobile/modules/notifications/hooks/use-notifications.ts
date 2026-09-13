import { useInfiniteQuery } from "@tanstack/react-query";
import { getNotifications } from "../queries/get-notifications";
import { notificationKeys } from "@/constants/keys";

export const useNotifications = () =>
  useInfiniteQuery({
    queryKey: notificationKeys.lists(),
    queryFn: ({ pageParam, signal }) => getNotifications(pageParam, signal),
    initialPageParam: 1,
    getNextPageParam: (page) =>
      page.pagination.hasMore ? page.pagination.nextPage : undefined,
  });
