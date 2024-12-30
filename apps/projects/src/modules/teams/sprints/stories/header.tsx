"use client";
import { BreadCrumbs, Flex, Badge } from "ui";
import { SprintsIcon, StoryIcon } from "icons";
import { HeaderContainer } from "@/components/shared";
import type { StoriesLayout } from "@/components/ui";
import {
  LayoutSwitcher,
  StoriesViewOptionsButton,
  SideDetailsSwitch,
} from "@/components/ui";
import { useSprintStories } from "./provider";
import { useParams } from "next/navigation";
import { useSprints } from "@/lib/hooks/sprints";
import { useTeams } from "../../hooks/teams";

export const Header = ({
  isExpanded,
  allStories,
  setIsExpanded,
  layout,
  setLayout,
}: {
  isExpanded: boolean | null;
  allStories: number;
  setIsExpanded: (isExpanded: boolean) => void;
  layout: StoriesLayout;
  setLayout: (value: StoriesLayout) => void;
}) => {
  const { viewOptions, setViewOptions } = useSprintStories();
  const { teamId, sprintId } = useParams<{
    teamId: string;
    sprintId: string;
  }>();
  const { data: sprints = [] } = useSprints();
  const { data: teams = [] } = useTeams();

  const team = teams.find((team) => team.id === teamId)!;
  const sprint = sprints.find((sprint) => sprint.id === sprintId)!;

  return (
    <HeaderContainer className="justify-between">
      <Flex gap={2}>
        <BreadCrumbs
          breadCrumbs={[
            {
              name: team?.name,
              icon: team?.icon,
              url: `/teams/${team?.id}/stories`,
            },
            {
              name: sprint?.name,
              icon: <SprintsIcon className="h-4 w-auto" />,
              url: `/teams/${team?.id}/sprints/${sprint?.id}`,
            },
            {
              name: "Stories",
              icon: <StoryIcon className="h-[1.1rem] w-auto" strokeWidth={2} />,
            },
          ]}
        />
        <Badge className="bg-opacity-50" color="tertiary" rounded="full">
          {allStories} stories
        </Badge>
      </Flex>
      <Flex align="center" gap={2}>
        <LayoutSwitcher layout={layout} setLayout={setLayout} />
        <StoriesViewOptionsButton
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
