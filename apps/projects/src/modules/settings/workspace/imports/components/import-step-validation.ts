import type { ImportDraft } from "../schema";
import { importDestinationSchema } from "../schema";
import type { ImportStructureMode } from "../import-run-model";
import type { DestinationChoice } from "./import-wizard-model";
import { WIZARD_STEP } from "./import-wizard-model";
import type { ImportSelection } from "./use-import-selection";
import type { useImportPreflight } from "./use-import-preflight";
import type {
  ObjectiveImportDestinationPreview,
  StrategicPillarImportDestinationPreview,
} from "./import-review-types";

export const canContinueImportStep = ({
  step,
  draft,
  destination,
  structureMode,
  analysisPending,
  uploadPending,
  selection,
  preflight,
  workspaceTeamsPending,
  workspaceTeamsError,
  hasSourceTeamConflict,
  hasImportIdentities,
  strategicPillarDestinationMatches,
  objectiveDestinationMatches,
}: {
  step: number;
  draft: ImportDraft | null;
  destination: DestinationChoice;
  structureMode: ImportStructureMode;
  analysisPending: boolean;
  uploadPending: boolean;
  selection: ImportSelection;
  preflight: ReturnType<typeof useImportPreflight>;
  workspaceTeamsPending: boolean;
  workspaceTeamsError: boolean;
  hasSourceTeamConflict: boolean;
  hasImportIdentities: boolean;
  strategicPillarDestinationMatches: ReadonlyMap<
    string,
    StrategicPillarImportDestinationPreview
  >;
  objectiveDestinationMatches: ReadonlyMap<
    string,
    ObjectiveImportDestinationPreview
  >;
}) => {
  const {
    hasSelectedTeamScopedImport,
    hasPrivacyWideningRisk,
    selectedEntityCount,
    selectedTasks,
    selectedTaskParentCycleCount,
    selectedStrategicPillars,
    selectedObjectives,
  } = selection;
  const {
    objectivePreflightPending,
    objectivePreflightError,
    strategyPreflightPending,
    strategyPreflightError,
    peoplePreflightPending,
    peoplePreflightError,
  } = preflight;
  const destinationValid =
    importDestinationSchema.safeParse(destination).success;
  let canContinue = false;
  if (step === WIZARD_STEP.upload) {
    canContinue =
      Boolean(
        draft &&
          (draft.tasks.length > 0 ||
            draft.strategicPillars.length > 0 ||
            draft.objectives.length > 0 ||
            draft.sprints.length > 0 ||
            draft.labels.length > 0 ||
            draft.teams.length > 0),
      ) &&
      !analysisPending &&
      !uploadPending;
  } else if (step === WIZARD_STEP.teams) {
    canContinue =
      (!hasSelectedTeamScopedImport || destinationValid) &&
      !hasPrivacyWideningRisk &&
      (!hasSelectedTeamScopedImport ||
        structureMode === "single" ||
        (!workspaceTeamsPending &&
          !workspaceTeamsError &&
          !hasSourceTeamConflict)) &&
      (!hasSelectedTeamScopedImport ||
        (!objectivePreflightPending && !objectivePreflightError)) &&
      !strategyPreflightPending &&
      !strategyPreflightError;
  } else if (step === WIZARD_STEP.members) {
    canContinue =
      !hasImportIdentities ||
      (!peoplePreflightPending && !peoplePreflightError);
  } else if (step === WIZARD_STEP.review) {
    canContinue =
      selectedEntityCount > 0 &&
      selectedTasks.every((task) => task.title.trim().length > 0) &&
      selectedTaskParentCycleCount === 0 &&
      !hasPrivacyWideningRisk &&
      (!hasSelectedTeamScopedImport ||
        (!objectivePreflightPending && !objectivePreflightError)) &&
      (!hasImportIdentities ||
        (!peoplePreflightPending && !peoplePreflightError)) &&
      !strategyPreflightPending &&
      !strategyPreflightError &&
      selectedStrategicPillars.every((pillar) => {
        if (!pillar.name.trim()) return false;
        const match = strategicPillarDestinationMatches.get(pillar.sourceId);
        return match?.kind === "none" || match?.kind === "unique";
      }) &&
      selectedObjectives.every((objective) => {
        if (!objective.name.trim()) return false;
        const match = objectiveDestinationMatches.get(objective.sourceId);
        return match?.kind === "none" || match?.kind === "unique";
      });
  }

  return canContinue;
};
