import type { Metadata } from "next";
import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { notFound } from "next/navigation";
import { StoryPage } from "@/modules/story";
import { getQueryClient } from "@/app/get-query-client";
import { getStory } from "@/modules/story/queries/get-story";
import { storyKeys } from "@/modules/stories/constants";
import { auth } from "@/auth";

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
  const session = await auth();

  const { storyId } = params;
  const queryClient = getQueryClient();
  const story = await queryClient.fetchQuery({
    queryKey: storyKeys.detail(storyId),
    queryFn: () => getStory(storyId, session!),
  });

  if (!story) {
    notFound();
  }

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <StoryPage storyId={storyId} />
    </HydrationBoundary>
  );
}
