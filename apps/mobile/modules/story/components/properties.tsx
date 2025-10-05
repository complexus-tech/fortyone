import { useStory } from "@/modules/stories/hooks";
import React from "react";
import { useGlobalSearchParams } from "expo-router";
import { Row, Badge, Text } from "@/components/ui";

export const Properties = () => {
  const { storyId } = useGlobalSearchParams<{ storyId: string }>();
  const { data: story } = useStory(storyId);
  return (
    <Row wrap gap={2} asContainer className="my-4">
      <Badge color="tertiary" rounded="xl">
        <Text>Test</Text>
      </Badge>
    </Row>
  );
};
