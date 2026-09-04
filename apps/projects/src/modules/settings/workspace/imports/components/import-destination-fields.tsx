"use client";
import type { Dispatch, SetStateAction } from "react";
import { PlusIcon, TeamIcon } from "icons";
import { Box, Checkbox, ColorPicker, Flex, Input, Text } from "ui";
import type { Team } from "@/modules/teams/public/types";
import type { ImportStructureMode } from "../import-run-model";
import type { DestinationChoice } from "./import-wizard-model";
import { formatTeamCode, initialNewTeam } from "./import-wizard-model";
import { DestinationTeamPicker, SelectionCard } from "./import-team-controls";
import { useImportTerms } from "./use-import-terms";

export type ImportDestinationFieldsProps = {
  destination: DestinationChoice;
  hasAttemptedImport: boolean;
  teams: Team[];
  structureMode: ImportStructureMode;
  suggestedTeamName: string;
  setDestination: Dispatch<SetStateAction<DestinationChoice>>;
};
export const ImportDestinationFields = ({
  destination,
  hasAttemptedImport,
  teams,
  structureMode,
  suggestedTeamName,
  setDestination,
}: ImportDestinationFieldsProps) => {
  const { storyTerm } = useImportTerms();
  return (
    <>
      <Text className="mt-5 font-medium">
        {structureMode === "preserve"
          ? "Fallback destination"
          : "Destination team"}
      </Text>
      <Text className="mt-1 leading-6" color="muted">
        {structureMode === "preserve"
          ? "Work without a reliable source-team relationship will go here."
          : "Choose a team you belong to or create a focused destination."}
      </Text>

      <Box className="mt-3 grid gap-3 md:grid-cols-2">
        <SelectionCard
          description="Use a team you already belong to for imported work."
          disabled={hasAttemptedImport || teams.length === 0}
          icon={<TeamIcon />}
          label="Existing team"
          onClick={() => {
            setDestination({
              kind: "existing",
              teamId: teams[0]?.id ?? "",
            });
          }}
          selected={destination.kind === "existing"}
        />
        <SelectionCard
          description={`Create a team with its own workflow and ${storyTerm} sequence.`}
          disabled={hasAttemptedImport}
          icon={<PlusIcon />}
          label="Create a new team"
          onClick={() => {
            setDestination({
              ...initialNewTeam,
              code: formatTeamCode(suggestedTeamName),
              name: suggestedTeamName,
            });
          }}
          selected={destination.kind === "new"}
        />
      </Box>

      {destination.kind === "existing" ? (
        <DestinationTeamPicker
          disabled={hasAttemptedImport}
          onChange={(teamId) => {
            setDestination({ kind: "existing", teamId });
          }}
          teams={teams}
          value={destination.teamId}
        />
      ) : (
        <Box className="border-border bg-surface mt-5 rounded-xl border-[0.5px] p-5">
          <Box className="grid gap-4 md:grid-cols-[minmax(0,1fr)_8rem_7rem]">
            <Input
              className="text-base"
              disabled={hasAttemptedImport}
              label="Team name"
              maxLength={24}
              minLength={3}
              onChange={(event) => {
                const name = event.target.value;
                setDestination((current) =>
                  current.kind === "new"
                    ? {
                        ...current,
                        name,
                        code: current.code
                          ? current.code
                          : formatTeamCode(name),
                      }
                    : current,
                );
              }}
              placeholder="Product migration"
              required
              value={destination.name}
            />
            <Input
              className="text-base uppercase"
              disabled={hasAttemptedImport}
              label="Team code"
              maxLength={3}
              minLength={2}
              onChange={(event) => {
                setDestination((current) =>
                  current.kind === "new"
                    ? {
                        ...current,
                        code: formatTeamCode(event.target.value),
                      }
                    : current,
                );
              }}
              placeholder="PM"
              required
              value={destination.code}
            />
            <Box>
              <Text className="mb-2 font-medium">Team color</Text>
              <ColorPicker
                ariaLabel="Choose team color"
                className="h-12 w-12 rounded-xl"
                disabled={hasAttemptedImport}
                onChange={(color) => {
                  setDestination((current) =>
                    current.kind === "new" ? { ...current, color } : current,
                  );
                }}
                size="lg"
                value={destination.color}
              />
            </Box>
          </Box>
          <Flex
            align="start"
            className="border-border mt-4 border-t-[0.5px] pt-4"
            gap={3}
          >
            <Checkbox
              checked={destination.isPrivate}
              className="mt-1"
              disabled={hasAttemptedImport}
              id="import-private-team"
              onCheckedChange={(checked) => {
                setDestination((current) =>
                  current.kind === "new"
                    ? { ...current, isPrivate: checked === true }
                    : current,
                );
              }}
            />
            <label className="cursor-pointer" htmlFor="import-private-team">
              <Text className="font-medium">Private team</Text>
              <Text className="mt-0.5" color="muted">
                Only invited members can see its imported work.
              </Text>
            </label>
          </Flex>
        </Box>
      )}
    </>
  );
};
