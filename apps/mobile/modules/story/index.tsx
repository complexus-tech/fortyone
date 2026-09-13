import React from "react";
import { Properties } from "./components/properties";
import { useLocalSearchParams } from "expo-router";
import { Button, Col, Text } from "@/components/ui";
import { useStory } from "../stories/hooks/use-story";
import { Activity } from "./components/activity";
import { Description } from "./components/description";
import { Title } from "./components/title";
import { StorySkeleton } from "./components/story-skeleton";

export const Story = () => {
  const { storyId } = useLocalSearchParams<{ storyId: string }>();
  const {
    data: story,
    isPending,
    isError,
    error,
    refetch,
    isRefetching,
  } = useStory(storyId);

  if (isPending) {
    return <StorySkeleton />;
  }

  if (isError || !story) {
    return (
      <Col asContainer gap={3}>
        <Text fontSize="lg">Could not open this task</Text>
        <Text color="muted">
          {error?.message ??
            "This task may have been removed or you may no longer have access."}
        </Text>
        <Button loading={isRefetching} onPress={() => void refetch()}>
          Try again
        </Button>
      </Col>
    );
  }

  return (
    <Activity key={story.id} story={story}>
      <Title story={story} />
      <Properties story={story} />
      <Description story={story} />
    </Activity>
  );
};
