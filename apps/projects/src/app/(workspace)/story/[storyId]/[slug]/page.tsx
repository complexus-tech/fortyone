import { getQueryClient } from "@/app/get-query-client";
import { storyKeys } from "@/modules/stories/constants";
import { StoryPage } from "@/modules/story";
import { getStory } from "@/modules/story/queries/get-story";
import { HydrationBoundary, dehydrate } from "@tanstack/react-query";

type Props = {
  params: Promise<{
    storyId: string;
  }>;
};
export default async function Page(props: Props) {
  const params = await props.params;

  const {
    storyId
  } = params;

  const queryClient = getQueryClient();

  await queryClient.prefetchQuery({
    queryKey: storyKeys.detail(storyId),
    queryFn: () => getStory(storyId),
  });

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <StoryPage />
    </HydrationBoundary>
  );
}
