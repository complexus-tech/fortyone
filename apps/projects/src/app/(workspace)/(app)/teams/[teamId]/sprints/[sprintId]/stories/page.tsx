import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { getStories } from "@/modules/stories/queries/get-stories";
import { ListSprintStories } from "@/modules/sprints/stories/list-stories";
import { storyKeys } from "@/modules/stories/constants";
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

  queryClient.prefetchQuery({
    queryKey: storyKeys.sprint(sprintId),
    queryFn: () => getStories({ sprintId }),
  });

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <ListSprintStories />
    </HydrationBoundary>
  );
}
