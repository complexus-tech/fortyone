import type { Metadata } from "next";
import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { redirect } from "next/navigation";
import { StoryPage } from "@/modules/story";
import { getQueryClient } from "@/app/get-query-client";
import { StoryRouteNotFound } from "@/modules/story/components/story-route-not-found";
import { getStoryRouteData } from "@/modules/story/server/get-story-route-data";
import { getStoryPath, getStoryReference } from "@/shared/routing/story";
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
  searchParams,
}: {
  params: Promise<{ storyRef: string; workspaceSlug: string }>;
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const { storyRef, workspaceSlug } = await params;
  const story = await getStoryRouteData(storyRef, workspaceSlug);

  if (!story) {
    return <StoryRouteNotFound workspaceSlug={workspaceSlug} />;
  }

  const canonicalReference = getStoryReference(story);
  if (storyRef !== canonicalReference) {
    const query = new URLSearchParams();
    for (const [key, value] of Object.entries(await searchParams)) {
      if (Array.isArray(value)) {
        value.forEach((item) => {
          query.append(key, item);
        });
      } else if (value !== undefined) query.set(key, value);
    }
    const path = withWorkspacePath(getStoryPath(story), workspaceSlug);
    const suffix = query.toString();
    redirect(suffix ? `${path}?${suffix}` : path);
  }

  const queryClient = getQueryClient();
  queryClient.setQueryData(storyKeys.detail(workspaceSlug, story.id), story);

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <StoryPage storyId={story.id} />
    </HydrationBoundary>
  );
}
