import type { ReactNode } from "react";
import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { Box } from "ui";
import { BodyContainer } from "@/components/shared";
import { ListNotifications } from "@/modules/notifications/list";
import { getQueryClient } from "@/app/get-query-client";
import { getNotifications } from "@/modules/notifications/queries/get-notifications";
import { notificationKeys } from "@/constants/keys";
import { auth } from "@/auth";

export default async function Layout({
  children,
  params,
}: {
  children: ReactNode;
  params: Promise<{ workspaceSlug: string }>;
}) {
  const queryClient = getQueryClient();
  const session = await auth();
  const { workspaceSlug } = await params;
  const ctx = { session: session!, workspaceSlug };
  queryClient.prefetchQuery({
    queryKey: notificationKeys.all(workspaceSlug),
    queryFn: () => getNotifications(ctx),
  });
  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <BodyContainer className="grid h-dvh md:grid-cols-[340px_auto]">
        <ListNotifications />
        <Box className="hidden md:block">{children}</Box>
      </BodyContainer>
    </HydrationBoundary>
  );
}
