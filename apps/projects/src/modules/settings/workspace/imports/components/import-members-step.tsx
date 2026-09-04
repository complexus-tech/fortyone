"use client";
import { Box, Button, Flex, Text } from "ui";
import type { Member } from "@/types/member";
import type { ImportIdentityReview } from "./import-identity-review";
import { ImportMemberPicker } from "./import-member-picker";

export type ImportMembersStepProps = {
  hasAttemptedImport: boolean;
  hasImportIdentities: boolean;
  peoplePreflightPending: boolean;
  peoplePreflightError: string;
  retryPeople: () => void;
  peopleMappingPreview: {
    identity: ImportIdentityReview;
    identityKey: string;
    suggestedMember: Member | undefined;
  }[];
  reviewedMemberIdsByIdentityKey: ReadonlyMap<string, string | null>;
  workspaceMembersForReview: Member[];
  selectMember: (identityKey: string, memberId: string | null) => void;
};
export const ImportMembersStep = ({
  hasAttemptedImport,
  hasImportIdentities,
  peoplePreflightPending,
  peoplePreflightError,
  retryPeople,
  peopleMappingPreview,
  reviewedMemberIdsByIdentityKey,
  workspaceMembersForReview,
  selectMember,
}: ImportMembersStepProps) => (
  <Box>
    <Text as="h2" className="text-xl font-medium">
      Map members
    </Text>
    <Text className="mt-1 leading-6" color="muted">
      Review who imported assignments belong to. Likely matches are selected
      automatically, and anyone can remain unassigned.
    </Text>

    {hasAttemptedImport ? (
      <Text className="mt-4" color="muted">
        Member mappings are locked for safe retries.
      </Text>
    ) : null}

    {peoplePreflightPending ? (
      <Text className="mt-5" color="muted">
        Checking source identities against workspace members…
      </Text>
    ) : null}

    {peoplePreflightError ? (
      <Box className="bg-danger/8 mt-5 rounded-xl p-4">
        <Text className="text-danger font-medium">
          Workspace members could not be checked
        </Text>
        <Text className="mt-1" color="muted">
          Check again before continuing so assignments and team memberships
          remain reviewable.
        </Text>
        <Button
          className="mt-3"
          color="tertiary"
          onClick={() => {
            retryPeople();
          }}
          size="sm"
          variant="outline"
        >
          Check again
        </Button>
      </Box>
    ) : null}

    {!hasImportIdentities ? (
      <Text className="mt-4" color="muted">
        No source members need mapping for this import.
      </Text>
    ) : null}

    {peopleMappingPreview.length > 0 &&
    !peoplePreflightPending &&
    !peoplePreflightError ? (
      <Box className="border-border/70 bg-surface/60 mt-4 overflow-hidden rounded-xl border-[0.5px]">
        <Flex
          align="center"
          className="border-border border-b-[0.5px] px-4 py-2.5"
          justify="between"
        >
          <Text className="font-medium">Source members</Text>
          <Text color="muted">{peopleMappingPreview.length} detected</Text>
        </Flex>
        <Box className="divide-border max-h-72 divide-y overflow-y-auto">
          {peopleMappingPreview.map(
            ({ identity, identityKey, suggestedMember }) => {
              const sourceLabel =
                identity.name ||
                identity.email ||
                `Source identity ${identity.sourceId ?? "unknown"}`;
              const selectedMemberId =
                reviewedMemberIdsByIdentityKey.get(identityKey);
              const selectedMember = workspaceMembersForReview.find(
                (member) => member.id === selectedMemberId,
              );
              return (
                <Flex
                  align="center"
                  className="flex-col items-stretch px-4 py-2.5 md:flex-row md:items-center"
                  gap={3}
                  justify="between"
                  key={identityKey}
                >
                  <Text className="min-w-0 truncate font-medium">
                    {sourceLabel}
                  </Text>
                  <ImportMemberPicker
                    disabled={hasAttemptedImport}
                    members={workspaceMembersForReview}
                    onChange={(memberId) => {
                      selectMember(identityKey, memberId);
                    }}
                    suggestedMemberId={suggestedMember?.id}
                    value={selectedMember?.id ?? null}
                  />
                </Flex>
              );
            },
          )}
        </Box>
      </Box>
    ) : null}
  </Box>
);
