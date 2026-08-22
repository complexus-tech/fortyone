"use client";
import { BreadCrumbs, Flex } from "ui";
import { useParams } from "next/navigation";
import { useHotkeys } from "react-hotkeys-hook";
import { HeaderContainer, MobileMenuButton } from "@/components/shared";
import type { StoriesLayout } from "@/components/ui";
import {
  LayoutSwitcher,
  StoriesFilterButton,
  StoriesViewOptionsButton,
} from "@/components/ui";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useTerminology } from "@/hooks";
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
  const selectedTeam = teams.find((team) => team.id === teamId);
  const name = selectedTeam?.name ?? "Team";
  const { viewOptions, setViewOptions, filters, resetFilters, setFilters } =
    useTeamOptions();
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
            },
            {
              name: getTermDisplay("storyTerm", {
                variant: "plural",
                capitalize: true,
              }),
            },
          ]}
          className="hidden md:flex"
        />
        <BreadCrumbs
          breadCrumbs={[
            {
              name,
            },
          ]}
          className="md:hidden"
        />
      </Flex>
      <Flex align="center" gap={2}>
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
