import React from "react";
import { Row, Text, Back } from "@/components/ui";
import { StoryOptionsButton } from "@/modules/stories/components";
import { useGlobalSearchParams } from "expo-router";
import { useObjective } from "@/modules/objectives/hooks/use-objectives";
import { useTerminology } from "@/hooks/use-terminology";
import type { StoriesViewOptions } from "@/types/stories-view-options";
import { truncateText } from "@/lib/utils";

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
  const { objectiveId } = useGlobalSearchParams<{
    objectiveId: string;
  }>();
  const { data: objective } = useObjective(objectiveId);
  const { getTermDisplay } = useTerminology();

  return (
    <Row className="mb-3" asContainer gap={2} justify="between" align="center">
      <Back />
      <Text fontSize="2xl" fontWeight="semibold">
        {truncateText(objective?.name ?? "", 12)} /{" "}
        <Text
          fontSize="2xl"
          color="muted"
          fontWeight="semibold"
          className="opacity-80"
        >
          {getTermDisplay("storyTerm", { variant: "plural", capitalize: true })}
        </Text>
      </Text>
      <StoryOptionsButton
        viewOptions={viewOptions}
        setViewOptions={setViewOptions}
        resetViewOptions={resetViewOptions}
      />
    </Row>
  );
};
