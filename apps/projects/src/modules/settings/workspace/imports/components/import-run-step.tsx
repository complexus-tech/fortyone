"use client";
import { CheckIcon, WarningIcon } from "icons";
import { cn } from "lib";
import { Box, Button, ProgressBar, Text } from "ui";
import { useTerminology } from "@/hooks/use-terminology-display";
import type { ImportRunResult } from "../import-run-model";

const formatEntityCount = (
  count: number,
  singular: string,
  plural = `${singular}s`,
) => `${count} ${count === 1 ? singular : plural}`;

export const ImportRunStep = ({
  importPending,
  importProgress,
  outcome,
  runError,
}: {
  importPending: boolean;
  importProgress: number;
  outcome: ImportRunResult | null;
  runError: string;
}) => {
  const { getTermDisplay } = useTerminology();
  const storyTerm = getTermDisplay("storyTerm");
  const storyTermPlural = getTermDisplay("storyTerm", { variant: "plural" });
  const storyTermPluralCapitalized = getTermDisplay("storyTerm", {
    capitalize: true,
    variant: "plural",
  });
  const objectiveTerm = getTermDisplay("objectiveTerm");
  const objectiveTermPlural = getTermDisplay("objectiveTerm", {
    variant: "plural",
  });
  const keyResultTerm = getTermDisplay("keyResultTerm");
  const keyResultTermPlural = getTermDisplay("keyResultTerm", {
    variant: "plural",
  });
  const sprintTerm = getTermDisplay("sprintTerm");
  const sprintTermPlural = getTermDisplay("sprintTerm", { variant: "plural" });

  if (importPending) {
    const preparing = importProgress === 0;
    return (
      <Box
        aria-busy="true"
        aria-live="polite"
        className="mx-auto max-w-xl text-center"
        role="status"
      >
        <Text as="h2" className="text-2xl font-medium">
          {preparing ? "Preparing your import" : "Importing your work"}
        </Text>
        <Text className="mt-2 leading-6" color="muted">
          {preparing
            ? "FortyOne is checking teams, workflows, members, and relationships before creating work."
            : "FortyOne is creating the reviewed structure in dependency order. Keep this window open until the result is ready."}
        </Text>
        <ProgressBar className="mt-6 h-2" progress={importProgress} />
        <Text className="mt-2 font-medium">
          {preparing ? "Preparing…" : `${importProgress}% complete`}
        </Text>
      </Box>
    );
  }

  if (outcome) {
    const createdStructure =
      outcome.createdTeams +
      outcome.createdStrategicPillars +
      outcome.createdObjectives +
      outcome.createdKeyResults +
      outcome.createdSprints +
      outcome.createdLabels +
      outcome.createdLinks +
      outcome.addedMemberships +
      outcome.appliedCollaborators +
      outcome.createdAssociations +
      outcome.alignedObjectives;
    const successful = outcome.created + outcome.replayed + createdStructure;
    const hasIssues =
      Boolean(outcome.paused) ||
      Boolean(outcome.unresolvedComments) ||
      Boolean(outcome.unresolvedCustomFieldValues) ||
      Boolean(outcome.customFieldIssues?.length) ||
      outcome.failed > 0 ||
      outcome.destinationConflicts > 0 ||
      outcome.unresolvedAssociations > 0 ||
      outcome.unresolvedLinks > 0 ||
      outcome.unresolvedPeople > 0;
    const allFailed = successful === 0 && hasIssues;
    const partial = successful > 0 && hasIssues;
    let outcomeTitle = "Your import is ready";
    if (allFailed) outcomeTitle = "Nothing was imported";
    else if (partial) outcomeTitle = "Import finished with issues";
    if (outcome.paused) outcomeTitle = "Import paused with a saved receipt";
    let outcomeLead = "Applied the reviewed import";
    if (allFailed) outcomeLead = "Created no work";
    else if (outcome.created) {
      outcomeLead = `Created ${formatEntityCount(outcome.created, storyTerm, storyTermPlural)}`;
    }

    return (
      <Box aria-live="polite" className="mx-auto max-w-xl text-center">
        <Box
          className={cn(
            "mx-auto flex h-16 w-16 items-center justify-center rounded-3xl",
            hasIssues
              ? "bg-warning/10 text-warning"
              : "bg-success/10 text-success",
          )}
        >
          {hasIssues ? (
            <WarningIcon className="h-8" />
          ) : (
            <CheckIcon className="h-8" strokeWidth={2.5} />
          )}
        </Box>
        <Text as="h2" className="mt-5 text-2xl font-medium">
          {outcomeTitle}
        </Text>
        <Text className="mt-2 leading-6" color="muted">
          {outcomeLead}
          {outcome.replayed
            ? `, recognized ${formatEntityCount(outcome.replayed, `previously imported ${storyTerm}`, `previously imported ${storyTermPlural}`)}`
            : ""}
          {outcome.createdTeams
            ? `, ${formatEntityCount(outcome.createdTeams, "team")}`
            : ""}
          {outcome.createdStrategicPillars
            ? `, ${formatEntityCount(outcome.createdStrategicPillars, "strategic pillar")}`
            : ""}
          {outcome.createdObjectives
            ? `, ${formatEntityCount(outcome.createdObjectives, objectiveTerm, objectiveTermPlural)}`
            : ""}
          {outcome.createdKeyResults
            ? `, ${formatEntityCount(outcome.createdKeyResults, keyResultTerm, keyResultTermPlural)}`
            : ""}
          {outcome.createdSprints
            ? `, ${formatEntityCount(outcome.createdSprints, sprintTerm, sprintTermPlural)}`
            : ""}
          {outcome.createdLabels
            ? `, ${formatEntityCount(outcome.createdLabels, "label")}`
            : ""}
          {outcome.createdLinks
            ? `, ${formatEntityCount(outcome.createdLinks, `${storyTerm} link`)}`
            : ""}
          {outcome.addedMemberships
            ? `, ${formatEntityCount(outcome.addedMemberships, "team membership")}`
            : ""}
          {outcome.appliedCollaborators
            ? `, ${formatEntityCount(outcome.appliedCollaborators, "collaborator assignment")}`
            : ""}
          {outcome.createdAssociations
            ? `, ${formatEntityCount(outcome.createdAssociations, `${storyTerm} relationship`)}`
            : ""}
          {outcome.alignedObjectives
            ? `, ${formatEntityCount(outcome.alignedObjectives, `${objectiveTerm} alignment`)}`
            : ""}
          {outcome.failed
            ? `, and found ${formatEntityCount(outcome.failed, storyTerm, storyTermPlural)} that ${outcome.failed === 1 ? "needs" : "need"} attention`
            : ""}
          .
        </Text>
        {outcome.paused ? (
          <Text className="text-warning mt-3 leading-6">
            {outcome.paused} Saved work is retained.
            {outcome.remainingTasks
              ? ` ${outcome.remainingTasks} tasks remain.`
              : ""}
          </Text>
        ) : null}
        {outcome.sourceWarnings?.length ? (
          <details className="mt-4 text-left">
            <summary className="focus-visible:ring-ring cursor-pointer rounded-md px-1 py-2 font-medium outline-none focus-visible:ring-2">
              Source data notes
            </summary>
            <Box
              as="ul"
              className="text-text-muted mt-2 max-h-52 list-disc space-y-2 overflow-auto pl-5"
            >
              {[...new Set(outcome.sourceWarnings)].map((note) => (
                <li key={note}>{note}</li>
              ))}
            </Box>
          </details>
        ) : null}
        {outcome.comments ? (
          <Text className="mt-2 leading-6" color="muted">
            Preserved or recognized {outcome.comments} native comments with
            source attribution.
          </Text>
        ) : null}
        {outcome.createdCustomFields ? (
          <Text className="mt-2 leading-6" color="muted">
            Created {outcome.createdCustomFields} typed custom fields.
          </Text>
        ) : null}
        {outcome.unresolvedComments ||
        outcome.unresolvedCustomFieldValues ||
        outcome.customFieldIssues?.length ? (
          <Box className="bg-warning/8 mt-4 rounded-xl p-4 text-left">
            <Text className="font-medium">
              Additional data needing attention
            </Text>
            <Text className="mt-1">
              {outcome.unresolvedComments ?? 0} comments and{" "}
              {outcome.unresolvedCustomFieldValues ?? 0} field values could not
              be applied.
            </Text>
            <Box as="ul" className="mt-2 max-h-40 space-y-1 overflow-auto">
              {[
                ...new Set([
                  ...(outcome.commentIssues ?? []),
                  ...(outcome.customFieldIssues ?? []),
                ]),
              ].map((issue) => (
                <li key={issue}>{issue}</li>
              ))}
            </Box>
          </Box>
        ) : null}
        <Button
          className="mt-5"
          color="tertiary"
          onClick={() => {
            const blob = new Blob(
              [
                JSON.stringify(
                  {
                    format: "fortyone-import-receipt",
                    version: 1,
                    generatedAt: new Date().toISOString(),
                    ...outcome,
                  },
                  null,
                  2,
                ),
              ],
              { type: "application/json" },
            );
            const url = URL.createObjectURL(blob);
            const link = document.createElement("a");
            link.href = url;
            link.download = "fortyone-import-receipt.json";
            link.click();
            URL.revokeObjectURL(url);
          }}
        >
          Download import receipt
        </Button>
        {outcome.unresolvedPeople ? (
          <Text className="mt-2 leading-6" color="muted">
            {outcome.unresolvedPeople} people could not be matched safely and
            were left unassigned. No invitations were sent.
          </Text>
        ) : null}
        {outcome.destinationConflicts ? (
          <Text className="text-warning mt-2 leading-6">
            {outcome.destinationConflicts} source objects or relationships had
            ambiguous or incompatible destination matches and were left
            unchanged. Review the mapping notes or adjust the source names
            before retrying.
          </Text>
        ) : null}
        {outcome.unresolvedAssociations ? (
          <Text className="text-warning mt-2 leading-6">
            {outcome.unresolvedAssociations}{" "}
            {outcome.unresolvedAssociations === 1
              ? `${storyTerm} relationship could`
              : `${storyTerm} relationships could`}{" "}
            not be preserved because a source or destination {storyTerm} was
            unavailable, unselected, or in another team.
          </Text>
        ) : null}
        {outcome.unresolvedLinks ? (
          <Text className="text-warning mt-2 leading-6">
            {outcome.unresolvedLinks}{" "}
            {outcome.unresolvedLinks === 1
              ? `${storyTerm} link could`
              : `${storyTerm} links could`}{" "}
            not be preserved because the destination {storyTerm} or link service
            was unavailable. Existing links were not duplicated.
          </Text>
        ) : null}
        {outcome.failed ? (
          <Box className="bg-warning/8 mt-5 max-h-52 overflow-y-auto rounded-2xl p-4 text-left">
            <Text className="font-medium">
              {storyTermPluralCapitalized} needing attention
            </Text>
            <Box as="ul" className="text-text-muted mt-2 space-y-1">
              {outcome.items
                .filter((item) => item.error !== null)
                .map((item) => (
                  <li key={item.sourceKey}>
                    {item.sourceKey}: {item.error?.message}
                  </li>
                ))}
            </Box>
          </Box>
        ) : null}
      </Box>
    );
  }

  return (
    <Box aria-live="assertive" className="mx-auto max-w-xl text-center">
      <Box className="bg-danger/10 text-danger mx-auto flex h-16 w-16 items-center justify-center rounded-3xl">
        <WarningIcon className="h-8" />
      </Box>
      <Text as="h2" className="mt-5 text-2xl font-medium">
        The import paused
      </Text>
      <Text className="mt-2 leading-6" color="muted">
        {runError || "The import could not finish."} Retrying reuses exact
        structure matches and stable {storyTerm} source IDs.
      </Text>
    </Box>
  );
};
