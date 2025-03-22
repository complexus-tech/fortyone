import type { ReactNode } from "react";
import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { getQueryClient } from "@/app/get-query-client";
import { getMyStories } from "@/modules/my-work/queries/get-stories";
import { storyKeys } from "@/modules/stories/constants";

export default function Layout({ children }: { children: ReactNode }) {
  const queryClient = getQueryClient();
  queryClient.prefetchQuery({
    queryKey: storyKeys.mine(),
    queryFn: () => getMyStories(),
  });

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <> {children}</>
    </HydrationBoundary>
  );
}
