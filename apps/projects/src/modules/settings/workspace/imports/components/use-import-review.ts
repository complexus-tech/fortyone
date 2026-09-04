import { useMemo } from "react";
import type { Objective } from "@/modules/objectives/public/types";
import type { StrategyMap } from "@/shared/strategy-map/types";
import type { ImportDraft } from "../schema";
import type { ImportStructureMode } from "../import-run-model";
import type { ObjectiveTargetPlan } from "./import-wizard-model";
import type { ImportSelection } from "./use-import-selection";
import type { ImportTerms } from "./use-import-terms";
import {
  getObjectiveDestinationMatches,
  getStrategicPillarDestinationMatches,
} from "./import-destination-matches";
import {
  getImportRelationshipReview,
  getImportReviewWarnings,
} from "./import-review-warnings";

type ImportReviewInput = {
  draft: ImportDraft | null;
  selection: ImportSelection;
  objectiveTargetPlans: ReadonlyMap<string, ObjectiveTargetPlan>;
  objectivesByTeamId: ReadonlyMap<string, Objective[]>;
  sourceObjectiveMappingsForReview: ReadonlyMap<
    string,
    { id: string; teamId: string }
  >;
  strategyMapForReview: StrategyMap;
  fallbackTargetPlan: ObjectiveTargetPlan;
  sourceTeamTargetPlans: ReadonlyMap<string, ObjectiveTargetPlan>;
  structureMode: ImportStructureMode;
  terms: ImportTerms;
};

export const useImportReview = ({
  draft,
  selection,
  objectiveTargetPlans,
  objectivesByTeamId,
  sourceObjectiveMappingsForReview,
  strategyMapForReview,
  fallbackTargetPlan,
  sourceTeamTargetPlans,
  structureMode,
  terms,
}: ImportReviewInput) => {
  const {
    excludedObjectives,
    excludedStrategicPillars,
    importableKeyResults,
    importableSprints,
    selectedObjectives,
    selectedTasks,
    selectedTaskParentCycleCount,
  } = selection;
  const { objectiveTerm } = terms;
  const objectiveDestinationMatches = useMemo(
    () =>
      getObjectiveDestinationMatches({
        draft,
        excludedObjectives,
        excludedStrategicPillars,
        objectiveTerm,
        objectiveTargetPlans,
        objectivesByTeamId,
        sourceObjectiveMappingsForReview,
        strategyMapForReview,
      }),
    [
      draft,
      excludedObjectives,
      excludedStrategicPillars,
      objectiveTerm,
      objectiveTargetPlans,
      objectivesByTeamId,
      sourceObjectiveMappingsForReview,
      strategyMapForReview,
    ],
  );
  const strategicPillarDestinationMatches = useMemo(
    () =>
      getStrategicPillarDestinationMatches(
        draft,
        excludedStrategicPillars,
        strategyMapForReview,
      ),
    [draft, excludedStrategicPillars, strategyMapForReview],
  );
  const relationshipReview = useMemo(
    () =>
      getImportRelationshipReview({
        draft,
        selection: {
          selectedTasks,
          selectedObjectives,
          importableKeyResults,
          importableSprints,
        },
        fallbackTargetPlan,
        sourceTeamTargetPlans,
        structureMode,
      }),
    [
      draft,
      selectedTasks,
      selectedObjectives,
      importableKeyResults,
      importableSprints,
      fallbackTargetPlan,
      sourceTeamTargetPlans,
      structureMode,
    ],
  );
  const reviewWarnings = useMemo(
    () =>
      getImportReviewWarnings(
        draft,
        {
          excludedObjectives,
          excludedStrategicPillars,
          importableKeyResults,
          importableSprints,
          selectedTasks,
          selectedObjectives,
          selectedTaskParentCycleCount,
        },
        relationshipReview,
        terms,
      ),
    [
      draft,
      excludedObjectives,
      excludedStrategicPillars,
      importableKeyResults,
      importableSprints,
      selectedTasks,
      selectedObjectives,
      selectedTaskParentCycleCount,
      relationshipReview,
      terms,
    ],
  );
  return {
    objectiveDestinationMatches,
    strategicPillarDestinationMatches,
    reviewWarnings,
  };
};
