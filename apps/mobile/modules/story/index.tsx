import React from "react";
import { ScrollView } from "react-native";
import { Properties } from "./components/properties";
import { useGlobalSearchParams } from "expo-router";
import { useStory } from "../stories/hooks/use-story";
import { Text } from "@/components/ui";
import { Activity } from "./components/activity";
import { Description } from "./components/descrition";
import { Title } from "./components/title";

export const Story = () => {
  const { storyId } = useGlobalSearchParams<{ storyId: string }>();
  const { data: story, isPending } = useStory(storyId);
  if (isPending) {
    return <Text>Loading...</Text>;
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
