import type { Metadata } from "next";
import { HydrationBoundary, dehydrate } from "@tanstack/react-query";
import { getQueryClient } from "@/app/get-query-client";
import { ListSprintStories } from "@/modules/sprints/stories/list-stories";
import { sprintKeys } from "@/constants/keys";
import { getSprintAnalytics } from "@/modules/sprints/queries/get-sprint-analytics";
import { auth } from "@/auth";
import { getSprintDetails } from "@/modules/sprints/queries/get-sprint-details";

export async function generateMetadata({
  params,
}: {
  params: Promise<{
    teamId: string;
    sprintId: string;
  }>;
}): Promise<Metadata> {
  const { sprintId } = await params;
  const session = await auth();
  const sprintData = await getSprintDetails(sprintId, session!);

  return {
    title: `${sprintData?.name || "Sprint"} › Stories`,
  };
}

export default async function Page(props: {
  params: Promise<{
    teamId: string;
    sprintId: string;
  }>;
}) {
  const params = await props.params;
  const session = await auth();

  const queryClient = getQueryClient();
  await queryClient.prefetchQuery({
    queryKey: sprintKeys.analytics(params.sprintId),
    queryFn: () => getSprintAnalytics(params.sprintId, session!),
  });

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <ListSprintStories sprintId={params.sprintId} />
    </HydrationBoundary>
  );
}
