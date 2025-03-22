import type { Metadata } from "next";
import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { ListMyStories } from "@/modules/my-work";
import { getQueryClient } from "@/app/get-query-client";
import { getMyStories } from "@/modules/my-work/queries/get-stories";
import { storyKeys } from "@/modules/stories/constants";

export const metadata: Metadata = {
  title: "My Work",
};

export default async function Page() {
  const queryClient = getQueryClient();

  // Start timing
  const startTime = performance.now();

  await queryClient.prefetchQuery({
    queryKey: storyKeys.mine(),
    queryFn: () => getMyStories(),
  });

  // End timing
  const endTime = performance.now();
  // eslint-disable-next-line no-console -- debug
  console.log(`Query prefetching took ${endTime - startTime}ms`);

  const dehydratedState = dehydrate(queryClient);
  // eslint-disable-next-line no-console -- debug
  console.log(`Dehydration took ${performance.now() - endTime}ms`);
  // eslint-disable-next-line no-console -- debug
  console.log("Dehydrated state:", dehydratedState);

  return (
    <HydrationBoundary state={dehydratedState}>
      <ListMyStories />
    </HydrationBoundary>
  );
}
