import React from "react";
import { useLocalSearchParams } from "expo-router";
import { StoriesSkeleton } from "@/components/ui";
import { useStory } from "@/modules/stories/hooks";
import { List } from "./components";
import { QueryState } from "@/components/ui/query-state";

export const SubStories = () => {
  const { storyId } = useLocalSearchParams<{ storyId: string }>();
  const { data: story, isPending, error, refetch } = useStory(storyId);

  if (isPending) {
    return <StoriesSkeleton count={5} />;
  }

  if (error && !story)
    return (
      <QueryState
        title="Couldn’t load subtasks"
        message="Check your connection and try again."
        onRetry={() => {
          void refetch();
        }}
      />
    );

  return <List stories={story?.subStories || []} />;
};
