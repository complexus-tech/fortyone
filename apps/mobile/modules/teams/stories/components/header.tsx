import React from "react";
import { Row, Text, Back } from "@/components/ui";
import { StoryOptionsButton } from "@/modules/stories/components";
import { useGlobalSearchParams } from "expo-router";
import { useTeams } from "@/modules/teams/hooks/use-teams";
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
  const { teamId } = useGlobalSearchParams<{ teamId: string }>();
  const { data: teams = [] } = useTeams();
  const team = teams.find((team) => team.id === teamId)!;
  const { getTermDisplay } = useTerminology();

  return (
    <Row className="mb-4" asContainer align="center" justify="between">
      <Back />
      <Text fontSize="2xl" fontWeight="semibold">
        {truncateText(team?.name ?? "", 12)} /{" "}
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
