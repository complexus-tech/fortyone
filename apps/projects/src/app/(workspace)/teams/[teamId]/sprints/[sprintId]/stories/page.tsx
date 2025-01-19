import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import { getStories } from "@/modules/stories/queries/get-stories";
import { ListSprintStories } from "@/modules/sprints/stories/list-stories";
import { storyKeys, storyTags } from "@/modules/stories/constants";
import { getQueryClient } from "@/app/get-query-client";

export default async function Page(props: {
  params: Promise<{
    teamId: string;
    sprintId: string;
  }>;
}) {
  const params = await props.params;
  const queryClient = getQueryClient();

  const { sprintId } = params;

  await queryClient.prefetchQuery({
    queryKey: storyKeys.sprint(sprintId),
    queryFn: () =>
      getStories(
        { sprintId },
        {
          next: {
            revalidate: DURATION_FROM_SECONDS.MINUTE * 5,
            tags: [storyTags.sprints(), storyTags.sprint(sprintId)],
          },
        },
      ),
  });

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <ListSprintStories />
    </HydrationBoundary>
  );
}
