import React from "react";
import { Text, Row } from "@/components/ui";
import { useGlobalSearchParams } from "expo-router";
import { useStory } from "@/modules/stories/hooks";

export const Title = () => {
  const { storyId } = useGlobalSearchParams<{ storyId: string }>();
  const { data: story } = useStory(storyId);
  return (
    <Row asContainer>
      <Text fontSize="xl" fontWeight="semibold">
        {story?.title}
      </Text>
    </Row>
  );
};
