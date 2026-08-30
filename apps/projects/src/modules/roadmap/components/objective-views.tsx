"use client";

import { useState, type ReactNode } from "react";
import { Box } from "ui";
import { BoardSkeleton } from "@/components/ui/board-skeleton";
import { RoadmapGanttBoard } from "@/components/ui/roadmap-gantt-board";
import type { ZoomLevel } from "@/components/ui/base-gantt";
import { KeyResultDetails } from "@/modules/key-results/components/key-result-details";
import type { KeyResult, Objective } from "@/modules/objectives/types";
import type { RoadmapLayoutType } from "../types";
import type { ObjectiveViewOptions } from "../objective-board-utils";
import { RoadmapObjectiveDetails } from "./objective-details";
import { ObjectivesBoard } from "./objectives-board";

export const ObjectiveViews = ({
  emptyState,
  isPending,
  layout,
  objectives,
  onCreateObjective,
  onZoomLevelChange,
  setViewOptions,
  viewOptions,
  zoomLevel,
}: {
  emptyState: ReactNode;
  isPending: boolean;
  layout: RoadmapLayoutType;
  objectives: Objective[];
  onCreateObjective: () => void;
  onZoomLevelChange: (zoomLevel: ZoomLevel) => void;
  setViewOptions: (viewOptions: ObjectiveViewOptions) => void;
  viewOptions: ObjectiveViewOptions;
  zoomLevel: ZoomLevel;
}) => {
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

  let content: ReactNode;
  if (isPending) {
    content = <BoardSkeleton className="h-full" layout={layout} />;
  } else if (objectives.length === 0) {
    content = emptyState;
  } else if (layout === "gantt") {
    content = (
      <RoadmapGanttBoard
        className="h-full"
        objectives={objectives}
        onObjectiveSelect={selectObjective}
        onZoomLevelChange={onZoomLevelChange}
        selectedObjectiveId={
          selectedObjective?.id ?? selectedKeyResult?.objective.id
        }
        zoomLevel={zoomLevel}
      />
    );
  } else {
    content = (
      <ObjectivesBoard
        layout={layout}
        objectives={objectives}
        onCreateObjective={onCreateObjective}
        onKeyResultSelect={selectKeyResult}
        onObjectiveSelect={selectObjective}
        selectedObjectiveId={
          selectedObjective?.id ?? selectedKeyResult?.objective.id
        }
        setViewOptions={setViewOptions}
        viewOptions={viewOptions}
      />
    );
  }

  return (
    <Box className="relative h-full min-w-0">
      {content}
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
  );
};
