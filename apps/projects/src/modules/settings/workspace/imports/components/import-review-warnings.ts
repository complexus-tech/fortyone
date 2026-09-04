import type { ImportDraft } from "../schema";
import type { ImportStructureMode } from "../import-run-model";
import {
  getCanonicalImportAssociation,
  getImportAssociationKey,
} from "../import-association-model";
import { isValidImportDateRange } from "../execution";
import type { ObjectiveTargetPlan } from "./import-wizard-model";
import type { ImportSelection } from "./use-import-selection";
import type { ImportTerms } from "./use-import-terms";

export const getImportRelationshipReview = ({
  draft,
  selection,
  fallbackTargetPlan,
  sourceTeamTargetPlans,
  structureMode,
}: {
  draft: ImportDraft | null;
  selection: Pick<
    ImportSelection,
    | "selectedTasks"
    | "selectedObjectives"
    | "importableKeyResults"
    | "importableSprints"
  >;
  fallbackTargetPlan: ObjectiveTargetPlan;
  sourceTeamTargetPlans: ReadonlyMap<string, ObjectiveTargetPlan>;
  structureMode: ImportStructureMode;
}) => {
  const {
    selectedTasks,
    selectedObjectives,
    importableKeyResults,
    importableSprints,
  } = selection;
  if (!draft) return { crossTeam: 0, unresolved: 0 };
  const taskCounts = new Map<string, number>();
  for (const task of selectedTasks) {
    taskCounts.set(task.sourceId, (taskCounts.get(task.sourceId) ?? 0) + 1);
  }
  const tasksBySourceId = new Map(
    selectedTasks
      .filter((task) => taskCounts.get(task.sourceId) === 1)
      .map((task) => [task.sourceId, task]),
  );
  const objectivesBySourceId = new Map(
    selectedObjectives.map((objective) => [objective.sourceId, objective]),
  );
  const keyResultsBySourceId = new Map(
    importableKeyResults.map((keyResult) => [keyResult.sourceId, keyResult]),
  );
  const sprintsBySourceId = new Map(
    importableSprints.map((sprint) => [sprint.sourceId, sprint]),
  );
  const getTargetTeamKey = (sourceTeamId: string | null) => {
    if (structureMode === "single" || !sourceTeamId) {
      return fallbackTargetPlan.teamKey;
    }
    const plan = sourceTeamTargetPlans.get(sourceTeamId);
    return plan?.teamConflict
      ? null
      : plan?.teamKey ?? fallbackTargetPlan.teamKey;
  };
  let crossTeam = 0;
  let unresolved = 0;
  const seenAssociationKeys = new Set<string>();

  for (const task of selectedTasks) {
    const taskTeamKey = getTargetTeamKey(task.teamSourceId);
    if (task.parentSourceId) {
      const parent = tasksBySourceId.get(task.parentSourceId);
      if (!parent) unresolved += 1;
      else if (
        taskTeamKey &&
        getTargetTeamKey(parent.teamSourceId) !== taskTeamKey
      ) {
        crossTeam += 1;
      }
    }
    if (
      task.objectiveSourceId &&
      !objectivesBySourceId.has(task.objectiveSourceId)
    ) {
      unresolved += 1;
    }
    if (
      task.keyResultSourceId &&
      !keyResultsBySourceId.has(task.keyResultSourceId)
    ) {
      unresolved += 1;
    }
    if (task.sprintSourceId) {
      const sprint = sprintsBySourceId.get(task.sprintSourceId);
      if (!sprint) unresolved += 1;
      else if (
        taskTeamKey &&
        getTargetTeamKey(sprint.teamSourceId) !== taskTeamKey
      ) {
        crossTeam += 1;
      }
    }
    for (const association of task.associations) {
      const targetSourceId = association.targetSourceId.trim();
      const canonicalAssociation = getCanonicalImportAssociation(
        task.sourceId,
        targetSourceId,
        association.type,
      );
      const associationKey = getImportAssociationKey(canonicalAssociation);
      if (seenAssociationKeys.has(associationKey)) continue;
      seenAssociationKeys.add(associationKey);
      const target = tasksBySourceId.get(targetSourceId);
      if (!target || target.sourceId === task.sourceId) unresolved += 1;
      else if (
        taskTeamKey &&
        getTargetTeamKey(target.teamSourceId) !== taskTeamKey
      ) {
        crossTeam += 1;
      }
    }
  }
  for (const sprint of importableSprints) {
    if (!sprint.objectiveSourceId) continue;
    const objective = objectivesBySourceId.get(sprint.objectiveSourceId);
    if (!objective) unresolved += 1;
    else if (
      getTargetTeamKey(sprint.teamSourceId) !==
      getTargetTeamKey(objective.teamSourceId)
    ) {
      crossTeam += 1;
    }
  }
  return { crossTeam, unresolved };
};
export const getImportReviewWarnings = (
  draft: ImportDraft | null,
  selection: Pick<
    ImportSelection,
    | "excludedObjectives"
    | "excludedStrategicPillars"
    | "importableKeyResults"
    | "importableSprints"
    | "selectedTasks"
    | "selectedObjectives"
    | "selectedTaskParentCycleCount"
  >,
  relationshipReview: ReturnType<typeof getImportRelationshipReview>,
  terms: ImportTerms,
) => {
  const {
    excludedObjectives,
    excludedStrategicPillars,
    importableKeyResults,
    importableSprints,
    selectedTasks,
    selectedObjectives,
    selectedTaskParentCycleCount,
  } = selection;
  const {
    keyResultTerm,
    keyResultTermPlural,
    sprintTerm,
    sprintTermPlural,
    storyTerm,
    storyTermPlural,
    objectiveTerm,
    objectiveTermPlural,
  } = terms;
  if (!draft) return [];
  const importableKeyResultSourceIds = new Set(
    importableKeyResults.map((keyResult) => keyResult.sourceId),
  );
  const incompleteKeyResults = draft.keyResults.filter(
    (keyResult) =>
      keyResult.objectiveSourceId &&
      !excludedObjectives.has(keyResult.objectiveSourceId) &&
      !importableKeyResultSourceIds.has(keyResult.sourceId),
  ).length;
  const incompleteSprints = draft.sprints.length - importableSprints.length;
  const invalidStoryDates = selectedTasks.filter(
    (task) => !isValidImportDateRange(task.startDate, task.endDate),
  ).length;
  const invalidObjectiveDates = selectedObjectives.filter(
    (objective) =>
      !isValidImportDateRange(objective.startDate, objective.endDate),
  ).length;
  const omittedPillarAlignments = selectedObjectives.filter(
    (objective) =>
      objective.pillarSourceId &&
      excludedStrategicPillars.has(objective.pillarSourceId),
  ).length;
  return [
    ...(incompleteKeyResults
      ? [
          `${incompleteKeyResults} ${incompleteKeyResults === 1 ? `${keyResultTerm} is` : `${keyResultTermPlural} are`} missing an explicit measure or valid date range and will be skipped.`,
        ]
      : []),
    ...(incompleteSprints
      ? [
          `${incompleteSprints} ${incompleteSprints === 1 ? `${sprintTerm} is` : `${sprintTermPlural} are`} missing a valid date range and will be skipped.`,
        ]
      : []),
    ...(invalidStoryDates
      ? [
          `${invalidStoryDates} ${invalidStoryDates === 1 ? `${storyTerm} has` : `${storyTermPlural} have`} an invalid date range; the work will import without those dates.`,
        ]
      : []),
    ...(invalidObjectiveDates
      ? [
          `${invalidObjectiveDates} ${invalidObjectiveDates === 1 ? `${objectiveTerm} has` : `${objectiveTermPlural} have`} an invalid date range; the ${objectiveTerm} will import without those dates.`,
        ]
      : []),
    ...(omittedPillarAlignments
      ? [
          `${omittedPillarAlignments} selected ${omittedPillarAlignments === 1 ? `${objectiveTerm} references` : `${objectiveTermPlural} reference`} a pillar that is not selected and will import without that alignment.`,
        ]
      : []),
    ...(selectedTaskParentCycleCount
      ? [
          `${selectedTaskParentCycleCount} selected ${selectedTaskParentCycleCount === 1 ? `${storyTerm} forms` : `${storyTermPlural} form`} a circular parent chain. Deselect at least one ${storyTerm} in each cycle before importing.`,
        ]
      : []),
    ...(relationshipReview.crossTeam
      ? [
          `${relationshipReview.crossTeam} ${relationshipReview.crossTeam === 1 ? "relationship spans" : "relationships span"} destination teams. Parent, ${sprintTerm}, and ${storyTerm}-association links must stay within one team, so these links will be skipped and reported.`,
        ]
      : []),
    ...(relationshipReview.unresolved
      ? [
          `${relationshipReview.unresolved} ${relationshipReview.unresolved === 1 ? "relationship points" : "relationships point"} to an object that is missing or not selected and will not be linked.`,
        ]
      : []),
    ...draft.warnings,
  ];
};
