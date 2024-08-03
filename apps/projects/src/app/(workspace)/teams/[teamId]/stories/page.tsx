import { ListStories } from "@/modules/teams/stories/list-stories";
import { getStories } from "@/modules/stories/queries/get-stories";
import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from "@tanstack/react-query";
import { storyKeys } from "@/modules/stories/constants";

export default async function Page({
  params: { teamId },
}: {
  params: {
    teamId: string;
  };
}) {
  const queryClient = new QueryClient();

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
