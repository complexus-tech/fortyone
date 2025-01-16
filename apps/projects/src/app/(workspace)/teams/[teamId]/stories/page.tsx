import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import type { Metadata } from "next";
import { ListStories } from "@/modules/teams/stories/list-stories";
import { getStories } from "@/modules/stories/queries/get-stories";
import { storyKeys, storyTags } from "@/modules/stories/constants";
import { getQueryClient } from "@/app/get-query-client";
import { DURATION_FROM_SECONDS } from "@/constants/time";

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
    queryFn: () =>
      getStories(
        { teamId },
        {
          next: {
            revalidate: DURATION_FROM_SECONDS.MINUTE * 5,
            tags: [storyTags.team(teamId)],
          },
        },
      ),
  });
  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <ListStories />
    </HydrationBoundary>
  );
}
