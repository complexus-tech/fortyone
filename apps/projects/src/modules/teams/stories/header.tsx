"use client";
import { Badge, BreadCrumbs, Flex } from "ui";
import { StoryIcon } from "icons";
import { useParams } from "next/navigation";
import { HeaderContainer } from "@/components/shared";
import type { StoriesLayout } from "@/components/ui";
import {
  LayoutSwitcher,
  NewStoryButton,
  SideDetailsSwitch,
  StoriesFilterButton,
  StoriesViewOptionsButton,
  TeamColor,
} from "@/components/ui";
import { useTeamStories } from "@/modules/stories/hooks/team-stories";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useTeamOptions } from "./provider";

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
  const { teamId } = useParams<{
    teamId: string;
  }>();
  const { data: stories = [] } = useTeamStories(teamId);
  const { data: teams = [] } = useTeams();
  const { name, color } = teams.find((team) => team.id === teamId)!;
  const { viewOptions, setViewOptions, filters, setFilters, resetFilters } =
    useTeamOptions();
  return (
    <HeaderContainer className="justify-between">
      <Flex gap={2}>
        <BreadCrumbs
          breadCrumbs={[
            {
              name,
              icon: <TeamColor color={color} />,
            },
            {
              name: "Stories",
              icon: <StoryIcon className="h-[1.1rem] w-auto" strokeWidth={2} />,
            },
          ]}
        />
        <Badge className="bg-opacity-50" color="tertiary" rounded="full">
          {stories.length} stories
        </Badge>
      </Flex>
      <Flex align="center" gap={2}>
        <LayoutSwitcher layout={layout} setLayout={setLayout} />
        <StoriesFilterButton
          filters={filters}
          resetFilters={resetFilters}
          setFilters={setFilters}
        />
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
        <NewStoryButton teamId={teamId} />
      </Flex>
    </HeaderContainer>
  );
};
