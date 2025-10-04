import React from "react";
import { useGlobalSearchParams } from "expo-router";
import { StoriesSkeleton } from "@/components/ui";
import { useStory } from "@/modules/stories/hooks";
import { List } from "./components";

export const SubStories = () => {
  const { storyId } = useGlobalSearchParams<{ storyId: string }>();
  const { data: story, isPending } = useStory(storyId);

  if (isPending) {
    return <StoriesSkeleton count={5} />;
  }

  return <List stories={story?.subStories || []} />;
};
