import type { Metadata } from "next";
import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { ListMyStories } from "@/modules/my-work";
import { getQueryClient } from "@/app/get-query-client";
import { getMyStories } from "@/modules/my-work/queries/get-stories";
import { storyKeys } from "@/modules/stories/constants";

export const metadata: Metadata = {
  title: "My Work",
};

export default function Page() {
  const queryClient = getQueryClient();
  queryClient.prefetchQuery({
    queryKey: storyKeys.mine(),
    queryFn: () => getMyStories(),
  });

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <ListMyStories />
    </HydrationBoundary>
  );
}
