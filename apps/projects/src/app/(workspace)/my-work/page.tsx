import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from "@tanstack/react-query";
import { ListMyStories } from "@/modules/my-work";
import { getMyStories } from "@/modules/my-work/queries/get-stories";
import { storyKeys } from "@/modules/stories/constants";

export default async function Page() {
  const queryClient = new QueryClient();
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
