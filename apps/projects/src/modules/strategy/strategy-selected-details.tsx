"use client";

import type { Objective } from "@/modules/objectives/types";
import { RoadmapObjectiveDetails } from "@/modules/roadmap/components/objective-details";
import { StrategyNodeDetails } from "./strategy-node-details";
import type { StrategyMap } from "./types";

export type SelectedStrategyNode =
  | { type: "goal" }
  | { objectiveId: string; type: "objective" }
  | { pillarId: string; type: "pillar" };

export const StrategySelectedDetails = ({
  canEdit,
  objectives,
  onClose,
  onSaveGoal,
  onSavePillar,
  selectedNode,
  strategy,
}: {
  canEdit: boolean;
  objectives: Objective[];
  onClose: () => void;
  onSaveGoal: (ultimateGoal: string, description: string | null) => void;
  onSavePillar: (
    pillarId: string,
    name: string,
    description: string | null,
  ) => void;
  selectedNode: SelectedStrategyNode | null;
  strategy: StrategyMap;
}) => {
  const selectedObjective =
    selectedNode?.type === "objective"
      ? objectives.find(({ id }) => id === selectedNode.objectiveId)
      : undefined;
  const selectedPillar =
    selectedNode?.type === "pillar"
      ? strategy.pillars.find(({ id }) => id === selectedNode.pillarId)
      : undefined;
  const objectiveIds = new Set(objectives.map(({ id }) => id));
  const selectedPillarObjectiveCount =
    selectedPillar?.objectiveIds.filter((id) => objectiveIds.has(id)).length ??
    0;

  if (selectedObjective) {
    return (
      <RoadmapObjectiveDetails
        objective={selectedObjective}
        onClose={onClose}
      />
    );
  }

  if (selectedNode?.type === "goal") {
    return (
      <StrategyNodeDetails
        canEdit={canEdit}
        description={strategy.description}
        entityKey="ultimate-goal"
        kind="goal"
        name={strategy.ultimateGoal}
        objectiveCount={objectives.length}
        onClose={onClose}
        onSave={onSaveGoal}
        pillarCount={strategy.pillars.length}
      />
    );
  }

  if (selectedPillar) {
    return (
      <StrategyNodeDetails
        canEdit={canEdit}
        description={selectedPillar.description}
        entityKey={selectedPillar.id}
        kind="pillar"
        name={selectedPillar.name}
        objectiveCount={selectedPillarObjectiveCount}
        onClose={onClose}
        onSave={(name, description) => {
          onSavePillar(selectedPillar.id, name, description);
        }}
      />
    );
  }

  return null;
};
