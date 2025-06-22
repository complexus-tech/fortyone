"use client";
import { Box, BreadCrumbs, Flex } from "ui";
import { StoryIcon } from "icons";
import { useParams } from "next/navigation";
import { useHotkeys } from "react-hotkeys-hook";
import { HeaderContainer, MobileMenuButton } from "@/components/shared";
import type { StoriesLayout } from "@/components/ui";
import {
  LayoutSwitcher,
  NewStoryButton,
  SideDetailsSwitch,
  StoriesFilterButton,
  StoriesViewOptionsButton,
  TeamColor,
} from "@/components/ui";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useTerminology } from "@/hooks";
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
  const { data: teams = [] } = useTeams();
  const { name, color } = teams.find((team) => team.id === teamId)!;
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
        <span className="hidden text-gray-200 dark:text-dark-100 md:inline">
          |
        </span>
        <Box className="hidden md:block">
          <SideDetailsSwitch
            isExpanded={isExpanded}
            setIsExpanded={setIsExpanded}
          />
        </Box>
        <NewStoryButton className="hidden md:flex" teamId={teamId} />
      </Flex>
    </HeaderContainer>
  );
};
