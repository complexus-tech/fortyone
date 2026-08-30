"use client";

import { useParams } from "next/navigation";
import { ObjectiveIcon, PlusIcon } from "icons";
import { BreadCrumbs, Button, Flex } from "ui";
import { useAppCommandAction } from "@/components/shared/app-command-action-context";
import { HeaderContainer } from "@/components/shared/header-container";
import { MobileMenuButton } from "@/components/shared/mobile-menu";
import { TeamColor } from "@/components/ui/team-color";
import { RoadmapLayoutSwitcher } from "@/components/ui/roadmap-layout-switcher";
import { useMediaQuery } from "@/hooks/media";
import { useUserRole } from "@/hooks/role";
import { useTerminology } from "@/hooks/use-terminology-display";
import { useTeams } from "@/modules/teams/public/client";
import type { ObjectiveViewOptions } from "../objective-board-utils";
import { getRoadmapLayoutLabel, type RoadmapLayoutType } from "../types";
import { ObjectiveViewOptionsButton } from "./objective-view-options-button";

type TeamRoadmapHeaderProps = {
  layout: RoadmapLayoutType;
  onCreateObjective: () => void;
  setLayout: (layout: RoadmapLayoutType) => void;
  setViewOptions: (viewOptions: ObjectiveViewOptions) => void;
  viewOptions: ObjectiveViewOptions;
};

export const TeamRoadmapHeader = ({
  layout,
  onCreateObjective,
  setLayout,
  setViewOptions,
  viewOptions,
}: TeamRoadmapHeaderProps) => {
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
            if (userRole !== "guest") onCreateObjective();
          }}
          size="sm"
        >
          New {getTermDisplay("objectiveTerm", { capitalize: true })}
        </Button>
      </Flex>
    </HeaderContainer>
  );
};
