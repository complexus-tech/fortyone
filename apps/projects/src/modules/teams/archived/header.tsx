"use client";
import { BreadCrumbs, Flex } from "ui";
import { ArchiveIcon } from "icons";
import { useParams } from "next/navigation";
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
import { useTeamOptions } from "./provider";

export const Header = ({
  layout,
  setLayout,
}: {
  layout: StoriesLayout;
  setLayout: (value: StoriesLayout) => void;
}) => {
  const { teamId } = useParams<{
    teamId: string;
  }>();
  const { data: teams = [] } = useTeams();
  const { name, color } = teams.find((team) => team.id === teamId)!;
  const {
    viewOptions,
    setViewOptions,
    initialViewOptions,
    filters,
    resetFilters,
    setFilters,
  } = useTeamOptions();

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
              name: "Archived",
              icon: <ArchiveIcon className="h-[1.1rem]" />,
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
        <LayoutSwitcher
          layout={layout}
          options={
            viewOptions.groupBy === "none" ? ["list"] : ["list", "kanban"]
          }
          setLayout={setLayout}
        />
        <StoriesFilterButton
          filters={filters}
          resetFilters={resetFilters}
          setFilters={setFilters}
        />
        <StoriesViewOptionsButton
          groupByOptions={
            layout === "kanban"
              ? ["assignee", "priority"]
              : ["assignee", "priority"]
          }
          initialViewOptions={initialViewOptions}
          layout={layout}
          setViewOptions={setViewOptions}
          viewOptions={viewOptions}
        />
        <span className="hidden text-gray-200 dark:text-dark-100 md:inline">
          |
        </span>
        <NewStoryButton className="hidden md:flex" teamId={teamId} />
      </Flex>
    </HeaderContainer>
  );
};
