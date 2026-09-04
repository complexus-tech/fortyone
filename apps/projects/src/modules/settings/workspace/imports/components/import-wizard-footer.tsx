"use client";
import type { Dispatch, ReactNode, SetStateAction } from "react";
import { ArrowLeft2Icon, ArrowRight2Icon } from "icons";
import { Button, Dialog } from "ui";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import type { ImportRunResult } from "../import-run-model";
import { WIZARD_STEP } from "./import-wizard-model";

export type ImportWizardFooterProps = {
  step: number;
  setStep: Dispatch<SetStateAction<number>>;
  importPending: boolean;
  outcome: ImportRunResult | null;
  setOutcome: (outcome: ImportRunResult | null) => void;
  canContinue: boolean;
  selectedEntityCount: number;
  startImport: () => void;
  retryPreflight: () => void;
  handleOpenChange: (open: boolean) => void;
};
export const ImportWizardFooter = ({
  step,
  setStep,
  importPending,
  outcome,
  setOutcome,
  canContinue,
  selectedEntityCount,
  startImport,
  retryPreflight,
  handleOpenChange,
}: ImportWizardFooterProps) => {
  const { withWorkspace } = useWorkspacePath();
  let footerBack: ReactNode = <span />;
  if (step > WIZARD_STEP.upload && step < WIZARD_STEP.import) {
    footerBack = (
      <Button
        color="tertiary"
        leftIcon={<ArrowLeft2Icon />}
        onClick={() => {
          setStep((current) => Math.max(WIZARD_STEP.upload, current - 1));
        }}
        variant="outline"
      >
        Back
      </Button>
    );
  } else if (step === WIZARD_STEP.import && !importPending && !outcome) {
    footerBack = (
      <Button
        color="tertiary"
        leftIcon={<ArrowLeft2Icon />}
        onClick={() => {
          retryPreflight();
          setStep(WIZARD_STEP.review);
        }}
        variant="outline"
      >
        Review again
      </Button>
    );
  } else if (
    step === WIZARD_STEP.import &&
    (outcome?.failed ||
      outcome?.destinationConflicts ||
      outcome?.unresolvedAssociations ||
      outcome?.unresolvedLinks ||
      outcome?.unresolvedPeople)
  ) {
    footerBack = (
      <Button
        color="tertiary"
        leftIcon={<ArrowLeft2Icon />}
        onClick={() => {
          setOutcome(null);
          retryPreflight();
          setStep(WIZARD_STEP.review);
        }}
        variant="outline"
      >
        Review import
      </Button>
    );
  }

  const outcomeStrategyChanges = outcome
    ? outcome.createdStrategicPillars + outcome.alignedObjectives
    : 0;
  const outcomeNonStrategyChanges = outcome
    ? outcome.created +
      outcome.replayed +
      outcome.createdTeams +
      outcome.createdObjectives +
      outcome.createdKeyResults +
      outcome.createdSprints +
      outcome.createdLabels +
      outcome.createdLinks +
      outcome.addedMemberships +
      outcome.appliedCollaborators +
      outcome.createdAssociations
    : 0;
  const outcomeHasIssues = Boolean(
    outcome &&
      (outcome.failed ||
        outcome.destinationConflicts ||
        outcome.unresolvedAssociations ||
        outcome.unresolvedLinks ||
        outcome.unresolvedPeople),
  );
  const canViewOutcome = Boolean(
    outcome &&
      (outcomeStrategyChanges + outcomeNonStrategyChanges > 0 ||
        !outcomeHasIssues),
  );
  const viewOutcomeInStrategy = Boolean(
    outcome && outcomeStrategyChanges > 0 && outcomeNonStrategyChanges === 0,
  );

  let footerAction: ReactNode = null;
  if (step < WIZARD_STEP.review) {
    footerAction = (
      <Button
        color="invert"
        disabled={!canContinue}
        onClick={() => {
          setStep((current) => current + 1);
        }}
      >
        Continue
      </Button>
    );
  } else if (step === WIZARD_STEP.review) {
    footerAction = (
      <Button color="invert" disabled={!canContinue} onClick={startImport}>
        Import {selectedEntityCount}{" "}
        {selectedEntityCount === 1 ? "item" : "items"}
      </Button>
    );
  } else if (outcome && canViewOutcome) {
    footerAction = (
      <Button
        color="invert"
        href={withWorkspace(viewOutcomeInStrategy ? "/strategy" : "/summary")}
        onClick={() => {
          handleOpenChange(false);
        }}
        rightIcon={<ArrowRight2Icon />}
      >
        {viewOutcomeInStrategy ? "View strategy" : "View workspace summary"}
      </Button>
    );
  } else if (outcome) {
    footerAction = (
      <Button color="invert" onClick={startImport}>
        Retry safely
      </Button>
    );
  } else if (!importPending) {
    footerAction = (
      <Button color="invert" onClick={startImport}>
        Retry safely
      </Button>
    );
  }

  return (
    <Dialog.Footer
      className="bg-surface-muted/35 shrink-0 gap-3 px-6 py-4"
      justify="between"
      variant="bordered"
    >
      {footerBack}
      {footerAction}
    </Dialog.Footer>
  );
};
