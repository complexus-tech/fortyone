import type { Metadata } from "next";
import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { Box, Button, Text } from "ui";
import { ArrowLeft2Icon, StoryMissingIcon } from "icons";
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
    return (
      <Box className="flex h-screen items-center justify-center">
        <Box className="flex flex-col items-center">
          <StoryMissingIcon className="h-20 w-auto rotate-12" />
          <Text className="mb-6 mt-10" fontSize="3xl">
            404: Item not found
          </Text>
          <Text className="mb-6 max-w-md text-center" color="muted">
            This item might not exist or you do not have access to it.
          </Text>
          <Button
            className="gap-1 pl-2"
            color="tertiary"
            href="/my-work"
            leftIcon={<ArrowLeft2Icon className="h-[1.05rem] w-auto" />}
          >
            Go to my work
          </Button>
        </Box>
      </Box>
    );
  }

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <StoryPage storyId={storyId} />
    </HydrationBoundary>
  );
}
