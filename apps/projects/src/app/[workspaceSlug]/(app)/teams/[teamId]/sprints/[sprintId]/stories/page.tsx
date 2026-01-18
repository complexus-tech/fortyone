import type { Metadata } from "next";
import { HydrationBoundary, dehydrate } from "@tanstack/react-query";
import { getQueryClient } from "@/app/get-query-client";
import { ListSprintStories } from "@/modules/sprints/stories/list-stories";
import { sprintKeys } from "@/constants/keys";
import { getSprintAnalytics } from "@/modules/sprints/queries/get-sprint-analytics";
import { auth } from "@/auth";
import { getSprint } from "@/modules/sprints/queries/get-sprint-details";

export async function generateMetadata({
  params,
}: {
  params: Promise<{
    teamId: string;
    sprintId: string;
    workspaceSlug: string;
  }>;
}): Promise<Metadata> {
  const { sprintId, workspaceSlug } = await params;
  const session = await auth();
  const ctx = { session: session!, workspaceSlug };
  const sprintData = await getSprint(sprintId, ctx);

  return {
    title: `${sprintData?.name || "Sprint"} › Stories`,
  };
}

export default async function Page(props: {
  params: Promise<{
    teamId: string;
    sprintId: string;
    workspaceSlug: string;
  }>;
}) {
  const params = await props.params;
  const session = await auth();
  const ctx = { session: session!, workspaceSlug: params.workspaceSlug };

  const queryClient = getQueryClient();
  await queryClient.prefetchQuery({
    queryKey: sprintKeys.analytics(params.workspaceSlug, params.sprintId),
    queryFn: () => getSprintAnalytics(params.sprintId, ctx),
  });

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <ListSprintStories sprintId={params.sprintId} />
    </HydrationBoundary>
  );
}
