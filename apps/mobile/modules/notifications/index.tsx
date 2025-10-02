import React from "react";
import { ScrollView } from "react-native";
import { SafeContainer } from "@/components/ui";
import { Header } from "./components/header";
import { useNotifications } from "./hooks/use-notifications";
import {
  NotificationCard,
  NotificationsSkeleton,
  EmptyState,
} from "./components";

export const Notifications = () => {
  const { data: notifications = [], isPending } = useNotifications();

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
      <ScrollView className="flex-1">
        {notifications.map((notification, index) => (
          <NotificationCard
            key={notification.id}
            {...notification}
            index={index}
          />
        ))}
      </ScrollView>
    </SafeContainer>
  );
};
