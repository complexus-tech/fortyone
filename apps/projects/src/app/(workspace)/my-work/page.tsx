import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { ListMyStories } from "@/modules/my-work";
import { getMyStories } from "@/modules/my-work/queries/get-stories";
import { storyKeys } from "@/modules/stories/constants";
import { getQueryClient } from "@/app/get-query-client";

export default async function Page() {
  const queryClient = getQueryClient();
  await queryClient.prefetchQuery({
    queryKey: storyKeys.mine(),
    queryFn: getMyStories,
  });

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <ListMyStories />
    </HydrationBoundary>
  );
}
