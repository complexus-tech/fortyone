import { Row, Text } from "@/components/ui";
import { DetailedStory } from "@/modules/stories/types";
import React from "react";

export const Description = ({ story }: { story: DetailedStory }) => {
  if (!story.description && !story.descriptionHTML) {
    return (
      <Row asContainer className="mb-6 mt-2">
        <Text color="muted" className="pl-0.5" fontWeight="medium">
          Enter description
        </Text>
      </Row>
    );
  }
  return (
    <Row asContainer className="mb-5 mt-2">
      <Text color="muted" className="pl-0.5" fontWeight="medium">
        {story.description}
      </Text>
    </Row>
  );
};
