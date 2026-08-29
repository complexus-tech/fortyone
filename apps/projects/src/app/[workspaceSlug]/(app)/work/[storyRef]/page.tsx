import type { Metadata } from "next";
import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { redirect } from "next/navigation";
import { StoryPage } from "@/modules/story";
import { getQueryClient } from "@/app/get-query-client";
import { StoryRouteNotFound } from "@/modules/story/components/story-route-not-found";
import { getStoryRouteData } from "@/modules/story/server/get-story-route-data";
import {
  getStoryPath,
  getStoryReference,
} from "@/modules/story/utils/story-url";
import { storyKeys } from "@/modules/stories/constants";
import { withWorkspacePath } from "@/utils";

export async function generateMetadata({
  params,
}: {
  params: Promise<{ storyRef: string; workspaceSlug: string }>;
}): Promise<Metadata> {
  const { storyRef, workspaceSlug } = await params;
  const story = await getStoryRouteData(storyRef, workspaceSlug);

  return {
    title: story?.title || "Work item",
  };
}

export default async function Page({
  params,
}: {
  params: Promise<{ storyRef: string; workspaceSlug: string }>;
}) {
  const { storyRef, workspaceSlug } = await params;
  const story = await getStoryRouteData(storyRef, workspaceSlug);

  if (!story) {
    return <StoryRouteNotFound workspaceSlug={workspaceSlug} />;
  }

  const canonicalReference = getStoryReference(story);
  if (storyRef !== canonicalReference) {
    redirect(withWorkspacePath(getStoryPath(story), workspaceSlug));
  }

  const queryClient = getQueryClient();
  queryClient.setQueryData(storyKeys.detail(workspaceSlug, story.id), story);

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <StoryPage storyId={story.id} />
    </HydrationBoundary>
  );
}
