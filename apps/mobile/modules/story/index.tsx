import React from "react";
import { ScrollView } from "react-native";
import { Properties } from "./components/properties";
import { useGlobalSearchParams } from "expo-router";
import { useStory } from "../stories/hooks/use-story";
import { Activity } from "./components/activity";
import { Description } from "./components/description";
import { Title } from "./components/title";
import { StorySkeleton } from "./components/story-skeleton";

export const Story = () => {
  const { storyId } = useGlobalSearchParams<{ storyId: string }>();
  const { data: story, isPending } = useStory(storyId);

  if (isPending) {
    return <StorySkeleton />;
  }

  return (
    <ScrollView
      className="flex-1"
      contentContainerStyle={{ paddingBottom: 100 }}
    >
      <Title story={story!} />
      <Properties story={story!} />
      <Description story={story!} />
      <Activity story={story!} />
    </ScrollView>
  );
};
