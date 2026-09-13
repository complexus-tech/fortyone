import React from "react";
import { SafeContainer } from "@/components/ui";
import { QueryState } from "@/components/ui/query-state";
import { Header } from "./components/header";
import { useNotifications } from "./hooks/use-notifications";
import { NotificationList, NotificationsSkeleton } from "./components";

export const Notifications = () => {
  const query = useNotifications();
  const notifications =
    query.data?.pages.flatMap((page) => page.notifications) ?? [];

  return (
    <SafeContainer isFull>
      <Header />
      {query.isPending ? (
        <NotificationsSkeleton />
      ) : query.isError && !query.data ? (
        <QueryState
          title="Could not load your inbox"
          message={query.error.message}
          onRetry={() => {
            void query.refetch();
          }}
        />
      ) : (
        <NotificationList
          notifications={notifications}
          isLoading={query.isRefetching && !query.isFetchingNextPage}
          onRefresh={() => {
            void query.refetch();
          }}
          hasMore={query.hasNextPage}
          isLoadingMore={query.isFetchingNextPage}
          onLoadMore={() => {
            void query.fetchNextPage();
          }}
          error={query.isError ? query.error : null}
        />
      )}
    </SafeContainer>
  );
};
