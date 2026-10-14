"use client";
import { BreadCrumbs, Flex, Tabs } from "ui";
import { ObjectiveIcon, StoryIcon } from "icons";
import { useParams } from "next/navigation";
import { parseAsString, useQueryState } from "nuqs";
import { useHotkeys } from "react-hotkeys-hook";
import { HeaderContainer, MobileMenuButton } from "@/components/shared";
import type { StoriesLayout } from "@/components/ui";
import {
  LayoutSwitcher,
  StoriesFilterButton,
  StoriesViewOptionsButton,
  TeamColor,
} from "@/components/ui";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useObjective } from "@/modules/objectives/hooks/use-objective";
import { useTerminology } from "@/hooks";
import { useObjectiveOptions } from "./provider";
import { HeaderActions } from "./header-actions";

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
  const { data: objective } = useObjective(objectiveId, teamId);
  const team = teams.find((team) => team.id === teamId);
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
    <HeaderContainer className="h-auto min-h-(--app-page-header-height) flex-wrap justify-between gap-x-4 gap-y-2 py-3">
      <Flex align="center" className="min-w-0 flex-1 basis-48" gap={2}>
        <MobileMenuButton />
        <BreadCrumbs
          breadCrumbs={[
            {
              name: objectiveName,
              icon: <ObjectiveIcon />,
            },
          ]}
          className="min-w-0 md:hidden [&>span]:min-w-0"
        />
        <BreadCrumbs
          breadCrumbs={[
            {
              name: team?.name || "",
              icon: <TeamColor color={team?.color} />,
              className: "hidden xl:flex",
            },
            {
              name: objectiveName,
              icon: <ObjectiveIcon />,
              className: "min-w-0",
            },
          ]}
          className="hidden min-w-0 md:flex"
        />
      </Flex>
      <Flex
        align="center"
        className="ml-auto max-w-full flex-wrap justify-end"
        gap={2}
      >
        <Tabs.List
          aria-label="Objective sections"
          className="mx-0 h-min shrink-0 flex-nowrap md:mx-0"
        >
          <Tabs.Tab
            leftIcon={<ObjectiveIcon className="size-4" />}
            value="overview"
          >
            Overview
          </Tabs.Tab>
          <Tabs.Tab leftIcon={<StoryIcon className="size-4" />} value="stories">
            {getTermDisplay("storyTerm", {
              variant: "plural",
              capitalize: true,
            })}
          </Tabs.Tab>
        </Tabs.List>
        {tab === "stories" && (
          <>
            <LayoutSwitcher layout={layout} setLayout={setLayout} />
            <StoriesFilterButton
              filters={filters}
              hiddenFields={["teamIds", "objectiveId"]}
              resetFilters={resetFilters}
              setFilters={setFilters}
            />
            <StoriesViewOptionsButton
              layout={layout}
              setViewOptions={setViewOptions}
              viewOptions={viewOptions}
            />
          </>
        )}
        {objective ? (
          <HeaderActions key={objective.id} objective={objective} />
        ) : null}
      </Flex>
    </HeaderContainer>
  );
};
