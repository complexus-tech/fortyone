"use client";
import type { Dispatch, ReactNode, SetStateAction } from "react";
import { PlusIcon, TeamIcon } from "icons";
import { Box, Button, Text } from "ui";
import type { Team } from "@/modules/teams/public/types";
import type { ImportDraft } from "../schema";
import type { ImportStructureMode } from "../import-run-model";
import type { DestinationChoice } from "./import-wizard-model";
import { ImportPlanSummary, SourceTeamPreview } from "./import-graph-review";
import { ImportDestinationFields } from "./import-destination-fields";
import { SelectionCard } from "./import-team-controls";
import { useImportTerms } from "./use-import-terms";

export type ImportTeamsStepProps = {
  draft: ImportDraft | null;
  destination: DestinationChoice;
  hasAttemptedImport: boolean;
  teams: Team[];
  knownWorkspaceTeams: Team[];
  structureMode: ImportStructureMode;
  setStructureMode: (mode: ImportStructureMode) => void;
  setDestination: Dispatch<SetStateAction<DestinationChoice>>;
  suggestedTeamName: string;
  hasSelectedTeamScopedImport: boolean;
  hasPrivacyWideningRisk: boolean;
  privateSourceTeamCount: number;
  workspaceTeamsPending: boolean;
  workspaceTeamsError: boolean;
  conflictedSourceTeamIds: ReadonlySet<string>;
  joinedTeamIds: ReadonlySet<string>;
  hasSourceTeamConflict: boolean;
  objectiveTargetTeamIds: string[];
  objectivePreflightPending: boolean;
  objectivePreflightError: string;
  strategyPreflightPending: boolean;
  strategyPreflightError: string;
  retryObjectives: () => void;
  retryStrategy: () => void;
};
export const ImportTeamsStep = ({
  draft,
  destination,
  hasAttemptedImport,
  teams,
  knownWorkspaceTeams,
  structureMode,
  setStructureMode,
  setDestination,
  suggestedTeamName,
  hasSelectedTeamScopedImport,
  hasPrivacyWideningRisk,
  privateSourceTeamCount,
  workspaceTeamsPending,
  workspaceTeamsError,
  conflictedSourceTeamIds,
  joinedTeamIds,
  hasSourceTeamConflict,
  objectiveTargetTeamIds,
  objectivePreflightPending,
  objectivePreflightError,
  strategyPreflightPending,
  strategyPreflightError,
  retryObjectives,
  retryStrategy,
}: ImportTeamsStepProps) => {
  const {
    storyTermPlural,
    objectiveTerm,
    objectiveTermPlural,
    objectiveTermPluralCapitalized,
    sprintTermPlural,
  } = useImportTerms();
  let sourceTeamReview: ReactNode = null;
  if (structureMode === "preserve" && draft) {
    if (workspaceTeamsPending) {
      sourceTeamReview = (
        <Text className="mt-5" color="muted">
          Checking workspace teams for safe reuse…
        </Text>
      );
    } else if (workspaceTeamsError) {
      sourceTeamReview = (
        <Box className="bg-danger/8 mt-5 rounded-xl p-4">
          <Text className="text-danger font-medium">
            Workspace teams could not be checked
          </Text>
          <Text className="mt-1" color="muted">
            Try again before preserving source teams, or combine this import
            into one destination.
          </Text>
        </Box>
      );
    } else {
      sourceTeamReview = (
        <>
          <SourceTeamPreview
            conflictedSourceIds={conflictedSourceTeamIds}
            draft={draft}
            existingTeams={knownWorkspaceTeams}
            joinedTeamIds={joinedTeamIds}
          />
          {hasSourceTeamConflict ? (
            <Box className="bg-warning/8 mt-3 rounded-xl p-4">
              <Text className="font-medium">Team mapping needs review</Text>
              <Text className="mt-1" color="muted">
                Combine this import into one destination, or make the workspace
                team names and codes unambiguous before preserving source teams.
              </Text>
            </Box>
          ) : null}
        </>
      );
    }
  }

  return (
    <Box>
      <Text as="h2" className="text-xl font-medium">
        {hasSelectedTeamScopedImport
          ? "Where should this work live?"
          : "Review workspace-level objects"}
      </Text>
      <Text className="mt-1 leading-6" color="muted">
        {hasSelectedTeamScopedImport
          ? "Keep useful source structure or combine everything into one destination. No team will be created before the final review."
          : "Strategic pillars and global labels live at workspace level. This import does not need or create a destination team."}
      </Text>

      {draft ? <ImportPlanSummary draft={draft} /> : null}

      {hasAttemptedImport ? (
        <Box
          className="bg-warning/8 mt-5 rounded-xl p-4"
          id="import-destination-lock-note"
        >
          <Text className="font-medium">
            Import setup is locked for safe retries
          </Text>
          <Text className="mt-1 leading-6" color="muted">
            The destination, team structure, and privacy stay fixed after the
            first attempt. Upload a new file to change them.
          </Text>
        </Box>
      ) : null}

      {strategyPreflightPending ? (
        <Text className="mt-5" color="muted">
          Checking existing strategic pillars for safe reuse…
        </Text>
      ) : null}
      {strategyPreflightError ? (
        <Box className="bg-danger/8 mt-5 rounded-xl p-4">
          <Text className="text-danger font-medium">
            Strategic pillars could not be checked
          </Text>
          <Text className="mt-1" color="muted">
            Check again before continuing so the import cannot create a
            duplicate pillar or reuse an ambiguous match.
          </Text>
          <Button
            className="mt-3"
            color="tertiary"
            onClick={() => {
              retryStrategy();
            }}
            size="sm"
            variant="outline"
          >
            Check again
          </Button>
        </Box>
      ) : null}

      {hasSelectedTeamScopedImport ? (
        <>
          {draft?.teams.length ? (
            <>
              <Box className="mt-5 grid gap-3 md:grid-cols-2">
                <SelectionCard
                  description="Reuse or create source teams, joining matched teams when needed."
                  disabled={hasAttemptedImport}
                  icon={<TeamIcon />}
                  label="Preserve source teams"
                  onClick={() => {
                    setStructureMode("preserve");
                  }}
                  selected={structureMode === "preserve"}
                />
                <SelectionCard
                  description={`Put teams, ${objectiveTermPlural}, ${sprintTermPlural}, and ${storyTermPlural} into one team you choose.`}
                  disabled={hasAttemptedImport}
                  icon={<PlusIcon />}
                  label="Combine into one team"
                  onClick={() => {
                    setStructureMode("single");
                  }}
                  selected={structureMode === "single"}
                />
              </Box>
              {sourceTeamReview}
            </>
          ) : null}
          <ImportDestinationFields
            destination={destination}
            hasAttemptedImport={hasAttemptedImport}
            setDestination={setDestination}
            structureMode={structureMode}
            suggestedTeamName={suggestedTeamName}
            teams={teams}
          />

          {hasPrivacyWideningRisk ? (
            <Box className="bg-danger/8 mt-5 rounded-xl p-4">
              <Text className="text-danger font-medium">
                Private source work needs a private destination
              </Text>
              <Text className="mt-1 leading-6" color="muted">
                {privateSourceTeamCount}{" "}
                {privateSourceTeamCount === 1
                  ? "private source team is"
                  : "private source teams are"}{" "}
                included. Choose a private destination team or preserve the
                source teams before continuing, so this work is not exposed more
                broadly.
              </Text>
            </Box>
          ) : null}

          {objectiveTargetTeamIds.length > 0 && objectivePreflightPending ? (
            <Text className="mt-5" color="muted">
              Checking existing {objectiveTermPlural} for safe reuse…
            </Text>
          ) : null}
          {objectivePreflightError ? (
            <Box className="bg-danger/8 mt-5 rounded-xl p-4">
              <Text className="text-danger font-medium">
                {objectiveTermPluralCapitalized} could not be checked
              </Text>
              <Text className="mt-1" color="muted">
                Check again before continuing so the import cannot create a
                duplicate or reuse an incompatible {objectiveTerm}.
              </Text>
              <Button
                className="mt-3"
                color="tertiary"
                onClick={() => {
                  retryObjectives();
                }}
                size="sm"
                variant="outline"
              >
                Check again
              </Button>
            </Box>
          ) : null}
        </>
      ) : (
        <Box className="border-border bg-surface mt-5 rounded-xl border-[0.5px] p-4">
          <Text className="font-medium">No team changes</Text>
          <Text className="mt-1 leading-6" color="muted">
            The selected pillars will be created or safely reused in the
            workspace strategy map. Existing teams and memberships will stay
            unchanged.
          </Text>
        </Box>
      )}
    </Box>
  );
};
