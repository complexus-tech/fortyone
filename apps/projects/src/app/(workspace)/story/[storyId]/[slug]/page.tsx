import { HydrationBoundary, dehydrate } from "@tanstack/react-query";
import type { Metadata } from "next";
import { getQueryClient } from "@/app/get-query-client";
import { storyKeys } from "@/modules/stories/constants";
import { StoryPage } from "@/modules/story";
import { getStoryActivities } from "@/modules/story/queries/get-activities";
import { getStory } from "@/modules/story/queries/get-story";

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
  await Promise.all([
    queryClient.prefetchQuery({
      queryKey: storyKeys.detail(storyId),
      queryFn: () => getStory(storyId),
    }),
    queryClient.prefetchQuery({
      queryKey: storyKeys.activities(storyId),
      queryFn: () => getStoryActivities(storyId),
    }),
  ]);

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <StoryPage />
    </HydrationBoundary>
  );
}
