import type { ReactNode } from "react";
import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { Box } from "ui";
import { BodyContainer } from "@/components/shared";
import { auth } from "@/auth";
import { getQueryClient } from "@/app/get-query-client";
import { integrationRequestKeys } from "@/constants/keys";
import { ListIntegrationRequests } from "@/modules/integration-requests/list";
import { getTeamIntegrationRequests } from "@/modules/integration-requests/queries/get-team-requests";

export default async function Layout({
  children,
  params,
}: {
  children: ReactNode;
  params: Promise<{ teamId: string; workspaceSlug: string }>;
}) {
  const queryClient = getQueryClient();
  const session = await auth();
  const { teamId, workspaceSlug } = await params;
  const ctx = { session: session!, workspaceSlug };

  queryClient.prefetchQuery({
    queryKey: integrationRequestKeys.team(workspaceSlug, teamId),
    queryFn: () => getTeamIntegrationRequests(teamId, ctx),
  });

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <BodyContainer className="grid h-dvh md:grid-cols-[340px_auto]">
        <ListIntegrationRequests />
        <Box className="hidden md:block">{children}</Box>
      </BodyContainer>
    </HydrationBoundary>
  );
}
