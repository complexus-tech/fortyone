"use client";
import { BreadCrumbs, Flex } from "ui";
import { StoryIcon } from "icons";
import { useParams } from "next/navigation";
import { useHotkeys } from "react-hotkeys-hook";
import type { ReactNode } from "react";
import { HeaderContainer, MobileMenuButton } from "@/components/shared";
import type { StoriesLayout } from "@/components/ui";
import {
  LayoutSwitcher,
  StoriesFilterButton,
  StoriesViewOptionsButton,
  TeamColor,
} from "@/components/ui";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useTerminology } from "@/hooks";
import type { SavedViewConfiguration } from "@/shared/story/view-configuration";
import { useTeamOptions } from "./provider";

export type SavedViewsAction = (props: {
  configuration: SavedViewConfiguration;
  onApply: (configuration: SavedViewConfiguration) => void;
  teamId: string;
}) => ReactNode;

export const Header = ({
  layout,
  setLayout,
  renderSavedViews,
}: {
  layout: StoriesLayout;
  setLayout: (value: StoriesLayout) => void;
  renderSavedViews?: SavedViewsAction;
}) => {
  const { teamId } = useParams<{
    teamId: string;
  }>();
  const { data: teams = [] } = useTeams();
  const selectedTeam = teams.find((team) => team.id === teamId);
  const name = selectedTeam?.name ?? "Team";
  const color = selectedTeam?.color;
  const {
    viewOptions,
    setViewOptions,
    filters,
    resetFilters,
    setFilters,
    applyView,
  } = useTeamOptions();
  const { getTermDisplay } = useTerminology();

  useHotkeys("v+l", () => {
    setLayout("list");
  });

  useHotkeys("v+k", () => {
    setLayout("kanban");
  });
  return (
    <HeaderContainer className="justify-between">
      <Flex gap={2}>
        <MobileMenuButton />
        <BreadCrumbs
          breadCrumbs={[
            {
              name,
              icon: <TeamColor color={color} />,
            },
            {
              name: getTermDisplay("storyTerm", {
                variant: "plural",
                capitalize: true,
              }),
              icon: <StoryIcon className="h-[1.1rem] w-auto" strokeWidth={2} />,
            },
          ]}
          className="hidden md:flex"
        />
        <BreadCrumbs
          breadCrumbs={[
            {
              name,
              icon: <TeamColor color={color} />,
            },
          ]}
          className="md:hidden"
        />
      </Flex>
      <Flex align="center" gap={2}>
        {renderSavedViews?.({
          configuration: { version: 1, layout, filters, viewOptions },
          onApply: (configuration) => {
            applyView(configuration);
            setLayout(configuration.layout);
          },
          teamId,
        })}
        <LayoutSwitcher layout={layout} setLayout={setLayout} />
        <StoriesFilterButton
          filters={filters}
          resetFilters={resetFilters}
          setFilters={setFilters}
        />
        <StoriesViewOptionsButton
          groupByOptions={["status", "assignee", "priority"]}
          layout={layout}
          setViewOptions={setViewOptions}
          viewOptions={viewOptions}
        />
      </Flex>
    </HeaderContainer>
  );
};
