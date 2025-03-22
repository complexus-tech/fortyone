import { HydrationBoundary } from "@tanstack/react-query";
import { SelectNotificationMessage } from "@/modules/notifications/message";
import { getQueryClient } from "@/app/get-query-client";
import { notificationKeys } from "@/constants/keys";
import { getNotifications } from "@/modules/notifications/queries/get-notifications";

export default async function Page() {
  const queryClient = getQueryClient();
  await queryClient.prefetchQuery({
    queryKey: notificationKeys.all,
    queryFn: () => getNotifications(),
  });

  return (
    <HydrationBoundary>
      <SelectNotificationMessage />
    </HydrationBoundary>
  );
}
