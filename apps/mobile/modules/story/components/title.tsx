import React from "react";
import { Text, Badge, Col } from "@/components/ui";
import { DetailedStory } from "@/modules/stories/types";
import { differenceInDays, addDays } from "date-fns";

export const Title = ({ story }: { story: DetailedStory }) => {
  const isDeleted = story.deletedAt !== null;
  const getDaysLeft = () => {
    if (!story.deletedAt) return 0;
    const daysLeft = differenceInDays(
      addDays(new Date(story.deletedAt!), 30),
      new Date(story.deletedAt)
    );
    return daysLeft;
  };

  return (
    <Col asContainer>
      {isDeleted && (
        <Badge color="tertiary">
          <Text>{getDaysLeft()} days left in bin</Text>
        </Badge>
      )}
      <Text fontSize="2xl" fontWeight="semibold">
        {story.title}
      </Text>
    </Col>
  );
};
