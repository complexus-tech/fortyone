import React from "react";
import { Row, Text } from "@/components/ui";
import { StoryOptionsButton } from "@/modules/stories/components";
import type { StoriesViewOptions } from "@/types/stories-view-options";

type HeaderProps = {
  viewOptions: StoriesViewOptions;
  setViewOptions: (options: Partial<StoriesViewOptions>) => void;
  resetViewOptions: () => void;
};

export const Header = ({
  viewOptions,
  setViewOptions,
  resetViewOptions,
}: HeaderProps) => {
  return (
    <Row className="mb-2" asContainer justify="between" align="center">
      <Text fontSize="2xl" fontWeight="semibold">
        My Work
      </Text>
      <StoryOptionsButton
        viewOptions={viewOptions}
        setViewOptions={setViewOptions}
        resetViewOptions={resetViewOptions}
      />
    </Row>
  );
};
