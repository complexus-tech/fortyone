"use client";
import { Box, BreadCrumbs, Button, Flex, Tooltip } from "ui";
import { AiIcon, SprintsIcon, StoryIcon } from "icons";
import { useParams } from "next/navigation";
import { format } from "date-fns";
import { useHotkeys } from "react-hotkeys-hook";
import { HeaderContainer, MobileMenuButton } from "@/components/shared";
import type { StoriesLayout } from "@/components/ui";
import {
  LayoutSwitcher,
  StoriesViewOptionsButton,
  SideDetailsSwitch,
  TeamColor,
} from "@/components/ui";
import { useTerminology, useUserRole } from "@/hooks";
import { useChatContext } from "@/context/chat-context";
import { useTeams } from "../../teams/hooks/teams";
import { useSprint } from "../hooks/sprint-details";
import { useSprintOptions } from "./provider";

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
  const { getTermDisplay } = useTerminology();
  const { viewOptions, setViewOptions } = useSprintOptions();
  const { teamId, sprintId } = useParams<{
    teamId: string;
    sprintId: string;
  }>();
  const { data: teams = [] } = useTeams();
  const { data: sprint } = useSprint(sprintId);
  const { userRole } = useUserRole();
  const { openChat } = useChatContext();

  useHotkeys("v+l", () => {
    setLayout("list");
  });

  useHotkeys("v+k", () => {
    setLayout("kanban");
  });

  if (!sprint) {
    return null;
  }

  const team = teams.find((team) => team.id === teamId)!;
  const startDate = format(new Date(sprint.startDate), "MMM d");
  const endDate = format(new Date(sprint.endDate), "MMM d");
  const sprintName = `${sprint.name} (${startDate} - ${endDate})`;

  return (
    <HeaderContainer className="justify-between gap-4">
      <Flex align="center" gap={2}>
        <MobileMenuButton />
        <BreadCrumbs
          breadCrumbs={[
            {
              name: team.name,
              icon: <TeamColor color={team.color} />,
              url: `/teams/${team.id}/stories`,
            },
            {
              name: sprintName,
              icon: <SprintsIcon className="h-4 w-auto" />,
              url: `/teams/${team.id}/sprints/${sprint.id}`,
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
              name: sprintName,
              icon: <SprintsIcon className="h-4 w-auto" />,
              url: `/teams/${team.id}/sprints/${sprint.id}`,
            },
          ]}
          className="md:hidden"
        />
      </Flex>
      <Flex align="center" gap={2}>
        {userRole !== "guest" && (
          <Tooltip
            className="max-w-60"
            title="Intelligently suggest work items from backlog to fill this sprint"
          >
            <Button
              color="tertiary"
              leftIcon={<AiIcon className="text-primary dark:text-primary" />}
              onClick={() => {
                openChat(
                  `Suggest stories from "${team.name}" team backlog for sprint "${sprint.name}"`,
                );
              }}
              size="sm"
              variant="naked"
            >
              Smart Fill
            </Button>
          </Tooltip>
        )}
        <LayoutSwitcher
          layout={layout}
          options={["list", "kanban"]}
          setLayout={setLayout}
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
      </Flex>
    </HeaderContainer>
  );
};
