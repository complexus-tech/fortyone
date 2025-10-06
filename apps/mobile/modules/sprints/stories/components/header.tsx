import React from "react";
import { Row, Text, Back } from "@/components/ui";
import { StoryOptionsButton } from "@/modules/stories/components";
import { useGlobalSearchParams } from "expo-router";
import { useSprint } from "@/modules/sprints/hooks/use-sprints";
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
  const { sprintId } = useGlobalSearchParams<{ sprintId: string }>();
  const { data: sprint } = useSprint(sprintId);
  const { getTermDisplay } = useTerminology();
  // truncate sprint name to 16 characters and add ellipsis if needed
  const sprintName =
    sprint?.name && sprint.name.length > 16
      ? sprint.name.slice(0, 16) + "..."
      : sprint?.name;

  return (
    <Row className="mb-3" asContainer justify="between" align="center">
      <Back />
      <Text fontSize="2xl" fontWeight="semibold">
        {sprintName} /{" "}
        <Text
          fontSize="2xl"
          color="muted"
          fontWeight="semibold"
          className="opacity-80"
        >
          {getTermDisplay("storyTerm", {
            variant: "plural",
            capitalize: true,
          })}
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
