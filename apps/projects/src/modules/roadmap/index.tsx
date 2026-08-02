"use client";

import { useState } from "react";
import { BreadCrumbs, Flex, Button, Box, Text } from "ui";
import { ObjectiveIcon, PlusIcon, RoadmapIcon } from "icons";
import { HeaderContainer, MobileMenuButton } from "@/components/shared";
import { useObjectives } from "@/modules/objectives/hooks/use-objectives";
import { useLocalStorage, useTerminology, useUserRole } from "@/hooks";
import { RoadmapGanttBoard } from "@/components/ui/roadmap-gantt-board";
import { BoardSkeleton } from "@/components/ui/board-skeleton";
import { NewObjectiveDialog } from "@/components/ui";
import { RoadmapLayoutSwitcher } from "@/components/ui/roadmap-layout-switcher";
import type { ZoomLevel } from "@/components/ui/base-gantt";
import { KeyResultDetails } from "@/modules/key-results/components/key-result-details";
import type { KeyResult, Objective } from "@/modules/objectives/types";
import { RoadmapObjectiveDetails } from "./components/objective-details";
import { ObjectivesBoard } from "./components/objectives-board";
import { ObjectiveViewOptionsButton } from "./components/objective-view-options-button";
import {
  DEFAULT_OBJECTIVE_VIEW_OPTIONS,
  type ObjectiveViewOptions,
} from "./objective-board-utils";
import type { RoadmapLayoutType } from "./types";

export const RoadmapPage = () => {
  const { userRole } = useUserRole();
  const { getTermDisplay } = useTerminology();
  const [layout, setLayout] = useLocalStorage<RoadmapLayoutType>(
    "objectivesLayout",
    "gantt",
  );
  const { data: objectives = [], isPending } = useObjectives();
  const [isOpen, setIsOpen] = useState(false);
  const [zoomLevel, setZoomLevel] = useLocalStorage<ZoomLevel>(
    "roadmapZoomLevel",
    "months",
  );
  const [viewOptions, setViewOptions] = useLocalStorage<ObjectiveViewOptions>(
    "objectivesViewOptions",
    DEFAULT_OBJECTIVE_VIEW_OPTIONS,
  );
  const [selectedObjective, setSelectedObjective] = useState<Objective | null>(
    null,
  );
  const [selectedKeyResult, setSelectedKeyResult] = useState<{
    keyResult: KeyResult;
    objective: Objective;
  } | null>(null);

  const selectObjective = (objective: Objective) => {
    setSelectedKeyResult(null);
    setSelectedObjective(objective);
  };

  const selectKeyResult = (objective: Objective, keyResult: KeyResult) => {
    setSelectedObjective(null);
    setSelectedKeyResult({ keyResult, objective });
  };

  const renderContent = () => {
    if (isPending) {
      return <BoardSkeleton className="h-full" layout={layout} />;
    }

    if (objectives.length === 0) {
      return (
        <Box className="flex h-full items-center justify-center">
          <Box className="flex flex-col items-center">
            <ObjectiveIcon className="h-12 w-auto" strokeWidth={1.3} />
            <Text className="mt-8 mb-6" fontSize="3xl">
              Set your first{" "}
              {getTermDisplay("objectiveTerm", { capitalize: true })}
            </Text>
            <Text className="mb-6 max-w-md text-center" color="muted">
              Define what the workspace wants to achieve, then connect
              measurable key results and the work that moves them forward.
            </Text>
            <Flex gap={2}>
              <Button
                color="tertiary"
                disabled={userRole === "guest"}
                leftIcon={<PlusIcon className="h-[1.1rem]" />}
                onClick={() => {
                  if (userRole !== "guest") {
                    setIsOpen(true);
                  }
                }}
                size="md"
              >
                Set your first {getTermDisplay("objectiveTerm")}
              </Button>
            </Flex>
          </Box>
        </Box>
      );
    }

    switch (layout) {
      case "gantt":
        return (
          <RoadmapGanttBoard
            className="h-full"
            objectives={objectives}
            onObjectiveSelect={selectObjective}
            onZoomLevelChange={setZoomLevel}
            selectedObjectiveId={selectedObjective?.id}
            zoomLevel={zoomLevel}
          />
        );
      case "kanban":
      case "list":
        return (
          <ObjectivesBoard
            layout={layout}
            objectives={objectives}
            onCreateObjective={() => {
              if (userRole !== "guest") setIsOpen(true);
            }}
            onKeyResultSelect={selectKeyResult}
            onObjectiveSelect={selectObjective}
            setViewOptions={setViewOptions}
            viewOptions={viewOptions}
          />
        );
      default:
        return null;
    }
  };

  return (
    <>
      <HeaderContainer className="justify-between">
        <Flex gap={2}>
          <MobileMenuButton />
          <BreadCrumbs
            breadCrumbs={[
              {
                name: "Roadmap",
                icon: (
                  <RoadmapIcon className="h-[1.1rem] w-auto" strokeWidth={2} />
                ),
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
            color="invert"
            disabled={userRole === "guest"}
            leftIcon={
              <PlusIcon className="h-[1.1rem] text-current dark:text-current" />
            }
            onClick={() => {
              if (userRole !== "guest") {
                setIsOpen(true);
              }
            }}
            size="sm"
          >
            New {getTermDisplay("objectiveTerm", { capitalize: true })}
          </Button>
        </Flex>
      </HeaderContainer>

      <Box className="relative h-[calc(100dvh-4rem)] min-w-0">
        {renderContent()}
        {selectedObjective ? (
          <RoadmapObjectiveDetails
            objective={selectedObjective}
            onClose={() => {
              setSelectedObjective(null);
            }}
            onKeyResultSelect={(keyResult) => {
              selectKeyResult(selectedObjective, keyResult);
            }}
          />
        ) : null}
        {selectedKeyResult ? (
          <KeyResultDetails
            initialKeyResult={selectedKeyResult.keyResult}
            objective={selectedKeyResult.objective}
            onClose={() => {
              setSelectedKeyResult(null);
            }}
          />
        ) : null}
      </Box>
      <NewObjectiveDialog isOpen={isOpen} setIsOpen={setIsOpen} />
    </>
  );
};
