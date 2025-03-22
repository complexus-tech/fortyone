import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import type { Metadata } from "next";
import { getMyStories } from "@/modules/my-work/queries/get-stories";
import { storyKeys } from "@/modules/stories/constants";
import { getQueryClient } from "@/app/get-query-client";
import { ListMyStories } from "@/modules/my-work";

export const metadata: Metadata = {
  title: "My Work",
};

export default async function Page() {
  const queryClient = getQueryClient();
  await queryClient.prefetchQuery({
    queryKey: storyKeys.mine(),
    queryFn: () => getMyStories(),
  });

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <ListMyStories />
    </HydrationBoundary>
  );
}
