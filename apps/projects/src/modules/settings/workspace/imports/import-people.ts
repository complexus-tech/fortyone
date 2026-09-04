import type { ImportPerson, ImportTask } from "./schema";
import type { RunImportInput } from "./import-run-model";
import type { ImportSelection } from "./import-selection";
import type { ImportTeamDestinations } from "./import-team-destinations";
import type { ImportDestinationContext } from "./import-destination-context";
import { addExistingImportMemberToTeam } from "./api";
import {
  analyzeImportPersonIdentityClaims,
  getImportPersonIdentityKey,
  getImportPersonSourceIdentityKey,
  resolveImportPerson,
} from "./execution";

export const getTaskImportPerson = (
  task: ImportTask,
  peopleBySourceId: Map<string, ImportPerson>,
): Pick<ImportPerson, "email" | "name"> => {
  const person = task.assigneePersonSourceId
    ? peopleBySourceId.get(task.assigneePersonSourceId)
    : undefined;
  return {
    email: person?.email ?? task.assigneeEmail,
    name: person?.name ?? task.assigneeName,
  };
};

export const prepareImportPeople = async (
  {
    ctx,
    draft,
    confirmedMemberIdsByIdentityKey,
    onProgress,
  }: Pick<
    RunImportInput,
    "ctx" | "draft" | "confirmedMemberIdsByIdentityKey" | "onProgress"
  >,
  {
    selectedTasks,
    selectedObjectives,
    importableKeyResults,
  }: Pick<
    ImportSelection,
    "selectedTasks" | "selectedObjectives" | "importableKeyResults"
  >,
  { getTargetTeamId }: Pick<ImportTeamDestinations, "getTargetTeamId">,
  {
    getTeamContext,
    workspaceMembers,
  }: Pick<ImportDestinationContext, "getTeamContext" | "workspaceMembers">,
) => {
  let addedMembershipCount = 0;
  const peopleBySourceId = new Map(
    draft.people.map((person) => [person.sourceId, person]),
  );
  const objectivesBySourceId = new Map(
    selectedObjectives.map((objective) => [objective.sourceId, objective]),
  );
  const identityUses: {
    identity: Pick<ImportPerson, "email" | "name"> | undefined;
    sourceId?: string | null;
    teamId: string;
  }[] = [];
  const addIdentityUse = (
    identity: Pick<ImportPerson, "email" | "name"> | undefined,
    teamId: string | undefined,
    sourceId?: string | null,
  ) => {
    if (!teamId || !getImportPersonIdentityKey(identity, sourceId)) return;
    identityUses.push({ identity, sourceId, teamId });
  };

  for (const person of draft.people) {
    for (const sourceTeamId of person.teamSourceIds) {
      addIdentityUse(person, getTargetTeamId(sourceTeamId), person.sourceId);
    }
  }
  for (const task of selectedTasks) {
    const teamId = getTargetTeamId(task.teamSourceId);
    addIdentityUse(
      getTaskImportPerson(task, peopleBySourceId),
      teamId,
      task.assigneePersonSourceId,
    );
    for (const personSourceId of task.collaboratorPersonSourceIds) {
      addIdentityUse(
        peopleBySourceId.get(personSourceId),
        teamId,
        personSourceId,
      );
    }
  }
  for (const objective of selectedObjectives) {
    const lead = objective.leadPersonSourceId
      ? peopleBySourceId.get(objective.leadPersonSourceId)
      : undefined;
    addIdentityUse(
      lead,
      getTargetTeamId(objective.teamSourceId),
      objective.leadPersonSourceId,
    );
  }
  for (const keyResult of importableKeyResults) {
    if (!keyResult.objectiveSourceId) continue;
    const objective = objectivesBySourceId.get(keyResult.objectiveSourceId);
    if (!objective) continue;
    const teamId = getTargetTeamId(objective.teamSourceId);
    if (!teamId) continue;
    const personSourceIds = [
      keyResult.leadPersonSourceId,
      ...keyResult.contributorPersonSourceIds,
    ].filter((value): value is string => Boolean(value));
    for (const personSourceId of personSourceIds) {
      addIdentityUse(
        peopleBySourceId.get(personSourceId),
        teamId,
        personSourceId,
      );
    }
  }

  const { canonicalIdentities, conflictedIdentityKeys } =
    analyzeImportPersonIdentityClaims(identityUses);
  const getCanonicalIdentity = (use: (typeof identityUses)[number]) => {
    const identityKey = getImportPersonSourceIdentityKey(
      use.identity,
      use.sourceId,
    );
    if (!identityKey || conflictedIdentityKeys.has(identityKey)) {
      return undefined;
    }
    return canonicalIdentities.get(identityKey) ?? use.identity;
  };
  const resolveApprovedPerson = (
    use: (typeof identityUses)[number],
    members: Parameters<typeof resolveImportPerson>[1],
  ) => {
    const identityKey = getImportPersonSourceIdentityKey(
      use.identity,
      use.sourceId,
    );
    if (identityKey && confirmedMemberIdsByIdentityKey.has(identityKey)) {
      const selectedMemberId = confirmedMemberIdsByIdentityKey.get(identityKey);
      if (!selectedMemberId) return undefined;
      const selectedMember = members.find(
        (member) =>
          member.id === selectedMemberId && member.isActive && !member.isSystem,
      );
      return selectedMember
        ? {
            matchedBy: "fullName" as const,
            member: selectedMember,
            requiresReview: true,
          }
        : undefined;
    }

    const identity = getCanonicalIdentity(use);
    const resolution = identity
      ? resolveImportPerson(identity, members)
      : undefined;
    if (resolution && !resolution.requiresReview) return resolution;
    return undefined;
  };

  const addedMembershipKeys = new Set<string>();
  for (const use of identityUses) {
    const resolution = resolveApprovedPerson(use, workspaceMembers);
    if (!resolution) continue;
    const context = getTeamContext(use.teamId);
    if (context.members.some((member) => member.id === resolution.member.id)) {
      continue;
    }
    const membershipKey = `${use.teamId}:${resolution.member.id}`;
    if (addedMembershipKeys.has(membershipKey)) continue;
    // eslint-disable-next-line no-await-in-loop -- Membership changes are reviewed and applied once per team/member pair before dependent entities.
    await addExistingImportMemberToTeam(use.teamId, resolution.member.id, ctx);
    context.members.push(resolution.member);
    addedMembershipKeys.add(membershipKey);
    addedMembershipCount += 1;
  }
  onProgress(25);

  const unresolvedPeople = new Set<string>();
  for (const use of identityUses) {
    const identityKey = getImportPersonSourceIdentityKey(
      use.identity,
      use.sourceId,
    );
    if (!identityKey) continue;
    const resolution = resolveApprovedPerson(
      use,
      getTeamContext(use.teamId).members,
    );
    if (!resolution) {
      unresolvedPeople.add(identityKey);
    }
  }
  const resolveReviewedPerson = (
    person: Pick<ImportPerson, "email" | "name"> | undefined,
    teamId: string,
    sourceId?: string | null,
  ) => {
    const identityKey = getImportPersonSourceIdentityKey(person, sourceId);
    const resolution = resolveApprovedPerson(
      { identity: person, sourceId, teamId },
      getTeamContext(teamId).members,
    );
    if (!resolution) {
      if (identityKey) unresolvedPeople.add(identityKey);
      return undefined;
    }
    return resolution.member;
  };

  return {
    peopleBySourceId,
    resolveReviewedPerson,
    unresolvedPeople,
    addedMembershipCount,
  };
};

export type ImportPeople = Awaited<ReturnType<typeof prepareImportPeople>>;
