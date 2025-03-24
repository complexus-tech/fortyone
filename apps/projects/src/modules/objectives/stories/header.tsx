"use client";
import { BreadCrumbs, Flex } from "ui";
import { ObjectiveIcon, StoryIcon } from "icons";
import { useParams } from "next/navigation";
import { parseAsString, useQueryState } from "nuqs";
import { useHotkeys } from "react-hotkeys-hook";
import { HeaderContainer } from "@/components/shared";
import type { StoriesLayout } from "@/components/ui";
import {
  LayoutSwitcher,
  NewStoryButton,
  StoriesFilterButton,
  StoriesViewOptionsButton,
  TeamColor,
} from "@/components/ui";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useObjectives } from "@/modules/objectives/hooks/use-objectives";
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
  const { data: objectives = [] } = useObjectives();
  const { name: teamName, color: teamColor } = teams.find(
    (team) => team.id === teamId,
  )!;
  const { name: objectiveName } = objectives.find(
    (objective) => objective.id === objectiveId,
  )!;
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
      <Flex gap={2}>
        <BreadCrumbs
          breadCrumbs={[
            {
              name: teamName,
              icon: <TeamColor color={teamColor} />,
            },
            {
              name: objectiveName,
              icon: <ObjectiveIcon className="h-[1.05rem]" />,
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
                  <ObjectiveIcon className="h-[1.05rem]" />
                ),
            },
          ]}
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
            <span className="text-gray-200 dark:text-dark-100">|</span>
          </>
        )}
        <NewStoryButton objectiveId={objectiveId} teamId={teamId} />
      </Flex>
    </HeaderContainer>
  );
};
