import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import type { Metadata } from "next";
import dynamic from "next/dynamic";
import { getMyStories } from "@/modules/my-work/queries/get-stories";
import { storyKeys } from "@/modules/stories/constants";
import { getQueryClient } from "@/app/get-query-client";

const ListMyStories = dynamic(() => import("../../../../modules/my-work"), {
  ssr: false,
});

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
