import { useMemo, useState } from "react";
import type { Team } from "@/modules/teams/public/types";
import type { ImportDraft } from "../schema";
import type { ImportStructureMode } from "../import-run-model";
import {
  getImportParentCycleSourceIds,
  isValidImportDateRange,
} from "../execution";
import {
  getTaskIndexesBySourceId,
  getTrelloArchivedTaskSourceIds,
} from "./import-draft-model";
import type {
  DestinationChoice,
  ObjectiveTargetPlan,
} from "./import-wizard-model";

export const REVIEW_PAGE_SIZE = 50;
export const useImportSelection = ({
  draft,
  destination,
  structureMode,
  knownWorkspaceTeams,
  fallbackTargetPlan,
  sourceTeamTargetPlans,
}: {
  draft: ImportDraft | null;
  destination: DestinationChoice;
  structureMode: ImportStructureMode;
  knownWorkspaceTeams: Team[];
  fallbackTargetPlan: ObjectiveTargetPlan;
  sourceTeamTargetPlans: ReadonlyMap<string, ObjectiveTargetPlan>;
}) => {
  const [excludedRows, setExcludedRows] = useState<Set<number>>(new Set());
  const [includeArchivedTrelloCards, setIncludeArchivedTrelloCards] =
    useState(false);
  const [excludedObjectives, setExcludedObjectives] = useState<Set<string>>(
    new Set(),
  );
  const [excludedStrategicPillars, setExcludedStrategicPillars] = useState<
    Set<string>
  >(new Set());
  const [reviewPage, setReviewPage] = useState(0);
  const selectedTasks = useMemo(
    () => draft?.tasks.filter((_, index) => !excludedRows.has(index)) ?? [],
    [draft?.tasks, excludedRows],
  );
  const archivedTrelloTaskSourceIds = useMemo(
    () => getTrelloArchivedTaskSourceIds(draft),
    [draft],
  );
  const archivedTrelloTaskIndexes = useMemo(
    () => getTaskIndexesBySourceId(draft, archivedTrelloTaskSourceIds),
    [archivedTrelloTaskSourceIds, draft],
  );
  const selectedStrategicPillars = useMemo(
    () =>
      draft?.strategicPillars.filter(
        (pillar) => !excludedStrategicPillars.has(pillar.sourceId),
      ) ?? [],
    [draft?.strategicPillars, excludedStrategicPillars],
  );
  const selectedObjectives = useMemo(
    () =>
      draft?.objectives.filter(
        (objective) => !excludedObjectives.has(objective.sourceId),
      ) ?? [],
    [draft?.objectives, excludedObjectives],
  );
  const importableKeyResults = useMemo(
    () =>
      draft?.keyResults.filter(
        (keyResult) =>
          keyResult.objectiveSourceId &&
          !excludedObjectives.has(keyResult.objectiveSourceId) &&
          keyResult.measurementType !== null &&
          keyResult.startValue !== null &&
          keyResult.currentValue !== null &&
          keyResult.targetValue !== null &&
          keyResult.startDate !== null &&
          keyResult.endDate !== null &&
          isValidImportDateRange(keyResult.startDate, keyResult.endDate),
      ) ?? [],
    [draft?.keyResults, excludedObjectives],
  );
  const importableSprints = useMemo(
    () =>
      draft?.sprints.filter(
        (sprint) =>
          sprint.startDate !== null &&
          sprint.endDate !== null &&
          isValidImportDateRange(sprint.startDate, sprint.endDate),
      ) ?? [],
    [draft?.sprints],
  );
  const hasSelectedTeamScopedImport = Boolean(
    selectedTasks.length > 0 ||
      selectedObjectives.length > 0 ||
      importableSprints.length > 0 ||
      Boolean(draft?.labels.some((label) => label.teamSourceId !== null)) ||
      (structureMode === "preserve" && (draft?.teams.length ?? 0) > 0),
  );
  const selectedSourceTeamIds = useMemo(() => {
    const sourceTeamIds = new Set<string>();
    const addSourceTeamId = (sourceTeamId: string | null) => {
      if (sourceTeamId) sourceTeamIds.add(sourceTeamId);
    };
    for (const task of selectedTasks) addSourceTeamId(task.teamSourceId);
    for (const objective of selectedObjectives) {
      addSourceTeamId(objective.teamSourceId);
    }
    for (const sprint of importableSprints) {
      addSourceTeamId(sprint.teamSourceId);
    }
    for (const label of draft?.labels ?? []) {
      addSourceTeamId(label.teamSourceId);
    }
    return sourceTeamIds;
  }, [draft?.labels, importableSprints, selectedObjectives, selectedTasks]);
  const privateSourceTeamCount =
    draft?.teams.filter(
      (team) => team.isPrivate && selectedSourceTeamIds.has(team.sourceId),
    ).length ?? 0;
  const destinationIsPrivate =
    destination.kind === "new"
      ? destination.isPrivate
      : knownWorkspaceTeams.find((team) => team.id === destination.teamId)
          ?.isPrivate;
  const hasPrivacyWideningRisk = Boolean(
    hasSelectedTeamScopedImport &&
      structureMode === "single" &&
      privateSourceTeamCount > 0 &&
      destinationIsPrivate === false,
  );
  const selectedTaskParentCycleCount = useMemo(() => {
    const getTargetTeamKey = (sourceTeamId: string | null) => {
      if (structureMode === "single" || !sourceTeamId) {
        return fallbackTargetPlan.teamKey;
      }
      const plan = sourceTeamTargetPlans.get(sourceTeamId);
      return plan?.teamConflict
        ? null
        : plan?.teamKey ?? fallbackTargetPlan.teamKey;
    };
    return getImportParentCycleSourceIds(
      selectedTasks,
      (task, parent) =>
        getTargetTeamKey(task.teamSourceId) ===
        getTargetTeamKey(parent.teamSourceId),
    ).size;
  }, [
    fallbackTargetPlan.teamKey,
    selectedTasks,
    sourceTeamTargetPlans,
    structureMode,
  ]);
  const selectedEntityCount =
    selectedTasks.length +
    selectedStrategicPillars.length +
    selectedObjectives.length +
    importableKeyResults.length +
    importableSprints.length +
    (draft?.labels.length ?? 0) +
    (structureMode === "preserve" ? draft?.teams.length ?? 0 : 0);
  const toggleTask = (index: number, checked: boolean) => {
    setExcludedRows((current) => {
      const next = new Set(current);
      if (checked) next.delete(index);
      else next.add(index);
      return next;
    });
  };

  const toggleArchivedTrelloCards = (included: boolean) => {
    setIncludeArchivedTrelloCards(included);
    setExcludedRows((current) => {
      const next = new Set(current);
      for (const index of archivedTrelloTaskIndexes) {
        if (included) next.delete(index);
        else next.add(index);
      }
      return next;
    });
    setReviewPage(0);
  };

  const toggleObjective = (sourceId: string, checked: boolean) => {
    setExcludedObjectives((current) => {
      const next = new Set(current);
      if (checked) next.delete(sourceId);
      else next.add(sourceId);
      return next;
    });
  };

  const toggleStrategicPillar = (sourceId: string, checked: boolean) => {
    setExcludedStrategicPillars((current) => {
      const next = new Set(current);
      if (checked) next.delete(sourceId);
      else next.add(sourceId);
      return next;
    });
  };

  const reviewTasks =
    draft?.tasks.flatMap((task, taskIndex) =>
      !includeArchivedTrelloCards && archivedTrelloTaskIndexes.has(taskIndex)
        ? []
        : [{ task, taskIndex }],
    ) ?? [];
  const reviewPageCount = Math.max(
    1,
    Math.ceil(reviewTasks.length / REVIEW_PAGE_SIZE),
  );
  const reviewPageStart = reviewPage * REVIEW_PAGE_SIZE;
  const visibleReviewTasks = reviewTasks.slice(
    reviewPageStart,
    reviewPageStart + REVIEW_PAGE_SIZE,
  );

  const reset = () => {
    setExcludedRows(new Set());
    setIncludeArchivedTrelloCards(false);
    setExcludedObjectives(new Set());
    setExcludedStrategicPillars(new Set());
    setReviewPage(0);
  };
  const initializeArchivedRows = (uploadedDraft: ImportDraft | null) => {
    const archivedSourceIds = getTrelloArchivedTaskSourceIds(uploadedDraft);
    setExcludedRows(getTaskIndexesBySourceId(uploadedDraft, archivedSourceIds));
  };
  const resetTaskSelection = () => {
    setExcludedRows(new Set());
    setReviewPage(0);
  };
  return {
    excludedRows,
    includeArchivedTrelloCards,
    excludedObjectives,
    excludedStrategicPillars,
    reviewPage,
    setReviewPage,
    selectedTasks,
    archivedTrelloTaskIndexes,
    selectedStrategicPillars,
    selectedObjectives,
    importableKeyResults,
    importableSprints,
    hasSelectedTeamScopedImport,
    privateSourceTeamCount,
    hasPrivacyWideningRisk,
    selectedTaskParentCycleCount,
    selectedEntityCount,
    toggleTask,
    toggleArchivedTrelloCards,
    toggleObjective,
    toggleStrategicPillar,
    reviewTasks,
    reviewPageCount,
    reviewPageStart,
    visibleReviewTasks,
    reset,
    initializeArchivedRows,
    resetTaskSelection,
  };
};
export type ImportSelection = ReturnType<typeof useImportSelection>;
