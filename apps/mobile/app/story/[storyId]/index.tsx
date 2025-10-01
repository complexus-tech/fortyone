import React from "react";
import { SafeContainer, Text, Row } from "@/components/ui";
import { useLocalSearchParams } from "expo-router";

export default function StoryOverview() {
  const { storyId } = useLocalSearchParams();

  return (
    <SafeContainer edges={["bottom"]}>
      <Row className="pb-2" justify="between" align="center">
        <Text fontSize="2xl" fontWeight="semibold">
          Overview
        </Text>
      </Row>

      <Text color="muted" className="mt-4">
        Story ID: {storyId}
      </Text>
      <Text color="muted" className="mt-2">
        This is the Overview tab for the story details page.
      </Text>
    </SafeContainer>
  );
}
