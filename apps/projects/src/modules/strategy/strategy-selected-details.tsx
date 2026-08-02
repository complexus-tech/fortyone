"use client";

import type { KeyResult, Objective } from "@/modules/objectives/types";
import { useKeyResults } from "@/modules/objectives/hooks/use-key-results";
import { KeyResultDetails } from "@/modules/key-results/components/key-result-details";
import { RoadmapObjectiveDetails } from "@/modules/roadmap/components/objective-details";
import { StrategyNodeDetails } from "./strategy-node-details";
import type { StrategyMap } from "./types";

export type SelectedStrategyNode =
  | { type: "goal" }
  | { keyResultId: string; objectiveId: string; type: "key-result" }
  | { objectiveId: string; type: "objective" }
  | { pillarId: string; type: "pillar" };

export const StrategySelectedDetails = ({
  canEdit,
  objectives,
  onClose,
  onSaveGoal,
  onSavePillar,
  onSelectKeyResult,
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
  onSelectKeyResult: (objective: Objective, keyResult: KeyResult) => void;
  selectedNode: SelectedStrategyNode | null;
  strategy: StrategyMap;
}) => {
  const selectedObjectiveId =
    selectedNode?.type === "objective" || selectedNode?.type === "key-result"
      ? selectedNode.objectiveId
      : undefined;
  const selectedObjective = objectives.find(
    ({ id }) => id === selectedObjectiveId,
  );
  const { data: selectedObjectiveKeyResults = [] } = useKeyResults(
    selectedObjectiveId ?? "",
    selectedNode?.type === "key-result",
  );
  const selectedKeyResult =
    selectedNode?.type === "key-result"
      ? selectedObjectiveKeyResults.find(
          ({ id }) => id === selectedNode.keyResultId,
        )
      : undefined;
  const selectedPillar =
    selectedNode?.type === "pillar"
      ? strategy.pillars.find(({ id }) => id === selectedNode.pillarId)
      : undefined;
  const objectiveIds = new Set(objectives.map(({ id }) => id));
  const selectedPillarObjectiveCount =
    selectedPillar?.objectiveIds.filter((id) => objectiveIds.has(id)).length ??
    0;

  if (selectedObjective && selectedKeyResult) {
    return (
      <KeyResultDetails
        initialKeyResult={selectedKeyResult}
        objective={selectedObjective}
        onClose={onClose}
      />
    );
  }

  if (selectedObjective && selectedNode?.type === "objective") {
    return (
      <RoadmapObjectiveDetails
        objective={selectedObjective}
        onClose={onClose}
        onKeyResultSelect={(keyResult) => {
          onSelectKeyResult(selectedObjective, keyResult);
        }}
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
