"use client";
import { BreadCrumbs, Button, Flex } from "ui";
import { ObjectiveIcon, PlusIcon } from "icons";
import { useParams } from "next/navigation";
import {
  HeaderContainer,
  MobileMenuButton,
  useAppCommandAction,
} from "@/components/shared";
import { RoadmapLayoutSwitcher } from "@/components/ui/roadmap-layout-switcher";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useUserRole, useTerminology, useMediaQuery } from "@/hooks";
import { ObjectiveViewOptionsButton } from "@/modules/roadmap/components/objective-view-options-button";
import type { ObjectiveViewOptions } from "@/modules/roadmap/objective-board-utils";
import { TeamColor } from "@/components/ui";
import {
  getRoadmapLayoutLabel,
  type RoadmapLayoutType,
} from "@/modules/roadmap/types";

export const TeamObjectivesHeader = ({
  layout,
  onCreateObjective,
  setLayout,
  setViewOptions,
  viewOptions,
}: {
  layout: RoadmapLayoutType;
  onCreateObjective: () => void;
  setLayout: (layout: RoadmapLayoutType) => void;
  setViewOptions: (viewOptions: ObjectiveViewOptions) => void;
  viewOptions: ObjectiveViewOptions;
}) => {
  const { teamId } = useParams<{ teamId: string }>();
  const isMobile = useMediaQuery("(max-width: 768px)");
  const { data: teams = [] } = useTeams();
  const selectedTeam = teams.find((team) => team.id === teamId);
  const name = selectedTeam?.name ?? "Team";
  const { userRole } = useUserRole();
  const { getTermDisplay } = useTerminology();

  useAppCommandAction({
    disabled: userRole === "guest",
    id: `team:${teamId}:create-objective`,
    label: `Create ${getTermDisplay("objectiveTerm")}`,
    onSelect: onCreateObjective,
  });

  return (
    <HeaderContainer className="justify-between">
      <Flex gap={2}>
        <MobileMenuButton />
        <BreadCrumbs
          breadCrumbs={[
            {
              name: isMobile ? "" : name,
              icon: <TeamColor color={selectedTeam?.color} />,
            },
            {
              name: getTermDisplay("objectiveTerm", {
                variant: "plural",
                capitalize: true,
              }),
              icon: <ObjectiveIcon strokeWidth={2} />,
            },
            {
              name: getRoadmapLayoutLabel(layout),
            },
          ]}
        />
      </Flex>
      <Flex align="center" gap={2}>
        <RoadmapLayoutSwitcher
          className="hidden md:flex"
          layout={layout}
          setLayout={setLayout}
        />
        {layout !== "gantt" ? (
          <ObjectiveViewOptionsButton
            setViewOptions={setViewOptions}
            viewOptions={viewOptions}
          />
        ) : null}
        <Button
          className="md:hidden"
          color="primary"
          disabled={userRole === "guest"}
          leftIcon={
            <PlusIcon className="h-[1.1rem] text-current dark:text-current" />
          }
          onClick={() => {
            if (userRole !== "guest") {
              onCreateObjective();
            }
          }}
          size="sm"
        >
          New {getTermDisplay("objectiveTerm", { capitalize: true })}
        </Button>
      </Flex>
    </HeaderContainer>
  );
};
