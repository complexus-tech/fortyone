"use client";
import { BreadCrumbs, Flex } from "ui";
import { SprintsIcon, StoryIcon } from "icons";
import { useParams } from "next/navigation";
import { format } from "date-fns";
import { HeaderContainer } from "@/components/shared";
import type { StoriesLayout } from "@/components/ui";
import {
  LayoutSwitcher,
  StoriesViewOptionsButton,
  SideDetailsSwitch,
  TeamColor,
} from "@/components/ui";
import { useTerminology } from "@/hooks";
import { useTeams } from "../../teams/hooks/teams";
import { useSprints } from "../hooks/sprints";
import { useSprintOptions } from "./provider";

export const Header = ({
  isExpanded,
  setIsExpanded,
  layout,
  setLayout,
}: {
  isExpanded: boolean | null;
  setIsExpanded: (isExpanded: boolean) => void;
  layout: StoriesLayout;
  setLayout: (value: StoriesLayout) => void;
}) => {
  const { getTermDisplay } = useTerminology();
  const { viewOptions, setViewOptions } = useSprintOptions();
  const { teamId, sprintId } = useParams<{
    teamId: string;
    sprintId: string;
  }>();
  const { data: sprints = [] } = useSprints();
  const { data: teams = [] } = useTeams();

  const team = teams.find((team) => team.id === teamId)!;
  const sprint = sprints.find((sprint) => sprint.id === sprintId)!;
  const startDate = format(new Date(sprint.startDate), "MMM d");
  const endDate = format(new Date(sprint.endDate), "MMM d");
  const sprintName = `${sprint.name} (${startDate} - ${endDate})`;

  return (
    <HeaderContainer className="justify-between">
      <Flex gap={2}>
        <BreadCrumbs
          breadCrumbs={[
            {
              name: team.name,
              icon: <TeamColor color={team.color} />,
              url: `/teams/${team.id}/stories`,
            },
            {
              name: sprintName,
              icon: <SprintsIcon className="h-4 w-auto" />,
              url: `/teams/${team.id}/sprints/${sprint.id}`,
            },
            {
              name: getTermDisplay("storyTerm", {
                variant: "plural",
                capitalize: true,
              }),
              icon: <StoryIcon className="h-[1.1rem] w-auto" strokeWidth={2} />,
            },
          ]}
        />
      </Flex>
      <Flex align="center" gap={2}>
        <LayoutSwitcher layout={layout} setLayout={setLayout} />
        <StoriesViewOptionsButton
          layout={layout}
          setViewOptions={setViewOptions}
          viewOptions={viewOptions}
        />
        <span className="text-gray-200 dark:text-dark-100">|</span>
        <SideDetailsSwitch
          isExpanded={isExpanded}
          setIsExpanded={setIsExpanded}
        />
      </Flex>
    </HeaderContainer>
  );
};
