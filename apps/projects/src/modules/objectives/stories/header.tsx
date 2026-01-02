"use client";
import { BreadCrumbs, Flex } from "ui";
import { ObjectiveIcon, StoryIcon } from "icons";
import { useParams } from "next/navigation";
import { parseAsString, useQueryState } from "nuqs";
import { useHotkeys } from "react-hotkeys-hook";
import { HeaderContainer, MobileMenuButton } from "@/components/shared";
import type { StoriesLayout } from "@/components/ui";
import {
  LayoutSwitcher,
  NewStoryButton,
  StoriesFilterButton,
  StoriesViewOptionsButton,
  TeamColor,
} from "@/components/ui";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useTeamObjectives } from "@/modules/objectives/hooks/use-objectives";
import { useObjective } from "@/modules/objectives/hooks/use-objective";
import { useTerminology } from "@/hooks";
import { useObjectiveOptions } from "./provider";

export const Header = ({
  layout,
  setLayout,
}: {
  layout: StoriesLayout;
  setLayout: (value: StoriesLayout) => void;
}) => {
  const { teamId, objectiveId } = useParams<{
    teamId: string;
    objectiveId: string;
  }>();
  const [tab] = useQueryState("tab", parseAsString.withDefault("overview"));
  const { getTermDisplay } = useTerminology();
  const { data: teams = [] } = useTeams();
  const { data: objectives = [] } = useTeamObjectives(teamId);
  const { data: objective } = useObjective(objectiveId, teamId);
  const { name: teamName, color: teamColor } = teams.find(
    (team) => team.id === teamId,
  )!;
  const objectiveName = objective?.name || "";
  const { viewOptions, setViewOptions, filters, setFilters, resetFilters } =
    useObjectiveOptions();

  useHotkeys("v+l", () => {
    setLayout("list");
  });

  useHotkeys("v+k", () => {
    setLayout("kanban");
  });
  return (
    <HeaderContainer className="justify-between">
      <Flex className="mr-2" gap={2}>
        <MobileMenuButton />
        <BreadCrumbs
          breadCrumbs={[
            {
              name: objectiveName,
              icon: <ObjectiveIcon />,
            },
          ]}
          className="md:hidden"
        />
        <BreadCrumbs
          breadCrumbs={[
            {
              name: teamName,
              icon: <TeamColor color={teamColor} />,
            },
            {
              name: objectiveName,
              icon: <ObjectiveIcon />,
            },
            {
              name:
                tab === "stories"
                  ? getTermDisplay("storyTerm", {
                      variant: "plural",
                      capitalize: true,
                    })
                  : "Overview",
              icon:
                tab === "stories" ? (
                  <StoryIcon className="h-[1.1rem]" />
                ) : (
                  <ObjectiveIcon />
                ),
            },
          ]}
          className="hidden md:flex"
        />
      </Flex>
      <Flex align="center" gap={2}>
        {tab === "stories" && (
          <>
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
            <span className="hidden text-text-secondary md:inline">
              |
            </span>
          </>
        )}
        <NewStoryButton
          className="hidden md:flex"
          objectiveId={objectiveId}
          teamId={teamId}
        />
      </Flex>
    </HeaderContainer>
  );
};
