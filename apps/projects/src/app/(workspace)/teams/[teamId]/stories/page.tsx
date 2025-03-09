import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import type { Metadata } from "next";
import { ListStories } from "@/modules/teams/stories/list-stories";
import { getStories } from "@/modules/stories/queries/get-stories";
import { storyKeys } from "@/modules/stories/constants";
import { getQueryClient } from "@/app/get-query-client";

export const metadata: Metadata = {
  title: "Stories",
};

export default async function Page(props: {
  params: Promise<{
    teamId: string;
  }>;
}) {
  const params = await props.params;

  const { teamId } = params;

  const queryClient = getQueryClient();

  await queryClient.prefetchQuery({
    queryKey: storyKeys.team(teamId),
    queryFn: () => getStories({ teamId }),
  });
  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <ListStories />
    </HydrationBoundary>
  );
}
