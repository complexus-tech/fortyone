import React from "react";
import { SafeContainer, Text, Row, Back } from "@/components/ui";
import { useLocalSearchParams } from "expo-router";

export default function StoryOverview() {
  const { storyId } = useLocalSearchParams();

  return (
    <SafeContainer>
      <Row align="center" gap={2}>
        <Back />
        <Text fontSize="2xl" fontWeight="semibold">
          PRO-259 /{" "}
          <Text
            fontSize="2xl"
            color="muted"
            fontWeight="semibold"
            className="opacity-80"
          >
            Overview
          </Text>
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
