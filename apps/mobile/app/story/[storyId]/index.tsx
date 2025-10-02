import React from "react";
import { SafeContainer, Text } from "@/components/ui";
import { useLocalSearchParams } from "expo-router";
import { Header } from "@/modules/story/components/header";

export default function StoryOverview() {
  const { storyId } = useLocalSearchParams();

  return (
    <SafeContainer>
      <Header />

      <Text color="muted" className="mt-4">
        Story ID: {storyId}
      </Text>
      <Text color="muted" className="mt-2">
        This is the Overview tab for the story details page.
      </Text>
    </SafeContainer>
  );
}
