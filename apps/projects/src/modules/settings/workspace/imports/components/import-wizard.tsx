"use client";

import { useEffect, useMemo, useState } from "react";
import { Box, Dialog } from "ui";
import { toast } from "sonner";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useJoinedTeams, useTeams } from "@/modules/teams/public/queries";
import type { ImportStructureMode } from "../import-run-model";
import { getImportSourceTeamDestination } from "../import-team-model";
import type { DestinationChoice } from "./import-wizard-model";
import {
  formatTeamCode,
  getSuggestedTeamName,
  initialNewTeam,
  WIZARD_STEP,
} from "./import-wizard-model";
import { collectImportIdentities } from "./import-identity-review";
import { canContinueImportStep } from "./import-step-validation";
import { useImportAnalysis } from "./use-import-analysis";
import { useImportExecution } from "./use-import-execution";
import { useImportPeopleReview } from "./use-import-people-review";
import { useImportPreflight } from "./use-import-preflight";
import { useImportReview } from "./use-import-review";
import { useImportSelection } from "./use-import-selection";
import { useImportTargetPlans } from "./use-import-target-plans";
import { useImportTerms } from "./use-import-terms";
import { ImportUploadStep } from "./import-upload-step";
import { ImportTeamsStep } from "./import-teams-step";
import { ImportMembersStep } from "./import-members-step";
import { ImportReviewStep } from "./import-review-step";
import { ImportRunStep } from "./import-run-step";
import { ImportWizardFooter } from "./import-wizard-footer";
import { WizardProgress } from "./import-wizard-progress";

type ImportWizardProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

export const ImportWizard = ({ onOpenChange, open }: ImportWizardProps) => {
  const { data: session } = useSession();
  const terms = useImportTerms();
  const { workspaceSlug } = useWorkspacePath();
  const { data: teams = [] } = useJoinedTeams();
  const {
    data: workspaceTeams = [],
    isError: workspaceTeamsError,
    isPending: workspaceTeamsPending,
  } = useTeams();
  const knownWorkspaceTeams = workspaceTeamsPending ? teams : workspaceTeams;
  const joinedTeamIds = useMemo(
    () => new Set(teams.map((team) => team.id)),
    [teams],
  );
  const [step, setStep] = useState<number>(WIZARD_STEP.upload);
  const [destination, setDestination] =
    useState<DestinationChoice>(initialNewTeam);
  const [structureMode, setStructureMode] =
    useState<ImportStructureMode>("single");
  const run = useImportExecution(session, workspaceSlug);
  const analysis = useImportAnalysis({
    workspaceSlug,
    onNewFile: () => {
      resetReview();
    },
    onUploaded: (uploadedDraft, uploadedFileName) => {
      selection.initializeArchivedRows(uploadedDraft);
      if (!uploadedDraft) return;
      setStructureMode(uploadedDraft.teams.length > 0 ? "preserve" : "single");
      const teamName = getSuggestedTeamName(
        uploadedDraft.sourceType,
        uploadedFileName,
      );
      const sourceTeam = uploadedDraft.teams.at(0);
      setDestination((current) => {
        if (current.kind !== "new") return current;
        if (teams.length === 0 && sourceTeam) {
          return { kind: "new", ...getImportSourceTeamDestination(sourceTeam) };
        }
        return { ...current, name: teamName, code: formatTeamCode(teamName) };
      });
    },
    onCompleted: (completedAnalysis) => {
      selection.setReviewPage(0);
      if (completedAnalysis.teams.length === 0) return;
      setStructureMode("preserve");
      const sourceTeam = completedAnalysis.teams[0];
      if (teams.length === 0) {
        const sourceDestination = getImportSourceTeamDestination(sourceTeam);
        setDestination((current) =>
          current.kind === "new"
            ? { kind: "new", ...sourceDestination }
            : current,
        );
      }
    },
  });
  const { draft, fileHash, fileName, analysisPending, uploadPending } =
    analysis;
  const targets = useImportTargetPlans({
    draft,
    destination,
    fileHash,
    structureMode,
    knownWorkspaceTeams,
    createdFallbackTeamForReview: run.createdFallbackTeamForReview,
    sourceTeamMappingsForReview: run.sourceTeamMappingsForReview,
  });
  const {
    fallbackTargetPlan,
    sourceTeamTargetPlans,
    objectiveTargetPlans,
    objectiveTargetTeamIds,
  } = targets;
  const selection = useImportSelection({
    draft,
    destination,
    structureMode,
    knownWorkspaceTeams,
    fallbackTargetPlan,
    sourceTeamTargetPlans,
  });
  const importIdentities = useMemo(
    () => collectImportIdentities(draft),
    [draft],
  );
  const hasImportIdentities = importIdentities.length > 0;
  const strategyPreflightRequired = Boolean(
    draft?.strategicPillars.length ||
      draft?.objectives.some((objective) => objective.pillarSourceId),
  );
  const preflight = useImportPreflight({
    session,
    workspaceSlug,
    fileHash,
    objectiveTargetTeamIds,
    hasImportIdentities,
    strategyPreflightRequired,
    objectiveTermPlural: terms.objectiveTermPlural,
  });
  const people = useImportPeopleReview(
    importIdentities,
    preflight.workspaceMembersForReview,
  );
  const {
    excludedObjectives,
    excludedStrategicPillars,
    selectedObjectives,
    selectedStrategicPillars,
    selectedTasks,
  } = selection;
  const {
    objectiveDestinationMatches,
    strategicPillarDestinationMatches,
    reviewWarnings,
  } = useImportReview({
    draft,
    selection,
    objectiveTargetPlans,
    objectivesByTeamId: preflight.objectivesByTeamId,
    sourceObjectiveMappingsForReview: run.sourceObjectiveMappingsForReview,
    strategyMapForReview: preflight.strategyMapForReview,
    fallbackTargetPlan,
    sourceTeamTargetPlans,
    structureMode,
    terms,
  });
  const canContinue = canContinueImportStep({
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
    hasSourceTeamConflict: targets.hasSourceTeamConflict,
    hasImportIdentities,
    strategicPillarDestinationMatches,
    objectiveDestinationMatches,
  });

  const resetReview = () => {
    run.reset();
    selection.reset();
    people.reset();
    preflight.reset();
  };
  const reset = () => {
    analysis.reset();
    resetReview();
    setStep(WIZARD_STEP.upload);
    setDestination(
      teams[0] ? { kind: "existing", teamId: teams[0].id } : initialNewTeam,
    );
    setStructureMode("single");
  };
  useEffect(() => {
    if (!open || destination.kind !== "new" || destination.name) return;
    if (teams[0]) setDestination({ kind: "existing", teamId: teams[0].id });
  }, [destination, open, teams]);

  const startImport = () => {
    run.startImport({
      draft,
      destination,
      fileHash,
      selectedEntityCount: selection.selectedEntityCount,
      lockMemberMappings: people.lockMappings,
      onStart: () => {
        setStep(WIZARD_STEP.import);
        analysis.stop();
      },
      hasSelectedTeamScopedImport: selection.hasSelectedTeamScopedImport,
      knownWorkspaceTeams,
      joinedTeamIds,
      selectedObjectives,
      selectedStrategicPillars,
      excludedRows: selection.excludedRows,
      structureMode,
    });
  };
  const handleOpenChange = (nextOpen: boolean) => {
    if (!nextOpen && (analysisPending || run.importPending || uploadPending)) {
      toast.info("Keep this window open while the import is being prepared.");
      return;
    }
    onOpenChange(nextOpen);
    if (!nextOpen) reset();
  };

  return (
    <Dialog onOpenChange={handleOpenChange} open={open}>
      <Dialog.Content
        className="mt-0 flex max-h-[calc(100dvh-2rem)] max-w-3xl flex-col md:mt-0"
        onEscapeKeyDown={(event) => {
          event.preventDefault();
        }}
        onInteractOutside={(event) => {
          event.preventDefault();
        }}
        overlayClassName="items-center py-4"
        size="lg"
      >
        <Dialog.Header className="shrink-0 px-6 pt-5 pb-2">
          <Dialog.Title className="text-xl">Import work</Dialog.Title>
          <Dialog.Description className="mt-2 px-0 text-base leading-6">
            Upload an export from your work tool. FortyOne maps its objects and
            relationships, then gives you a full review before anything is
            created.
          </Dialog.Description>
          <WizardProgress step={step} />
        </Dialog.Header>
        <Dialog.Body className="min-h-0 flex-1 overflow-y-auto overscroll-contain px-6 pt-4 pb-6">
          {step === WIZARD_STEP.upload ? (
            <ImportUploadStep
              analysisError={analysis.analysisError}
              analysisNotice={analysis.analysisNotice}
              analysisPending={analysisPending}
              fileName={fileName}
              handleFile={analysis.handleFile}
              hasAttemptedImport={run.hasAttemptedImport}
              setAnalysisError={analysis.setAnalysisError}
              uploadPending={uploadPending}
            />
          ) : null}
          {step === WIZARD_STEP.teams ? (
            <ImportTeamsStep
              conflictedSourceTeamIds={targets.conflictedSourceTeamIds}
              destination={destination}
              draft={draft}
              hasAttemptedImport={run.hasAttemptedImport}
              hasPrivacyWideningRisk={selection.hasPrivacyWideningRisk}
              hasSelectedTeamScopedImport={
                selection.hasSelectedTeamScopedImport
              }
              hasSourceTeamConflict={targets.hasSourceTeamConflict}
              joinedTeamIds={joinedTeamIds}
              knownWorkspaceTeams={knownWorkspaceTeams}
              objectivePreflightError={preflight.objectivePreflightError}
              objectivePreflightPending={preflight.objectivePreflightPending}
              objectiveTargetTeamIds={objectiveTargetTeamIds}
              privateSourceTeamCount={selection.privateSourceTeamCount}
              retryObjectives={preflight.retryObjectives}
              retryStrategy={preflight.retryStrategy}
              setDestination={setDestination}
              setStructureMode={setStructureMode}
              strategyPreflightError={preflight.strategyPreflightError}
              strategyPreflightPending={preflight.strategyPreflightPending}
              structureMode={structureMode}
              suggestedTeamName={getSuggestedTeamName(
                draft?.sourceType,
                fileName,
              )}
              teams={teams}
              workspaceTeamsError={workspaceTeamsError}
              workspaceTeamsPending={workspaceTeamsPending}
            />
          ) : null}
          {step === WIZARD_STEP.members ? (
            <ImportMembersStep
              hasAttemptedImport={run.hasAttemptedImport}
              hasImportIdentities={hasImportIdentities}
              peopleMappingPreview={people.peopleMappingPreview}
              peoplePreflightError={preflight.peoplePreflightError}
              peoplePreflightPending={preflight.peoplePreflightPending}
              retryPeople={preflight.retryPeople}
              reviewedMemberIdsByIdentityKey={
                people.reviewedMemberIdsByIdentityKey
              }
              selectMember={people.selectMember}
              workspaceMembersForReview={preflight.workspaceMembersForReview}
            />
          ) : null}
          {step === WIZARD_STEP.review && draft ? (
            <ImportReviewStep
              draft={draft}
              excludedObjectives={excludedObjectives}
              excludedStrategicPillars={excludedStrategicPillars}
              hasAttemptedImport={run.hasAttemptedImport}
              objectiveDestinationMatches={objectiveDestinationMatches}
              reviewWarnings={reviewWarnings}
              strategicPillarDestinationMatches={
                strategicPillarDestinationMatches
              }
              taskReview={{
                archivedTrelloTaskIndexes: selection.archivedTrelloTaskIndexes,
                includeArchivedTrelloCards:
                  selection.includeArchivedTrelloCards,
                toggleArchivedTrelloCards: selection.toggleArchivedTrelloCards,
                selectedTasks,
                visibleReviewTasks: selection.visibleReviewTasks,
                excludedRows: selection.excludedRows,
                toggleTask: selection.toggleTask,
                updateTaskTitle: analysis.updateTaskTitle,
                reviewTasks: selection.reviewTasks,
                reviewPageStart: selection.reviewPageStart,
                reviewPage: selection.reviewPage,
                reviewPageCount: selection.reviewPageCount,
                setReviewPage: selection.setReviewPage,
              }}
              toggleObjective={selection.toggleObjective}
              toggleStrategicPillar={selection.toggleStrategicPillar}
              updateMapping={(field, value) => {
                if (
                  analysis.updateMapping(field, value, run.hasAttemptedImport)
                )
                  selection.resetTaskSelection();
              }}
              updateObjectiveName={analysis.updateObjectiveName}
              updateStrategicPillarName={analysis.updateStrategicPillarName}
            />
          ) : null}
          {step === WIZARD_STEP.import ? (
            <Box className="py-6">
              <ImportRunStep
                importPending={run.importPending}
                importProgress={run.importProgress}
                outcome={run.outcome}
                runError={run.runError}
              />
            </Box>
          ) : null}
        </Dialog.Body>
        <ImportWizardFooter
          canContinue={canContinue}
          handleOpenChange={handleOpenChange}
          importPending={run.importPending}
          outcome={run.outcome}
          retryPreflight={preflight.retryPreflight}
          selectedEntityCount={selection.selectedEntityCount}
          setOutcome={run.setOutcome}
          setStep={setStep}
          startImport={startImport}
          step={step}
        />
      </Dialog.Content>
    </Dialog>
  );
};
