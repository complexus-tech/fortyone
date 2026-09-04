import type { RunImportInput } from "./import-run-model";
import { isValidImportDateRange } from "./execution";

export const selectImportWork = ({
  draft,
  selectedTaskIndexes,
  selectedObjectiveSourceIds,
  selectedStrategicPillarSourceIds,
  fallbackTeamId,
  fallbackTeamIsPrivate,
  structureMode,
}: Pick<
  RunImportInput,
  | "draft"
  | "selectedTaskIndexes"
  | "selectedObjectiveSourceIds"
  | "selectedStrategicPillarSourceIds"
  | "fallbackTeamId"
  | "fallbackTeamIsPrivate"
  | "structureMode"
>) => {
  const selectedTasks = draft.tasks.filter((_, index) =>
    selectedTaskIndexes.has(index),
  );
  const selectedObjectives = draft.objectives.filter((objective) =>
    selectedObjectiveSourceIds.has(objective.sourceId),
  );
  const selectedStrategicPillars = draft.strategicPillars.filter((pillar) =>
    selectedStrategicPillarSourceIds.has(pillar.sourceId),
  );
  const importableKeyResults = draft.keyResults.filter(
    (keyResult) =>
      keyResult.objectiveSourceId &&
      selectedObjectiveSourceIds.has(keyResult.objectiveSourceId) &&
      keyResult.measurementType !== null &&
      keyResult.startValue !== null &&
      keyResult.currentValue !== null &&
      keyResult.targetValue !== null &&
      keyResult.startDate !== null &&
      keyResult.endDate !== null &&
      isValidImportDateRange(keyResult.startDate, keyResult.endDate),
  );
  const importableSprints = draft.sprints.filter(
    (sprint) =>
      sprint.startDate !== null &&
      sprint.endDate !== null &&
      isValidImportDateRange(sprint.startDate, sprint.endDate),
  );
  const hasTeamScopedImport = Boolean(
    selectedTasks.length > 0 ||
      selectedObjectives.length > 0 ||
      importableSprints.length > 0 ||
      draft.labels.some((label) => label.teamSourceId !== null) ||
      (structureMode === "preserve" && draft.teams.length > 0),
  );
  if (!fallbackTeamId && hasTeamScopedImport) {
    throw new Error("A destination team is required for team-scoped work");
  }
  if (structureMode === "single" && !fallbackTeamIsPrivate) {
    const selectedSourceTeamIds = new Set<string>();
    const addSourceTeamId = (sourceTeamId: string | null) => {
      if (sourceTeamId) selectedSourceTeamIds.add(sourceTeamId);
    };
    for (const task of selectedTasks) addSourceTeamId(task.teamSourceId);
    for (const objective of selectedObjectives) {
      addSourceTeamId(objective.teamSourceId);
    }
    for (const sprint of importableSprints) {
      addSourceTeamId(sprint.teamSourceId);
    }
    for (const label of draft.labels) addSourceTeamId(label.teamSourceId);
    const widensPrivateSourceWork = draft.teams.some(
      (team) => team.isPrivate && selectedSourceTeamIds.has(team.sourceId),
    );
    if (widensPrivateSourceWork) {
      throw new Error(
        "Private source work cannot be combined into a public destination team",
      );
    }
  }

  return {
    selectedTasks,
    selectedObjectives,
    selectedStrategicPillars,
    importableKeyResults,
    importableSprints,
  };
};
export type ImportSelection = ReturnType<typeof selectImportWork>;
