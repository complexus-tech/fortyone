"use client";
import { Badge, BreadCrumbs, Flex } from "ui";
import { ObjectiveIcon, StoryIcon } from "icons";
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
import { useObjectiveOptions } from "./provider";
import { useParams } from "next/navigation";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useObjectives } from "@/modules/objectives/hooks/use-objectives";
import { useObjectiveStories } from "@/modules/stories/hooks/objective-stories";

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
  const { teamId, objectiveId } = useParams<{
    teamId: string;
    objectiveId: string;
  }>();
  const { data: stories = [] } = useObjectiveStories(objectiveId);
  const { data: teams = [] } = useTeams();
  const { data: objectives = [] } = useObjectives();
  const { name: teamName, color: teamColor } = teams.find(
    (team) => team.id === teamId,
  )!!;
  const { name: objectiveName } = objectives.find(
    (objective) => objective.id === objectiveId,
  )!!;
  const { viewOptions, setViewOptions, filters, setFilters, resetFilters } =
    useObjectiveOptions();
  return (
    <HeaderContainer className="justify-between">
      <Flex gap={2}>
        <BreadCrumbs
          breadCrumbs={[
            {
              name: teamName,
              icon: <TeamColor color={teamColor} />,
            },
            {
              name: objectiveName,
              icon: (
                <ObjectiveIcon className="h-[1.1rem] w-auto" strokeWidth={2} />
              ),
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
          setViewOptions={setViewOptions}
          viewOptions={viewOptions}
        />
        <NewStoryButton teamId={teamId} />
        <span className="text-gray-200 dark:text-dark-100">|</span>
        <SideDetailsSwitch
          isExpanded={isExpanded}
          setIsExpanded={setIsExpanded}
        />
      </Flex>
    </HeaderContainer>
  );
};
