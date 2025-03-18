"use client";
import { Suspense } from "react";
import { ListNotifications } from "./list";
import { NotificationsSkeleton } from "./notifications-skeleton";

export const NotificationsContainer = () => {
  return (
    <Suspense fallback={<NotificationsSkeleton />}>
      <ListNotifications />
    </Suspense>
  );
};
