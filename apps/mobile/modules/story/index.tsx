import React from "react";
import { View } from "react-native";
import { Title } from "./components/title";
import { Properties } from "./components/properties";
import { useGlobalSearchParams } from "expo-router";
import { useStory } from "../stories/hooks/use-story";
import { Text } from "@/components/ui";
import { Activity } from "./components/activity";

export const Story = () => {
  const { storyId } = useGlobalSearchParams<{ storyId: string }>();
  const { data: story, isPending } = useStory(storyId);
  if (isPending) {
    return <Text>Loading...</Text>;
  }
  return (
    <View className="flex-1">
      <Title title={story?.title} />
      <Properties story={story!} />
      <Activity story={story!} />
    </View>
  );
};
