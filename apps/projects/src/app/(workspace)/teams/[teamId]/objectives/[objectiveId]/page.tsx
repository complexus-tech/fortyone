import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { getStories } from "@/modules/stories/queries/get-stories";
import { storyKeys, storyTags } from "@/modules/stories/constants";
import { getQueryClient } from "@/app/get-query-client";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import { ListStories } from "@/modules/objectives/stories/list-stories";
import { getObjective } from "@/modules/objectives/queries/get-objective";
import { objectiveKeys } from "@/modules/objectives/constants";

export default async function Page(props: {
  params: Promise<{
    teamId: string;
    objectiveId: string;
  }>;
}) {
  const params = await props.params;

  const { objectiveId } = params;

  const queryClient = getQueryClient();

  await Promise.all([
    queryClient.prefetchQuery({
      queryKey: storyKeys.objective(objectiveId),
      queryFn: () =>
        getStories(
          { objectiveId },
          {
            next: {
              revalidate: DURATION_FROM_SECONDS.MINUTE * 5,
              tags: [storyTags.objectives(), storyTags.objective(objectiveId)],
            },
          },
        ),
    }),
    queryClient.prefetchQuery({
      queryKey: objectiveKeys.objective(objectiveId),
      queryFn: () => getObjective(objectiveId),
    }),
  ]);

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <ListStories />
    </HydrationBoundary>
  );
}
