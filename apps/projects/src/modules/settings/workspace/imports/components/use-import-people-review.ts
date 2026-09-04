import { useMemo, useState } from "react";
import type { Member } from "@/types/member";
import { resolveImportPerson, suggestImportPersonMember } from "../execution";
import type { ImportIdentityReview } from "./import-identity-review";

export const useImportPeopleReview = (
  importIdentities: ImportIdentityReview[],
  workspaceMembersForReview: Member[],
) => {
  const [selectedMemberIdsByIdentityKey, setSelectedMemberIdsByIdentityKey] =
    useState<Map<string, string | null>>(new Map());
  const [lockedMemberIdsByIdentityKey, setLockedMemberIdsByIdentityKey] =
    useState<Map<string, string | null> | null>(null);
  const peopleMappingPreview = useMemo(
    () =>
      importIdentities.map((identity) => {
        const resolution = identity.hasConflictingClaims
          ? undefined
          : resolveImportPerson(identity, workspaceMembersForReview);
        const suggestion = identity.hasConflictingClaims
          ? undefined
          : suggestImportPersonMember(identity, workspaceMembersForReview);
        return {
          identity,
          identityKey: identity.identityKey,
          suggestedMember: resolution?.member ?? suggestion?.member,
        };
      }),
    [importIdentities, workspaceMembersForReview],
  );
  const confirmedMemberIdsByIdentityKey = useMemo(() => {
    const eligibleMemberIds = new Set(
      workspaceMembersForReview
        .filter((member) => member.isActive && !member.isSystem)
        .map((member) => member.id),
    );
    const selected = new Map<string, string | null>();
    for (const preview of peopleMappingPreview) {
      const memberId = selectedMemberIdsByIdentityKey.has(preview.identityKey)
        ? selectedMemberIdsByIdentityKey.get(preview.identityKey)
        : preview.suggestedMember?.id;
      selected.set(
        preview.identityKey,
        memberId && eligibleMemberIds.has(memberId) ? memberId : null,
      );
    }
    return selected;
  }, [
    peopleMappingPreview,
    selectedMemberIdsByIdentityKey,
    workspaceMembersForReview,
  ]);
  const hasImportIdentities = importIdentities.length > 0;
  const reviewedMemberIdsByIdentityKey =
    lockedMemberIdsByIdentityKey ?? confirmedMemberIdsByIdentityKey;

  const selectMember = (identityKey: string, memberId: string | null) => {
    setSelectedMemberIdsByIdentityKey((current) => {
      const next = new Map(current);
      next.set(identityKey, memberId);
      return next;
    });
  };
  const lockMappings = () => {
    const memberIdsByIdentityKey =
      lockedMemberIdsByIdentityKey ??
      new Map<string, string | null>(confirmedMemberIdsByIdentityKey);
    if (!lockedMemberIdsByIdentityKey)
      setLockedMemberIdsByIdentityKey(memberIdsByIdentityKey);
    return memberIdsByIdentityKey;
  };
  const reset = () => {
    setSelectedMemberIdsByIdentityKey(new Map());
    setLockedMemberIdsByIdentityKey(null);
  };
  return {
    hasImportIdentities,
    peopleMappingPreview,
    reviewedMemberIdsByIdentityKey,
    selectMember,
    lockMappings,
    reset,
  };
};
