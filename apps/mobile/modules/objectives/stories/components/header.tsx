import React from "react";
import { Row, Text, Back } from "@/components/ui";
import { StoryOptionsButton } from "@/modules/stories/components";
import { useGlobalSearchParams } from "expo-router";
import { useObjective } from "@/modules/objectives/hooks/use-objectives";
import { useTerminology } from "@/hooks/use-terminology";
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
  const { objectiveId } = useGlobalSearchParams<{
    objectiveId: string;
  }>();
  const { data: objective } = useObjective(objectiveId);
  const { getTermDisplay } = useTerminology();

  // truncate objective name to 16 characters and add ellipsis if needed
  const objectiveName =
    objective?.name && objective.name.length > 16
      ? objective.name.slice(0, 16) + "..."
      : objective?.name;

  return (
    <Row className="mb-3" asContainer gap={2} justify="between" align="center">
      <Back />
      <Text fontSize="2xl" fontWeight="semibold">
        {objectiveName} /{" "}
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
