"use client";
import { BreadCrumbs, Button, Flex } from "ui";
import { PlusIcon } from "icons";
import { useParams } from "next/navigation";
import { HeaderContainer, MobileMenuButton } from "@/components/shared";
import { RoadmapLayoutSwitcher } from "@/components/ui/roadmap-layout-switcher";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useUserRole, useTerminology, useMediaQuery } from "@/hooks";
import { ObjectiveViewOptionsButton } from "@/modules/roadmap/components/objective-view-options-button";
import type { ObjectiveViewOptions } from "@/modules/roadmap/objective-board-utils";
import type { RoadmapLayoutType } from "@/modules/roadmap/types";

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

  return (
    <HeaderContainer className="justify-between">
      <Flex gap={2}>
        <MobileMenuButton />
        <BreadCrumbs
          breadCrumbs={[
            {
              name: isMobile ? "" : name,
            },
            {
              name: getTermDisplay("objectiveTerm", {
                variant: "plural",
                capitalize: true,
              }),
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
          color="tertiary"
          disabled={userRole === "guest"}
          leftIcon={<PlusIcon className="h-[1.1rem]" />}
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
