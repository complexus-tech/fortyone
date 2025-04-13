import type { Metadata } from "next";
import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { StoryPage } from "@/modules/story";
import { getQueryClient } from "@/app/get-query-client";
import { getStory } from "@/modules/story/queries/get-story";
import { storyKeys } from "@/modules/stories/constants";

export const metadata: Metadata = {
  title: "Story",
};

type Props = {
  params: Promise<{
    storyId: string;
  }>;
};
export default async function Page(props: Props) {
  const params = await props.params;

  const { storyId } = params;
  const queryClient = getQueryClient();
  queryClient.prefetchQuery({
    queryKey: storyKeys.detail(storyId),
    queryFn: () => getStory(storyId),
  });

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <StoryPage storyId={storyId} />
    </HydrationBoundary>
  );
}
