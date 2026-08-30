"use client";

import { useState } from "react";
import { useParams } from "next/navigation";
import { CrownIcon, PlusIcon } from "icons";
import { Box, Button, Text } from "ui";
import type { ZoomLevel } from "@/components/ui/base-gantt";
import { FeatureGuard } from "@/components/ui/feature-guard";
import { RoadmapEmptyIllustration } from "@/components/ui/illustrations/empty-state-illustrations";
import { NewObjectiveDialog } from "@/components/ui/new-objective";
import { useLocalStorage } from "@/hooks/local-storage";
import { useUserRole } from "@/hooks/role";
import { useTerminology } from "@/hooks/use-terminology-display";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useTeamObjectives } from "@/modules/objectives/public/client";
import {
  DEFAULT_OBJECTIVE_VIEW_OPTIONS,
  type ObjectiveViewOptions,
} from "./objective-board-utils";
import { ObjectiveViews } from "./components/objective-views";
import { TeamRoadmapHeader } from "./components/team-roadmap-header";
import { useRoadmapLayout } from "./use-roadmap-layout";

const TeamRoadmapUpgradeGuard = () => {
  const { userRole } = useUserRole();
  const { getTermDisplay } = useTerminology();
  const { withWorkspace } = useWorkspacePath();

  return (
    <Box className="flex h-[80%] items-center justify-center">
      <Box className="flex flex-col items-center">
        <CrownIcon className="text-warning h-12" strokeWidth={1.3} />
        <Text className="mt-8 mb-6" fontSize="3xl">
          Upgrade your plan
        </Text>
        <Text className="mb-6 max-w-md text-center" color="muted">
          {userRole === "admin" ? "Upgrade " : "Ask your admin to upgrade "}
          your plan to create{" "}
          {getTermDisplay("objectiveTerm", { variant: "plural" })}, unlimited{" "}
          {getTermDisplay("objectiveTerm", { variant: "plural" })}, and unlock
          more premium features.
        </Text>
        {userRole === "admin" ? (
          <Button
            color="warning"
            href={withWorkspace("/settings/workspace/billing")}
          >
            Upgrade now
          </Button>
        ) : null}
      </Box>
    </Box>
  );
};

export const TeamRoadmapPage = () => {
  const { teamId } = useParams<{ teamId: string }>();
  const { data: objectives = [], isPending } = useTeamObjectives(teamId);
  const { userRole } = useUserRole();
  const { getTermDisplay } = useTerminology();
  const [isOpen, setIsOpen] = useState(false);
  const { layout, setLayout } = useRoadmapLayout();
  const [zoomLevel, setZoomLevel] = useLocalStorage<ZoomLevel>(
    "roadmapZoomLevel",
    "months",
  );
  const [viewOptions, setViewOptions] = useLocalStorage<ObjectiveViewOptions>(
    "objectivesViewOptions",
    DEFAULT_OBJECTIVE_VIEW_OPTIONS,
  );
  const openNewObjectiveDialog = () => {
    if (userRole !== "guest") setIsOpen(true);
  };
  const emptyState = (
    <Box className="flex h-full items-center justify-center">
      <Box className="flex flex-col items-center">
        <RoadmapEmptyIllustration />
        <Text className="mt-8 mb-6" fontSize="3xl">
          No {getTermDisplay("objectiveTerm", { variant: "plural" })} found
        </Text>
        <Text className="mb-6 max-w-md text-center" color="muted">
          This team doesn&apos;t have any{" "}
          {getTermDisplay("objectiveTerm", { variant: "plural" })} yet. Create a
          new {getTermDisplay("objectiveTerm")} to get started.
        </Text>
        <Button
          color="primary"
          disabled={userRole === "guest"}
          leftIcon={<PlusIcon className="h-[1.1rem]" />}
          onClick={openNewObjectiveDialog}
          size="md"
        >
          Create new {getTermDisplay("objectiveTerm")}
        </Button>
      </Box>
    </Box>
  );

  return (
    <>
      <TeamRoadmapHeader
        layout={layout}
        onCreateObjective={openNewObjectiveDialog}
        setLayout={setLayout}
        setViewOptions={setViewOptions}
        viewOptions={viewOptions}
      />
      <FeatureGuard fallback={<TeamRoadmapUpgradeGuard />} feature="objective">
        <Box className="h-[calc(100%-3.6rem)] min-w-0">
          <ObjectiveViews
            emptyState={emptyState}
            isPending={isPending}
            key={teamId}
            layout={layout}
            objectives={objectives}
            onCreateObjective={openNewObjectiveDialog}
            onZoomLevelChange={setZoomLevel}
            setViewOptions={setViewOptions}
            viewOptions={viewOptions}
            zoomLevel={zoomLevel}
          />
        </Box>
      </FeatureGuard>
      <NewObjectiveDialog
        isOpen={isOpen}
        setIsOpen={setIsOpen}
        teamId={teamId}
      />
    </>
  );
};
