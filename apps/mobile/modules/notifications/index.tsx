import React from "react";
import { SafeContainer } from "@/components/ui";
import { Header } from "./components/header";
import { useNotifications } from "./hooks/use-notifications";
import {
  NotificationList,
  NotificationsSkeleton,
  EmptyState,
} from "./components";
import { useQueryClient } from "@tanstack/react-query";
import { notificationKeys } from "@/constants/keys";

export const Notifications = () => {
  const queryClient = useQueryClient();
  const { data: notifications = [], isPending, refetch } = useNotifications();

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
        onRefresh={() => {
          refetch();
          queryClient.invalidateQueries({
            queryKey: notificationKeys.all,
          });
        }}
      />
    </SafeContainer>
  );
};
