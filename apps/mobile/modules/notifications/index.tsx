import React from "react";
import { SafeContainer } from "@/components/ui";
import { Header } from "./components/header";
import { useNotifications } from "./hooks/use-notifications";
import {
  NotificationList,
  NotificationsSkeleton,
  EmptyState,
} from "./components";

export const Notifications = () => {
  const {
    data: notifications = [],
    isPending,
    refetch,
    isRefetching,
  } = useNotifications();

  if (isPending) {
    return (
      <SafeContainer isFull>
        <Header />
        <NotificationsSkeleton />
      </SafeContainer>
    );
  }

  if (notifications.length === 0) {
    return (
      <SafeContainer isFull>
        <Header />
        <EmptyState />
      </SafeContainer>
    );
  }

  return (
    <SafeContainer isFull>
      <Header />
      <NotificationList
        notifications={notifications}
        isLoading={isRefetching}
        onRefresh={() => refetch()}
      />
    </SafeContainer>
  );
};
