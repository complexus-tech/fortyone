import type { Metadata } from "next";
import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { ArrowLeft2Icon, StoryMissingIcon } from "icons";
import { Box, Button, Text } from "ui";
import { cache } from "react";
import { StoryPage } from "@/modules/story";
import { getQueryClient } from "@/app/get-query-client";
import { getStory } from "@/modules/story/queries/get-story";
import { storyKeys } from "@/modules/stories/constants";
import { auth } from "@/auth";
import { withWorkspacePath } from "@/utils";

const getSession = cache(async () => auth());
const getStoryData = cache(async (storyId: string, workspaceSlug: string) => {
  const session = await getSession();
  if (!session) return null;

  return getStory(storyId, { session, workspaceSlug });
});

export async function generateMetadata({
  params,
}: {
  params: Promise<{ storyId: string; workspaceSlug: string }>;
}): Promise<Metadata> {
  const { storyId, workspaceSlug } = await params;
  const story = await getStoryData(storyId, workspaceSlug);
  return {
    title: story?.title || "Story",
  };
}

type Props = {
  params: Promise<{
    storyId: string;
    workspaceSlug: string;
  }>;
};
export default async function Page(props: Props) {
  const params = await props.params;

  const { storyId, workspaceSlug } = params;
  const queryClient = getQueryClient();
  const story = await queryClient.fetchQuery({
    queryKey: storyKeys.detail(workspaceSlug, storyId),
    queryFn: () => getStoryData(storyId, workspaceSlug),
  });

  if (!story) {
    return (
      <Box className="flex h-screen items-center justify-center">
        <Box className="flex flex-col items-center">
          <StoryMissingIcon className="h-20 w-auto rotate-12" />
          <Text className="mt-10 mb-6" fontSize="3xl">
            404: Item not found
          </Text>
          <Text className="mb-6 max-w-md text-center" color="muted">
            This item might not exist or you do not have access to it.
          </Text>
          <Button
            className="gap-1 pl-2"
            color="tertiary"
            href={withWorkspacePath("/my-work", workspaceSlug)}
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
