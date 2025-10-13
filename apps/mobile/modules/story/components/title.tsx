import React from "react";
import { Text, Badge, Col } from "@/components/ui";
import { DetailedStory } from "@/modules/stories/types";
import { differenceInDays } from "date-fns";

export const Title = ({ story }: { story: DetailedStory }) => {
  const isDeleted = story.deletedAt !== null;
  const getDaysLeft = () => {
    if (!story.deletedAt) return 0;
    const daysLeft = differenceInDays(new Date(story.deletedAt), new Date());
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
