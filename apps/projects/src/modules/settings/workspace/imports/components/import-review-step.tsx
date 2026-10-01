"use client";
import { Box, Text } from "ui";
import type { ImportDraft, ImportMapping } from "../schema";
import type {
  ObjectiveImportDestinationPreview,
  StrategicPillarImportDestinationPreview,
} from "./import-review-types";
import {
  ImportPlanSummary,
  ObjectiveImportReview,
  StrategicPillarImportReview,
} from "./import-graph-review";
import { ImportFieldMapping } from "./import-field-mapping";
import { ImportResumePanel } from "./import-resume-panel";
import {
  ImportTaskReview,
  type ImportTaskReviewProps,
} from "./import-task-review";

export type ImportReviewStepProps = {
  draft: ImportDraft;
  hasAttemptedImport: boolean;
  reviewWarnings: string[];
  strategicPillarDestinationMatches: ReadonlyMap<
    string,
    StrategicPillarImportDestinationPreview
  >;
  objectiveDestinationMatches: ReadonlyMap<
    string,
    ObjectiveImportDestinationPreview
  >;
  excludedStrategicPillars: Set<string>;
  excludedObjectives: Set<string>;
  toggleStrategicPillar: (sourceId: string, checked: boolean) => void;
  toggleObjective: (sourceId: string, checked: boolean) => void;
  updateStrategicPillarName: (sourceId: string, name: string) => void;
  updateObjectiveName: (sourceId: string, name: string) => void;
  updateMapping: (field: keyof ImportMapping, value: string) => void;
  taskReview: Omit<ImportTaskReviewProps, "draft">;
};
export const ImportReviewStep = ({
  draft,
  hasAttemptedImport,
  reviewWarnings,
  strategicPillarDestinationMatches,
  objectiveDestinationMatches,
  excludedStrategicPillars,
  excludedObjectives,
  toggleStrategicPillar,
  toggleObjective,
  updateStrategicPillarName,
  updateObjectiveName,
  updateMapping,
  taskReview,
}: ImportReviewStepProps) => (
  <Box>
    <Text as="h2" className="text-xl font-medium">
      Review your import
    </Text>
    <Text className="mt-1 leading-6" color="muted">
      Confirm the proposed structure, field mapping, and work that FortyOne will
      create.
    </Text>

    <ImportPlanSummary draft={draft} />
    <ImportResumePanel draft={draft} />
    {draft.customFields?.length ? (
      <Box className="border-border mt-5 rounded-xl border p-4">
        <Text className="font-medium">Custom fields to preserve</Text>
        <Text className="mt-1 leading-6" color="muted">
          Reuse matching definitions; incompatible fields are reported.
        </Text>
        <Box as="ul" className="mt-3 max-h-60 space-y-2 overflow-auto">
          {draft.customFields.map((field) => (
            <li key={`${field.teamSourceId}:${field.sourceId}`}>
              {field.name} · {field.type}
              {field.currency ? ` (${field.currency})` : ""}
              {field.archivedAt ? " · archived after import" : ""}
            </li>
          ))}
        </Box>
      </Box>
    ) : null}

    {reviewWarnings.length ? (
      <Box className="bg-warning/8 mt-5 rounded-xl p-4">
        <Text className="text-foreground/90 font-medium dark:text-white/90">
          Mapping notes
        </Text>
        <Box
          as="ul"
          className="text-foreground/75 mt-2 max-h-24 overflow-y-auto pr-2 pl-5 dark:text-white/70"
        >
          {reviewWarnings.map((warning) => (
            <li className="list-disc leading-6" key={warning}>
              {warning}
            </li>
          ))}
        </Box>
      </Box>
    ) : null}

    <StrategicPillarImportReview
      destinationMatches={strategicPillarDestinationMatches}
      draft={draft}
      excludedSourceIds={excludedStrategicPillars}
      onCheckedChange={toggleStrategicPillar}
      onNameChange={updateStrategicPillarName}
    />

    <ObjectiveImportReview
      destinationMatches={objectiveDestinationMatches}
      draft={draft}
      excludedSourceIds={excludedObjectives}
      onCheckedChange={toggleObjective}
      onNameChange={updateObjectiveName}
    />

    <ImportFieldMapping
      draft={draft}
      hasAttemptedImport={hasAttemptedImport}
      updateMapping={updateMapping}
    />
    <ImportTaskReview draft={draft} {...taskReview} />
  </Box>
);
