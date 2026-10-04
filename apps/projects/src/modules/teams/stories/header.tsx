"use client";
import { BreadCrumbs, Flex } from "ui";
import { StoryIcon } from "icons";
import { useParams, useSearchParams } from "next/navigation";
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
import type {
  SavedViewConfiguration,
  SavedViewLoadState,
} from "@/shared/story/view-configuration";
import { useTeamOptions } from "./provider";

export type SavedViewsAction = (props: {
  configuration: SavedViewConfiguration;
  onApply: (configuration: SavedViewConfiguration) => void;
  teamId: string;
  onLoadStateChange?: (state: SavedViewLoadState) => void;
}) => ReactNode;

export const Header = ({
  layout,
  setLayout,
  renderSavedViews,
  onViewLoadStateChange,
}: {
  layout: StoriesLayout;
  setLayout: (value: StoriesLayout) => void;
  renderSavedViews?: SavedViewsAction;
  onViewLoadStateChange?: (state: SavedViewLoadState) => void;
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
    viewMetadata,
  } = useTeamOptions();
  const { getTermDisplay } = useTerminology();
  const hasSavedView = Boolean(
    useSearchParams().get("view") && renderSavedViews,
  );
  const savedViewControl = renderSavedViews?.({
    configuration: {
      ...viewMetadata,
      version: 1,
      layout,
      filters,
      viewOptions,
    },
    onApply: (configuration) => {
      applyView(configuration);
      setLayout(configuration.layout);
    },
    teamId,
    onLoadStateChange: onViewLoadStateChange,
  });

  useHotkeys("v+l", () => {
    setLayout("list");
  });

  useHotkeys("v+k", () => {
    setLayout("kanban");
  });
  return (
    <HeaderContainer className="h-auto min-h-(--app-page-header-height) flex-wrap justify-between gap-x-3 gap-y-3 py-3 md:h-(--app-page-header-height) md:flex-nowrap md:py-0">
      <Flex className="min-w-0 flex-1 overflow-hidden" gap={2}>
        <MobileMenuButton />
        {hasSavedView ? (
          savedViewControl
        ) : (
          <>
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
                  icon: (
                    <StoryIcon className="h-[1.1rem] w-auto" strokeWidth={2} />
                  ),
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
          </>
        )}
      </Flex>
      {!hasSavedView && savedViewControl ? (
        <Flex align="center" className="max-w-[60%] min-w-0 md:max-w-none">
          {savedViewControl}
        </Flex>
      ) : null}
      <Flex
        align="center"
        className="w-full flex-wrap gap-2 md:w-auto md:flex-nowrap"
      >
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
