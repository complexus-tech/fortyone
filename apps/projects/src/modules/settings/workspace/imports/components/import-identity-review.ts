import type { ImportDraft } from "../schema";
import {
  analyzeImportPersonIdentityClaims,
  getImportPersonSourceIdentityKey,
} from "../execution";

export type ImportIdentityReview = {
  email: string | null;
  hasConflictingClaims: boolean;
  identityKey: string;
  name: string | null;
  sourceId: string | null;
};

type ImportIdentityDraft = Pick<
  ImportDraft,
  "keyResults" | "objectives" | "people" | "tasks"
>;

export const collectImportIdentities = (
  draft: ImportIdentityDraft | null,
): ImportIdentityReview[] => {
  if (!draft) return [];
  const identities = new Map<
    string,
    Omit<ImportIdentityReview, "identityKey">
  >();
  const identityClaims: {
    identity: Pick<ImportIdentityReview, "email" | "name"> | undefined;
    sourceId: string | null;
  }[] = [];
  const peopleBySourceId = new Map(
    draft.people.map((person) => [person.sourceId, person]),
  );
  const addIdentity = (
    identity: Pick<ImportIdentityReview, "email" | "name"> | undefined,
    sourceId: string | null,
  ) => {
    const identityKey = getImportPersonSourceIdentityKey(identity, sourceId);
    if (!identityKey) return;
    identityClaims.push({ identity, sourceId });
    if (identities.has(identityKey)) return;
    identities.set(identityKey, {
      email: identity?.email ?? null,
      hasConflictingClaims: false,
      name: identity?.name ?? null,
      sourceId,
    });
  };

  for (const person of draft.people) {
    addIdentity(person, person.sourceId);
  }
  for (const task of draft.tasks) {
    const referencedAssignee = task.assigneePersonSourceId
      ? peopleBySourceId.get(task.assigneePersonSourceId)
      : undefined;
    addIdentity(
      {
        email: referencedAssignee?.email ?? task.assigneeEmail,
        name: referencedAssignee?.name ?? task.assigneeName,
      },
      task.assigneePersonSourceId,
    );
    for (const sourceId of task.collaboratorPersonSourceIds) {
      addIdentity(peopleBySourceId.get(sourceId), sourceId);
    }
  }
  for (const objective of draft.objectives) {
    if (objective.leadPersonSourceId) {
      addIdentity(
        peopleBySourceId.get(objective.leadPersonSourceId),
        objective.leadPersonSourceId,
      );
    }
  }
  for (const keyResult of draft.keyResults) {
    for (const sourceId of [
      keyResult.leadPersonSourceId,
      ...keyResult.contributorPersonSourceIds,
    ]) {
      if (sourceId) addIdentity(peopleBySourceId.get(sourceId), sourceId);
    }
  }

  const { canonicalIdentities, conflictedIdentityKeys } =
    analyzeImportPersonIdentityClaims(identityClaims);
  return [...identities.entries()].map(([identityKey, firstClaim]) => {
    const hasConflictingClaims = conflictedIdentityKeys.has(identityKey);
    const canonicalIdentity = hasConflictingClaims
      ? undefined
      : canonicalIdentities.get(identityKey);
    return {
      ...firstClaim,
      email: canonicalIdentity?.email ?? firstClaim.email,
      hasConflictingClaims,
      identityKey,
      name: canonicalIdentity?.name ?? firstClaim.name,
    };
  });
};
