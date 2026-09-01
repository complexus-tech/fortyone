import type { WorkspaceCtx } from "@/lib/http";
import type { Team } from "@/modules/teams/types";
import type { StoryAssociationType } from "@/shared/story/types";
import type { ImportDraft, ImportPerson, ImportTask } from "./schema";
import { JIRA_ISSUE_KEY_PATTERN, normalizeImportTaskLinks } from "./schema";
import type { ImportStoryResult } from "./api";
import {
  addExistingImportMemberToTeam,
  alignImportObjectiveToPillar,
  buildImportStoryRequests,
  createImportKeyResults,
  createImportLabel,
  createImportObjective,
  createImportSprint,
  createImportStrategicPillar,
  createImportStoryAssociation,
  createImportStoryLink,
  createImportTeam,
  getImportObjectiveKeyResults,
  getImportObjectiveStatuses,
  getImportStoryAssociations,
  getImportStoryCollaboratorIds,
  getImportStoryLinks,
  getImportStrategyMap,
  getImportTeamLabels,
  getImportTeamMembers,
  getImportTeamObjectives,
  getImportTeamSprints,
  getImportTeamStatuses,
  getImportWorkspaceLabels,
  getImportWorkspaceMembers,
  importStoriesBatch,
  updateImportStoryCollaborators,
} from "./api";
import {
  deriveImportTeamCode,
  deriveImportTeamColor,
  analyzeImportPersonIdentityClaims,
  getBoundedImportSourceKey,
  getImportPersonIdentityKey,
  getImportPersonSourceIdentityKey,
  isValidImportDateRange,
  resolveImportEntityNameMatch,
  resolveImportPerson,
  resolveImportStatus,
  toImportStoryPayload,
} from "./execution";

export type ImportStructureMode = "preserve" | "single";

export type ImportRunResult = {
  created: number;
  replayed: number;
  failed: number;
  items: ImportStoryResult[];
  teamId: string | null;
  createdTeams: number;
  createdStrategicPillars: number;
  createdObjectives: number;
  createdKeyResults: number;
  createdSprints: number;
  createdLabels: number;
  createdLinks: number;
  addedMemberships: number;
  appliedCollaborators: number;
  createdAssociations: number;
  alignedObjectives: number;
  destinationConflicts: number;
  unresolvedAssociations: number;
  unresolvedLinks: number;
  unresolvedPeople: number;
};

export type RunImportInput = {
  actorUserId: string;
  draft: ImportDraft;
  selectedTaskIndexes: ReadonlySet<number>;
  selectedObjectiveSourceIds: ReadonlySet<string>;
  selectedStrategicPillarSourceIds: ReadonlySet<string>;
  structureMode: ImportStructureMode;
  fallbackTeamId: string | null;
  fallbackTeamIsPrivate: boolean;
  fallbackTeamName: string;
  fallbackTeamCode: string;
  fallbackTeamIsNew: boolean;
  fallbackTeamCreated: boolean;
  existingTeams: readonly Team[];
  forceCreateObjectiveSourceIds: ReadonlySet<string>;
  joinedTeamIds: ReadonlySet<string>;
  confirmedMemberIdsByIdentityKey: ReadonlyMap<string, string | null>;
  sourceTeamCache: Map<string, string>;
  sourceObjectiveCache: Map<string, { id: string; teamId: string }>;
  ctx: WorkspaceCtx;
  onProgress: (progress: number) => void;
};

const normalizeImportMatch = (value: string) =>
  value.normalize("NFKC").trim().toLocaleLowerCase().replace(/\s+/g, " ");

const toImportTeamName = (value: string) => {
  const name = value.trim().replace(/\s+/g, " ").slice(0, 24);
  if (name.length >= 3) return name;
  return `${name || "New"} team`.slice(0, 24);
};

export const resolveImportSourceTeam = (
  sourceTeam: ImportDraft["teams"][number],
  existingTeams: readonly Team[],
) => {
  const privacyCompatibleTeams = existingTeams.filter(
    (team) => team.isPrivate === sourceTeam.isPrivate,
  );
  const sourceName = normalizeImportMatch(toImportTeamName(sourceTeam.name));
  const nameMatches = privacyCompatibleTeams.filter(
    (team) => normalizeImportMatch(team.name) === sourceName,
  );
  const sourceCode = sourceTeam.code
    ?.trim()
    .toUpperCase()
    .replace(/[^A-Z0-9]/g, "")
    .slice(0, 3);
  if (!sourceCode) {
    if (nameMatches.length === 1) {
      return { kind: "unique" as const, team: nameMatches[0] };
    }
    return nameMatches.length > 1
      ? { kind: "ambiguous" as const }
      : { kind: "none" as const };
  }
  const codeMatches = privacyCompatibleTeams.filter(
    (team) => team.code.trim().toUpperCase() === sourceCode,
  );
  if (nameMatches.length === 0) {
    return codeMatches.length > 0
      ? { kind: "ambiguous" as const }
      : { kind: "none" as const };
  }
  if (nameMatches.length === 1) {
    if (codeMatches.length === 0) {
      return { kind: "unique" as const, team: nameMatches[0] };
    }
    return codeMatches.length === 1 && codeMatches[0].id === nameMatches[0].id
      ? { kind: "unique" as const, team: nameMatches[0] }
      : { kind: "ambiguous" as const };
  }
  if (
    codeMatches.length === 1 &&
    nameMatches.some((team) => team.id === codeMatches[0].id)
  ) {
    return { kind: "unique" as const, team: codeMatches[0] };
  }
  return { kind: "ambiguous" as const };
};

const getTaskImportPerson = (
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

const chunkImportItems = <T>(items: T[], size: number) => {
  const chunks: T[][] = [];
  for (let index = 0; index < items.length; index += size) {
    chunks.push(items.slice(index, index + size));
  }
  return chunks;
};

export const getImportSourceTeamDestination = (
  sourceTeam: ImportDraft["teams"][number],
) => ({
  code: deriveImportTeamCode(sourceTeam),
  color: deriveImportTeamColor(sourceTeam),
  isPrivate: sourceTeam.isPrivate,
  name: toImportTeamName(sourceTeam.name),
});

const toImportEntityColor = (value: string | null) =>
  value && /^#[0-9A-Fa-f]{6}$/.test(value) ? value.toUpperCase() : "#697386";

const toOptionalImportEntityColor = (value: string | null) =>
  value && /^#[0-9A-Fa-f]{6}$/.test(value) ? value.toUpperCase() : undefined;

const mergeImportStoryCollaborators = async (
  storyId: string,
  importedCollaboratorIds: string[],
  ctx: WorkspaceCtx,
) => {
  const existingCollaboratorIds = await getImportStoryCollaboratorIds(
    storyId,
    ctx,
  );
  const existingIds = new Set(existingCollaboratorIds);
  const addedIds = [...new Set(importedCollaboratorIds)].filter(
    (memberId) => !existingIds.has(memberId),
  );
  if (addedIds.length === 0) return 0;

  await updateImportStoryCollaborators(
    storyId,
    [...existingIds, ...addedIds],
    ctx,
  );
  return addedIds.length;
};

export const getCanonicalImportAssociation = (
  sourceId: string,
  targetId: string,
  sourceType: ImportTask["associations"][number]["type"],
): {
  fromId: string;
  toId: string;
  type: StoryAssociationType;
} => {
  if (sourceType === "blocked_by") {
    return { fromId: targetId, toId: sourceId, type: "blocking" };
  }
  if (sourceType === "blocks") {
    return { fromId: sourceId, toId: targetId, type: "blocking" };
  }
  const [fromId, toId] =
    sourceId < targetId ? [sourceId, targetId] : [targetId, sourceId];
  return { fromId, toId, type: sourceType };
};

export const getImportAssociationKey = ({
  fromId,
  toId,
  type,
}: {
  fromId: string;
  toId: string;
  type: StoryAssociationType;
}) => {
  if (type === "blocking") return `${type}\u0000${fromId}\u0000${toId}`;
  const [firstId, secondId] = fromId < toId ? [fromId, toId] : [toId, fromId];
  return `${type}\u0000${firstId}\u0000${secondId}`;
};

export const runImport = async ({
  actorUserId,
  confirmedMemberIdsByIdentityKey,
  ctx,
  draft,
  existingTeams,
  fallbackTeamCode,
  fallbackTeamCreated,
  fallbackTeamId,
  fallbackTeamIsPrivate,
  fallbackTeamIsNew,
  fallbackTeamName,
  joinedTeamIds,
  onProgress,
  selectedObjectiveSourceIds,
  selectedStrategicPillarSourceIds,
  selectedTaskIndexes,
  sourceObjectiveCache,
  sourceTeamCache,
  structureMode,
}: RunImportInput): Promise<ImportRunResult> => {
  const selectedTasks = draft.tasks.filter((_, index) =>
    selectedTaskIndexes.has(index),
  );
  const selectedObjectives = draft.objectives.filter((objective) =>
    selectedObjectiveSourceIds.has(objective.sourceId),
  );
  const selectedStrategicPillars = draft.strategicPillars.filter((pillar) =>
    selectedStrategicPillarSourceIds.has(pillar.sourceId),
  );
  const importableKeyResults = draft.keyResults.filter(
    (keyResult) =>
      keyResult.objectiveSourceId &&
      selectedObjectiveSourceIds.has(keyResult.objectiveSourceId) &&
      keyResult.measurementType !== null &&
      keyResult.startValue !== null &&
      keyResult.currentValue !== null &&
      keyResult.targetValue !== null &&
      keyResult.startDate !== null &&
      keyResult.endDate !== null &&
      isValidImportDateRange(keyResult.startDate, keyResult.endDate),
  );
  const importableSprints = draft.sprints.filter(
    (sprint) =>
      sprint.startDate !== null &&
      sprint.endDate !== null &&
      isValidImportDateRange(sprint.startDate, sprint.endDate),
  );
  const hasTeamScopedImport = Boolean(
    selectedTasks.length > 0 ||
      selectedObjectives.length > 0 ||
      importableSprints.length > 0 ||
      draft.labels.some((label) => label.teamSourceId !== null) ||
      (structureMode === "preserve" && draft.teams.length > 0),
  );
  if (!fallbackTeamId && hasTeamScopedImport) {
    throw new Error("A destination team is required for team-scoped work");
  }
  if (structureMode === "single" && !fallbackTeamIsPrivate) {
    const selectedSourceTeamIds = new Set<string>();
    const addSourceTeamId = (sourceTeamId: string | null) => {
      if (sourceTeamId) selectedSourceTeamIds.add(sourceTeamId);
    };
    for (const task of selectedTasks) addSourceTeamId(task.teamSourceId);
    for (const objective of selectedObjectives) {
      addSourceTeamId(objective.teamSourceId);
    }
    for (const sprint of importableSprints) {
      addSourceTeamId(sprint.teamSourceId);
    }
    for (const label of draft.labels) addSourceTeamId(label.teamSourceId);
    const widensPrivateSourceWork = draft.teams.some(
      (team) => team.isPrivate && selectedSourceTeamIds.has(team.sourceId),
    );
    if (widensPrivateSourceWork) {
      throw new Error(
        "Private source work cannot be combined into a public destination team",
      );
    }
  }

  let createdTeams = fallbackTeamCreated ? 1 : 0;
  let createdStrategicPillars = 0;
  let createdObjectives = 0;
  let createdKeyResults = 0;
  let createdSprints = 0;
  let createdLabels = 0;
  let createdLinks = 0;
  let createdAssociations = 0;
  let alignedObjectives = 0;
  let addedMembershipCount = 0;
  let appliedCollaborators = 0;
  let destinationConflicts = 0;
  let unresolvedAssociations = 0;
  let unresolvedLinks = 0;
  const unavailableTeamCodes = new Set(
    existingTeams.map((team) => team.code.trim().toUpperCase()),
  );
  if (fallbackTeamIsNew && fallbackTeamId) {
    unavailableTeamCodes.add(fallbackTeamCode.trim().toUpperCase());
  }
  const teamIdsBySourceId = new Map<string, string>();
  const conflictedTeamSourceIds = new Set<string>();
  const actorTeamIds = new Set(joinedTeamIds);

  if (structureMode === "preserve") {
    const sourceTeamPlans = new Map<
      string,
      {
        cachedTeamId: string | undefined;
        collisionKey: string | null;
        matchesNewFallback: boolean;
        resolution: ReturnType<typeof resolveImportSourceTeam> | undefined;
      }
    >();
    const destinationGroups = new Map<string, string[]>();
    for (const sourceTeam of draft.teams) {
      const cachedTeamId = sourceTeamCache.get(sourceTeam.sourceId);
      const resolution = cachedTeamId
        ? undefined
        : resolveImportSourceTeam(sourceTeam, existingTeams);
      const matchesNewFallback =
        resolution?.kind === "none" &&
        fallbackTeamId !== null &&
        fallbackTeamIsNew &&
        sourceTeam.isPrivate === fallbackTeamIsPrivate &&
        normalizeImportMatch(toImportTeamName(sourceTeam.name)) ===
          normalizeImportMatch(fallbackTeamName);
      let collisionKey: string | null = null;
      if (cachedTeamId) {
        collisionKey = `existing:${cachedTeamId}`;
      } else if (resolution?.kind === "unique") {
        collisionKey = `existing:${resolution.team.id}`;
      } else if (resolution?.kind === "none") {
        collisionKey = matchesNewFallback
          ? `existing:${fallbackTeamId}`
          : `new-name:${sourceTeam.isPrivate}:${normalizeImportMatch(
              toImportTeamName(sourceTeam.name),
            )}`;
      }
      sourceTeamPlans.set(sourceTeam.sourceId, {
        cachedTeamId,
        collisionKey,
        matchesNewFallback,
        resolution,
      });
      if (!collisionKey) continue;
      const sourceIds = destinationGroups.get(collisionKey) ?? [];
      sourceIds.push(sourceTeam.sourceId);
      destinationGroups.set(collisionKey, sourceIds);
    }
    const collidedSourceTeamIds = new Set<string>();
    for (const sourceIds of destinationGroups.values()) {
      if (sourceIds.length < 2) continue;
      destinationConflicts += sourceIds.length;
      for (const sourceId of sourceIds) {
        collidedSourceTeamIds.add(sourceId);
        conflictedTeamSourceIds.add(sourceId);
      }
    }

    for (const sourceTeam of draft.teams) {
      if (collidedSourceTeamIds.has(sourceTeam.sourceId)) continue;
      const plan = sourceTeamPlans.get(sourceTeam.sourceId);
      if (!plan) throw new Error("Unable to resolve an import source team");
      const { cachedTeamId, matchesNewFallback, resolution } = plan;
      if (cachedTeamId) {
        teamIdsBySourceId.set(sourceTeam.sourceId, cachedTeamId);
        continue;
      }

      if (resolution?.kind === "unique") {
        if (!actorTeamIds.has(resolution.team.id)) {
          // eslint-disable-next-line no-await-in-loop -- Importer access must exist before reading or mutating this destination team.
          await addExistingImportMemberToTeam(
            resolution.team.id,
            actorUserId,
            ctx,
          );
          actorTeamIds.add(resolution.team.id);
          addedMembershipCount += 1;
        }
        teamIdsBySourceId.set(sourceTeam.sourceId, resolution.team.id);
        continue;
      }
      if (resolution?.kind === "ambiguous") {
        conflictedTeamSourceIds.add(sourceTeam.sourceId);
        destinationConflicts += 1;
        continue;
      }

      if (matchesNewFallback && fallbackTeamId) {
        teamIdsBySourceId.set(sourceTeam.sourceId, fallbackTeamId);
        continue;
      }

      const code = deriveImportTeamCode(sourceTeam, unavailableTeamCodes);
      // eslint-disable-next-line no-await-in-loop -- Team creation is ordered so generated codes stay unique and progress remains deterministic.
      const response = await createImportTeam(
        {
          code,
          color: deriveImportTeamColor(sourceTeam),
          isPrivate: sourceTeam.isPrivate,
          name: toImportTeamName(sourceTeam.name),
        },
        ctx,
      );
      if (response.error?.message || !response.data?.id) {
        throw new Error(
          response.error?.message ||
            `Unable to create source team ${sourceTeam.name}`,
        );
      }
      unavailableTeamCodes.add(code);
      sourceTeamCache.set(sourceTeam.sourceId, response.data.id);
      teamIdsBySourceId.set(sourceTeam.sourceId, response.data.id);
      createdTeams += 1;
    }
  }
  onProgress(12);

  const getTargetTeamId = (sourceTeamId: string | null) => {
    if (structureMode !== "preserve" || !sourceTeamId) {
      return fallbackTeamId ?? undefined;
    }
    if (conflictedTeamSourceIds.has(sourceTeamId)) return undefined;
    return teamIdsBySourceId.get(sourceTeamId) ?? fallbackTeamId ?? undefined;
  };
  const targetTeamIds = new Set<string>([
    ...(fallbackTeamId ? [fallbackTeamId] : []),
    ...teamIdsBySourceId.values(),
  ]);
  const shouldLoadTeamContexts = targetTeamIds.size > 0;
  const shouldLoadWorkspaceLabels = draft.labels.some(
    (label) => label.teamSourceId === null,
  );
  const [
    objectiveStatuses,
    workspaceMembers,
    workspaceLabels,
    teamContextEntries,
  ] = await Promise.all([
    shouldLoadTeamContexts
      ? getImportObjectiveStatuses(ctx)
      : Promise.resolve([]),
    shouldLoadTeamContexts
      ? getImportWorkspaceMembers(ctx)
      : Promise.resolve([]),
    shouldLoadWorkspaceLabels
      ? getImportWorkspaceLabels(ctx)
      : Promise.resolve([]),
    shouldLoadTeamContexts
      ? Promise.all(
          [...targetTeamIds].map(async (teamId) => {
            const [statuses, members, objectives, labels, sprints] =
              await Promise.all([
                getImportTeamStatuses(teamId, ctx),
                getImportTeamMembers(teamId, ctx),
                getImportTeamObjectives(teamId, ctx),
                getImportTeamLabels(teamId, ctx),
                getImportTeamSprints(teamId, ctx),
              ]);
            return {
              teamId,
              statuses,
              members,
              objectives,
              labels,
              sprints,
            };
          }),
        )
      : Promise.resolve([]),
  ]);
  const teamContexts = new Map(
    teamContextEntries.map((entry) => [entry.teamId, entry]),
  );
  const getTeamContext = (teamId: string) => {
    const context = teamContexts.get(teamId);
    if (!context) throw new Error("Unable to resolve an import team");
    return context;
  };
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

  const shouldLoadStrategyMap = selectedStrategicPillars.length > 0;
  const strategyMap = shouldLoadStrategyMap
    ? await getImportStrategyMap(ctx)
    : { description: null, pillars: [], ultimateGoal: "" };
  const pillarMappings = new Map<
    string,
    (typeof strategyMap.pillars)[number]
  >();
  const sourcePillarsByName = new Map<
    string,
    typeof selectedStrategicPillars
  >();
  for (const pillar of selectedStrategicPillars) {
    const normalizedName = normalizeImportMatch(pillar.name);
    const matches = sourcePillarsByName.get(normalizedName) ?? [];
    matches.push(pillar);
    sourcePillarsByName.set(normalizedName, matches);
  }
  const duplicateSourcePillarIds = new Set<string>();
  for (const matches of sourcePillarsByName.values()) {
    if (matches.length < 2) continue;
    destinationConflicts += matches.length;
    for (const pillar of matches) {
      duplicateSourcePillarIds.add(pillar.sourceId);
    }
  }
  for (const pillar of selectedStrategicPillars) {
    if (duplicateSourcePillarIds.has(pillar.sourceId)) continue;
    const existing = resolveImportEntityNameMatch(
      pillar.name,
      strategyMap.pillars,
    );
    if (existing.kind === "ambiguous") {
      destinationConflicts += 1;
      continue;
    }
    if (existing.kind === "unique") {
      pillarMappings.set(pillar.sourceId, existing.entity);
      continue;
    }
    // eslint-disable-next-line no-await-in-loop -- Pillars are deduplicated against live workspace strategy state before each ordered creation.
    const created = await createImportStrategicPillar(
      {
        name: pillar.name,
        description: pillar.description,
        orderIndex: pillar.orderIndex,
      },
      ctx,
    );
    strategyMap.pillars.push(created);
    pillarMappings.set(pillar.sourceId, created);
    createdStrategicPillars += 1;
  }

  const objectiveMappings = new Map<string, { id: string; teamId: string }>();
  const sourceObjectivesByDestinationName = new Map<
    string,
    typeof selectedObjectives
  >();
  for (const objective of selectedObjectives) {
    const teamId = getTargetTeamId(objective.teamSourceId);
    if (!teamId) continue;
    const destinationNameKey = `${teamId}\u0000${normalizeImportMatch(
      objective.name,
    )}`;
    const matches =
      sourceObjectivesByDestinationName.get(destinationNameKey) ?? [];
    matches.push(objective);
    sourceObjectivesByDestinationName.set(destinationNameKey, matches);
  }
  const duplicateSourceObjectiveIds = new Set<string>();
  for (const matches of sourceObjectivesByDestinationName.values()) {
    if (matches.length < 2) continue;
    destinationConflicts += matches.length;
    for (const objective of matches) {
      duplicateSourceObjectiveIds.add(objective.sourceId);
    }
  }

  for (const objective of selectedObjectives) {
    const teamId = getTargetTeamId(objective.teamSourceId);
    if (!teamId) continue;
    if (duplicateSourceObjectiveIds.has(objective.sourceId)) continue;
    const cachedObjective = sourceObjectiveCache.get(objective.sourceId);
    if (cachedObjective?.teamId === teamId) {
      objectiveMappings.set(objective.sourceId, cachedObjective);
      continue;
    }
    const context = getTeamContext(teamId);
    const exactNameMatches = context.objectives.filter(
      (candidate) =>
        candidate.teamId === teamId &&
        normalizeImportMatch(candidate.name) ===
          normalizeImportMatch(objective.name),
    );
    if (exactNameMatches.length > 0) {
      const compatibleMatch = resolveImportEntityNameMatch(
        objective.name,
        exactNameMatches.filter(
          (candidate) => candidate.isPrivate === objective.isPrivate,
        ),
      );
      if (compatibleMatch.kind !== "unique") {
        destinationConflicts += 1;
        continue;
      }
      const mapping = { id: compatibleMatch.entity.id, teamId };
      objectiveMappings.set(objective.sourceId, mapping);
      sourceObjectiveCache.set(objective.sourceId, mapping);
      continue;
    }
    const status = resolveImportStatus(
      objective.status,
      objectiveStatuses,
      objective.statusCategory,
    );
    if (!status) {
      throw new Error(
        `No objective workflow status is available for ${objective.name}`,
      );
    }
    const leadPerson = objective.leadPersonSourceId
      ? peopleBySourceId.get(objective.leadPersonSourceId)
      : undefined;
    const lead = resolveReviewedPerson(
      leadPerson,
      teamId,
      objective.leadPersonSourceId ?? `objective:${objective.sourceId}`,
    );
    const hasValidDates = isValidImportDateRange(
      objective.startDate,
      objective.endDate,
    );
    const objectiveColor = toOptionalImportEntityColor(objective.color);
    // eslint-disable-next-line no-await-in-loop -- Objectives are created in source order so dependent key results and stories receive stable IDs.
    const created = await createImportObjective(
      {
        name: objective.name,
        ...(objective.description
          ? { description: objective.description }
          : {}),
        ...(objective.shortSummary
          ? { shortSummary: objective.shortSummary }
          : {}),
        ...(objectiveColor ? { color: objectiveColor } : {}),
        ...(lead ? { leadUser: lead.id } : {}),
        teamId,
        ...(hasValidDates && objective.startDate
          ? { startDate: objective.startDate }
          : {}),
        ...(hasValidDates && objective.endDate
          ? { endDate: objective.endDate }
          : {}),
        isPrivate: objective.isPrivate,
        statusId: status.id,
        priority: objective.priority,
      },
      ctx,
    );
    context.objectives.push(created.objective);
    const mapping = {
      id: created.objective.id,
      teamId,
    };
    objectiveMappings.set(objective.sourceId, mapping);
    sourceObjectiveCache.set(objective.sourceId, mapping);
    createdObjectives += 1;
  }

  for (const objective of selectedObjectives) {
    if (
      !objective.pillarSourceId ||
      !selectedStrategicPillarSourceIds.has(objective.pillarSourceId)
    ) {
      continue;
    }
    const objectiveMapping = objectiveMappings.get(objective.sourceId);
    if (!objectiveMapping) continue;
    const destinationPillar = pillarMappings.get(objective.pillarSourceId);
    if (!destinationPillar) {
      destinationConflicts += 1;
      continue;
    }
    const currentPillars = strategyMap.pillars.filter((pillar) =>
      pillar.objectiveIds.includes(objectiveMapping.id),
    );
    if (currentPillars.some((pillar) => pillar.id === destinationPillar.id)) {
      continue;
    }
    if (currentPillars.length > 0) {
      destinationConflicts += 1;
      continue;
    }
    // eslint-disable-next-line no-await-in-loop -- Objective alignment is checked against current strategy state before each mutation.
    await alignImportObjectiveToPillar(
      objectiveMapping.id,
      destinationPillar.id,
      ctx,
    );
    destinationPillar.objectiveIds.push(objectiveMapping.id);
    alignedObjectives += 1;
  }
  onProgress(42);

  const keyResultMappings = new Map<
    string,
    { id: string; objectiveId: string; teamId: string }
  >();
  for (const objective of selectedObjectives) {
    const objectiveMapping = objectiveMappings.get(objective.sourceId);
    if (!objectiveMapping) continue;
    const sourceKeyResults = importableKeyResults.filter(
      (keyResult) => keyResult.objectiveSourceId === objective.sourceId,
    );
    if (sourceKeyResults.length === 0) continue;
    // eslint-disable-next-line no-await-in-loop -- Each objective has an independent key-result collection that must be checked before creation.
    const existingKeyResults = await getImportObjectiveKeyResults(
      objectiveMapping.id,
      ctx,
    );
    const sourceKeyResultsByName = new Map<string, typeof sourceKeyResults>();
    for (const keyResult of sourceKeyResults) {
      const normalizedName = normalizeImportMatch(keyResult.name);
      const matches = sourceKeyResultsByName.get(normalizedName) ?? [];
      matches.push(keyResult);
      sourceKeyResultsByName.set(normalizedName, matches);
    }
    const duplicateSourceKeyResultIds = new Set<string>();
    for (const matches of sourceKeyResultsByName.values()) {
      if (matches.length < 2) continue;
      destinationConflicts += matches.length;
      for (const keyResult of matches) {
        duplicateSourceKeyResultIds.add(keyResult.sourceId);
      }
    }

    const newKeyResults: typeof sourceKeyResults = [];
    for (const keyResult of sourceKeyResults) {
      if (duplicateSourceKeyResultIds.has(keyResult.sourceId)) continue;
      const existing = resolveImportEntityNameMatch(
        keyResult.name,
        existingKeyResults,
      );
      if (existing.kind === "ambiguous") {
        destinationConflicts += 1;
        continue;
      }
      if (existing.kind === "unique") {
        keyResultMappings.set(keyResult.sourceId, {
          id: existing.entity.id,
          objectiveId: objectiveMapping.id,
          teamId: objectiveMapping.teamId,
        });
        continue;
      }
      newKeyResults.push(keyResult);
    }

    for (const batch of chunkImportItems(newKeyResults, 20)) {
      const payload = batch.map((keyResult) => {
        const leadPerson = keyResult.leadPersonSourceId
          ? peopleBySourceId.get(keyResult.leadPersonSourceId)
          : undefined;
        const lead = resolveReviewedPerson(
          leadPerson,
          objectiveMapping.teamId,
          keyResult.leadPersonSourceId ?? `key-result:${keyResult.sourceId}`,
        );
        const contributors = keyResult.contributorPersonSourceIds.flatMap(
          (personSourceId) => {
            const member = resolveReviewedPerson(
              peopleBySourceId.get(personSourceId),
              objectiveMapping.teamId,
              personSourceId,
            );
            return member ? [member.id] : [];
          },
        );
        return {
          name: keyResult.name,
          measurementType: keyResult.measurementType!,
          startValue: keyResult.startValue!,
          currentValue: keyResult.currentValue!,
          targetValue: keyResult.targetValue!,
          ...(lead ? { lead: lead.id } : {}),
          contributors: [...new Set(contributors)],
          startDate: keyResult.startDate!,
          endDate: keyResult.endDate!,
        };
      });
      // eslint-disable-next-line no-await-in-loop -- Server key-result batches are capped at 20 and must finish before mapping their IDs.
      const created = await createImportKeyResults(
        objectiveMapping.id,
        payload,
        ctx,
      );
      if (created.length !== batch.length) {
        throw new Error(
          `The destination returned ${created.length} of ${batch.length} created key results`,
        );
      }
      createdKeyResults += created.length;
      for (let index = 0; index < batch.length; index += 1) {
        const keyResult = batch[index];
        const match = created[index];
        keyResultMappings.set(keyResult.sourceId, {
          id: match.id,
          objectiveId: objectiveMapping.id,
          teamId: objectiveMapping.teamId,
        });
      }
    }
  }
  onProgress(55);

  const labelMappings = new Map<string, string>();
  const labelsBySourceId = new Map(
    draft.labels.map((label) => [label.sourceId, label]),
  );
  const labelTargets = new Map<
    string,
    { label: ImportDraft["labels"][number]; teamId: string | null }
  >();
  const getLabelMappingKey = (
    label: ImportDraft["labels"][number],
    teamId: string,
  ) =>
    label.teamSourceId === null
      ? `global\u0000${label.sourceId}`
      : `${teamId}\u0000${label.sourceId}`;
  const sourceLabelIdsByDestinationIdentity = new Map<string, Set<string>>();
  for (const label of draft.labels) {
    const teamId =
      label.teamSourceId === null ? null : getTargetTeamId(label.teamSourceId);
    if (label.teamSourceId !== null && !teamId) continue;
    const destinationIdentity = JSON.stringify([
      teamId === null ? "workspace" : "team",
      teamId,
      normalizeImportMatch(label.name),
    ]);
    const sourceIds =
      sourceLabelIdsByDestinationIdentity.get(destinationIdentity) ?? new Set();
    sourceIds.add(label.sourceId);
    sourceLabelIdsByDestinationIdentity.set(destinationIdentity, sourceIds);
  }
  const duplicateSourceLabelIds = new Set<string>();
  for (const sourceIds of sourceLabelIdsByDestinationIdentity.values()) {
    if (sourceIds.size < 2) continue;
    destinationConflicts += sourceIds.size;
    for (const sourceId of sourceIds) {
      duplicateSourceLabelIds.add(sourceId);
    }
  }
  const addLabelTarget = (
    label: ImportDraft["labels"][number],
    teamId: string | null,
  ) => {
    if (duplicateSourceLabelIds.has(label.sourceId)) return;
    const mappingKey =
      teamId === null
        ? `global\u0000${label.sourceId}`
        : `${teamId}\u0000${label.sourceId}`;
    labelTargets.set(mappingKey, { label, teamId });
  };
  for (const label of draft.labels) {
    if (label.teamSourceId === null) {
      addLabelTarget(label, null);
      continue;
    }
    const teamId = getTargetTeamId(label.teamSourceId);
    if (teamId) addLabelTarget(label, teamId);
  }
  for (const task of selectedTasks) {
    const teamId = getTargetTeamId(task.teamSourceId);
    if (!teamId) continue;
    for (const labelSourceId of task.labelSourceIds) {
      const label = labelsBySourceId.get(labelSourceId);
      if (!label) continue;
      if (label.teamSourceId === null) {
        addLabelTarget(label, null);
        continue;
      }
      const labelTeamId = getTargetTeamId(label.teamSourceId);
      if (!labelTeamId || labelTeamId !== teamId) {
        destinationConflicts += 1;
        continue;
      }
      addLabelTarget(label, teamId);
    }
  }
  const globalLabels = workspaceLabels.filter((label) => label.teamId === null);
  for (const [mappingKey, { label, teamId }] of labelTargets) {
    const candidates =
      teamId === null
        ? globalLabels
        : getTeamContext(teamId).labels.filter(
            (candidate) => candidate.teamId === teamId,
          );
    const existing = resolveImportEntityNameMatch(label.name, candidates);
    if (existing.kind === "ambiguous") {
      destinationConflicts += 1;
      continue;
    }
    if (existing.kind === "unique") {
      labelMappings.set(mappingKey, existing.entity.id);
      continue;
    }
    // eslint-disable-next-line no-await-in-loop -- Labels are deduplicated against the live team scope before each mutation.
    const created = await createImportLabel(
      {
        name: label.name,
        color: toImportEntityColor(label.color),
        ...(teamId ? { teamId } : {}),
      },
      ctx,
    );
    if (teamId) {
      getTeamContext(teamId).labels.push(created);
    } else {
      globalLabels.push(created);
      for (const context of teamContexts.values()) {
        context.labels.push(created);
      }
    }
    labelMappings.set(mappingKey, created.id);
    createdLabels += 1;
  }
  onProgress(62);

  const sprintMappings = new Map<string, { id: string; teamId: string }>();
  const sourceSprintIdsByDestinationIdentity = new Map<string, Set<string>>();
  for (const sprint of importableSprints) {
    const teamId = getTargetTeamId(sprint.teamSourceId);
    if (!teamId) continue;
    const objective = sprint.objectiveSourceId
      ? objectiveMappings.get(sprint.objectiveSourceId)
      : undefined;
    const canLinkObjective = objective?.teamId === teamId;
    if (sprint.objectiveSourceId && !canLinkObjective) continue;
    const destinationIdentity = JSON.stringify([
      teamId,
      normalizeImportMatch(sprint.name),
      sprint.startDate,
      sprint.endDate,
      canLinkObjective ? objective.id : null,
    ]);
    const sourceIds =
      sourceSprintIdsByDestinationIdentity.get(destinationIdentity) ??
      new Set();
    sourceIds.add(sprint.sourceId);
    sourceSprintIdsByDestinationIdentity.set(destinationIdentity, sourceIds);
  }
  const duplicateSourceSprintIds = new Set<string>();
  for (const sourceIds of sourceSprintIdsByDestinationIdentity.values()) {
    if (sourceIds.size < 2) continue;
    destinationConflicts += sourceIds.size;
    for (const sourceId of sourceIds) {
      duplicateSourceSprintIds.add(sourceId);
    }
  }
  for (const sprint of importableSprints) {
    const teamId = getTargetTeamId(sprint.teamSourceId);
    if (!teamId) continue;
    if (duplicateSourceSprintIds.has(sprint.sourceId)) continue;
    const context = getTeamContext(teamId);
    const objective = sprint.objectiveSourceId
      ? objectiveMappings.get(sprint.objectiveSourceId)
      : undefined;
    const canLinkObjective = objective?.teamId === teamId;
    if (sprint.objectiveSourceId && !canLinkObjective) {
      destinationConflicts += 1;
      continue;
    }
    const expectedObjectiveId = canLinkObjective ? objective.id : null;
    const scheduleMatches = context.sprints.filter(
      (candidate) =>
        candidate.teamId === teamId &&
        candidate.startDate.slice(0, 10) === sprint.startDate &&
        candidate.endDate.slice(0, 10) === sprint.endDate,
    );
    const existing = resolveImportEntityNameMatch(
      sprint.name,
      scheduleMatches.filter(
        (candidate) => (candidate.objectiveId || null) === expectedObjectiveId,
      ),
    );
    if (existing.kind === "ambiguous") {
      destinationConflicts += 1;
      continue;
    }
    if (existing.kind === "unique") {
      sprintMappings.set(sprint.sourceId, {
        id: existing.entity.id,
        teamId,
      });
      continue;
    }
    const incompatibleObjectiveMatch = resolveImportEntityNameMatch(
      sprint.name,
      scheduleMatches.filter(
        (candidate) => (candidate.objectiveId || null) !== expectedObjectiveId,
      ),
    );
    if (incompatibleObjectiveMatch.kind !== "none") {
      destinationConflicts += 1;
      continue;
    }
    // eslint-disable-next-line no-await-in-loop -- Sprints require resolved objective and team IDs before creation.
    const created = await createImportSprint(
      {
        name: sprint.name,
        ...(sprint.goal ? { goal: sprint.goal } : {}),
        ...(canLinkObjective ? { objectiveId: objective.id } : {}),
        teamId,
        startDate: sprint.startDate!,
        endDate: sprint.endDate!,
      },
      ctx,
    );
    context.sprints.push(created);
    sprintMappings.set(sprint.sourceId, { id: created.id, teamId });
    createdSprints += 1;
  }
  onProgress(70);

  const allResults: ImportStoryResult[] = [];
  const normalizeSourceId = (sourceId: string) => {
    const value = sourceId.trim();
    const uppercaseValue = value.toUpperCase();
    return draft.sourceType === "jira_csv" &&
      JIRA_ISSUE_KEY_PATTERN.test(uppercaseValue)
      ? uppercaseValue
      : value;
  };
  const sourceIdCounts = selectedTasks.reduce<Map<string, number>>(
    (counts, task) => {
      const sourceId = normalizeSourceId(task.sourceId);
      counts.set(sourceId, (counts.get(sourceId) ?? 0) + 1);
      return counts;
    },
    new Map(),
  );
  const getTaskProvider = (task: ImportTask) => {
    const sourceId = normalizeSourceId(task.sourceId);
    return draft.sourceType === "jira_csv" &&
      JIRA_ISSUE_KEY_PATTERN.test(sourceId) &&
      sourceIdCounts.get(sourceId) === 1
      ? ("jira_csv" as const)
      : ("file" as const);
  };
  const preparedTasks = await Promise.all(
    draft.tasks
      .map((task, taskIndex) => ({ task, taskIndex }))
      .filter(({ taskIndex }) => selectedTaskIndexes.has(taskIndex))
      .map(async ({ task, taskIndex }) => {
        const sourceId = normalizeSourceId(task.sourceId);
        const provider = getTaskProvider(task);
        const sourceKey = await getBoundedImportSourceKey(
          provider === "jira_csv" || sourceIdCounts.get(sourceId) === 1
            ? sourceId
            : `${sourceId}#row-${taskIndex + 1}`,
        );
        return {
          provider,
          sourceId,
          sourceKey,
          parentSourceId: task.parentSourceId
            ? normalizeSourceId(task.parentSourceId)
            : null,
          task,
          teamId: getTargetTeamId(task.teamSourceId),
        };
      }),
  );
  const resolvedPreparedTasks = preparedTasks.flatMap((item) =>
    item.teamId ? [{ ...item, teamId: item.teamId }] : [],
  );
  const uniqueTasksBySourceId = new Map(
    resolvedPreparedTasks
      .filter(({ sourceId }) => sourceIdCounts.get(sourceId) === 1)
      .map((item) => [item.sourceId, item]),
  );
  const crossTeamParentSourceKeys = new Set(
    resolvedPreparedTasks.flatMap((item) => {
      if (!item.parentSourceId) return [];
      const parent = uniqueTasksBySourceId.get(item.parentSourceId);
      return parent && parent.teamId !== item.teamId ? [item.sourceKey] : [];
    }),
  );
  destinationConflicts += crossTeamParentSourceKeys.size;
  const unresolvedStorySprintSourceKeys = new Set(
    resolvedPreparedTasks.flatMap((item) => {
      if (!item.task.sprintSourceId) return [];
      const sprint = sprintMappings.get(item.task.sprintSourceId);
      return !sprint || sprint.teamId !== item.teamId ? [item.sourceKey] : [];
    }),
  );
  destinationConflicts += unresolvedStorySprintSourceKeys.size;
  const importedStoryIds = new Map<string, string>();
  const importedStoryIdsBySourceKey = new Map<string, string>();
  const failedStorySourceIds = new Set<string>();
  const unresolvedTeamTasks = preparedTasks.filter((item) => !item.teamId);
  for (const item of unresolvedTeamTasks) {
    allResults.push({
      sourceKey: item.sourceKey,
      storyId: null,
      created: false,
      error: {
        code: "destination_team_conflict",
        message: "The destination team match is ambiguous.",
      },
    });
    failedStorySourceIds.add(item.sourceId);
  }
  let processedStories = unresolvedTeamTasks.length;
  let pendingTasks = [...resolvedPreparedTasks];

  while (pendingTasks.length > 0) {
    const blocked = pendingTasks.filter(({ parentSourceId, teamId }) => {
      if (!parentSourceId) return false;
      const parent = uniqueTasksBySourceId.get(parentSourceId);
      return (
        parent?.teamId === teamId && failedStorySourceIds.has(parentSourceId)
      );
    });
    const blockedKeys = new Set(blocked.map(({ sourceKey }) => sourceKey));
    for (const item of blocked) {
      allResults.push({
        sourceKey: item.sourceKey,
        storyId: null,
        created: false,
        error: {
          code: "parent_import_failed",
          message: "The parent work item could not be imported.",
        },
      });
      failedStorySourceIds.add(item.sourceId);
    }
    pendingTasks = pendingTasks.filter(
      ({ sourceKey }) => !blockedKeys.has(sourceKey),
    );
    processedStories += blocked.length;

    const ready = pendingTasks.filter(({ parentSourceId, teamId }) => {
      if (!parentSourceId) return true;
      const parent = uniqueTasksBySourceId.get(parentSourceId);
      if (!parent || parent.teamId !== teamId) return true;
      return importedStoryIds.has(parentSourceId);
    });
    if (ready.length === 0 && pendingTasks.length > 0) {
      for (const item of pendingTasks) {
        allResults.push({
          sourceKey: item.sourceKey,
          storyId: null,
          created: false,
          error: {
            code: "parent_cycle",
            message: "The source contains a circular parent relationship.",
          },
        });
      }
      processedStories += pendingTasks.length;
      pendingTasks = [];
      break;
    }

    const readyKeys = new Set(ready.map(({ sourceKey }) => sourceKey));
    const preparedRequestItems = ready.map(
      ({ parentSourceId, provider, sourceId, sourceKey, task, teamId }) => {
        const context = getTeamContext(teamId);
        const identity = getTaskImportPerson(task, peopleBySourceId);
        const assignee = resolveReviewedPerson(
          identity,
          teamId,
          task.assigneePersonSourceId,
        );
        const collaboratorIds = [
          ...new Set(
            task.collaboratorPersonSourceIds.flatMap((personSourceId) => {
              const collaborator = resolveReviewedPerson(
                peopleBySourceId.get(personSourceId),
                teamId,
                personSourceId,
              );
              return collaborator && collaborator.id !== assignee?.id
                ? [collaborator.id]
                : [];
            }),
          ),
        ];
        const objective = task.objectiveSourceId
          ? objectiveMappings.get(task.objectiveSourceId)
          : undefined;
        const keyResult = task.keyResultSourceId
          ? keyResultMappings.get(task.keyResultSourceId)
          : undefined;
        const sprint = task.sprintSourceId
          ? sprintMappings.get(task.sprintSourceId)
          : undefined;
        const labelIds = task.labelSourceIds.flatMap((labelSourceId) => {
          const label = labelsBySourceId.get(labelSourceId);
          if (!label) return [];
          const labelId = labelMappings.get(getLabelMappingKey(label, teamId));
          return labelId ? [labelId] : [];
        });
        const parent = parentSourceId
          ? uniqueTasksBySourceId.get(parentSourceId)
          : undefined;
        const parentId =
          parentSourceId && parent?.teamId === teamId
            ? importedStoryIds.get(parentSourceId)
            : undefined;
        const linkedObjective = keyResult
          ? {
              id: keyResult.objectiveId,
              teamId: keyResult.teamId,
            }
          : objective;
        const taskWithSafeDates = isValidImportDateRange(
          task.startDate,
          task.endDate,
        )
          ? task
          : { ...task, startDate: null, endDate: null };
        return {
          collaboratorIds,
          provider,
          sourceId,
          sourceKey,
          requestItem: {
            sourceKey,
            story: toImportStoryPayload({
              allowAutomaticAssigneeResolution: false,
              ...(assignee ? { assigneeId: assignee.id } : {}),
              ...(linkedObjective ? { objectiveId: linkedObjective.id } : {}),
              ...(keyResult ? { keyResultId: keyResult.id } : {}),
              ...(sprint?.teamId === teamId ? { sprintId: sprint.id } : {}),
              ...(parentId ? { parentId } : {}),
              labelIds,
              members: context.members,
              statuses: context.statuses,
              task: taskWithSafeDates,
              teamId,
            }),
          },
        };
      },
    );
    const importRequests = (["jira_csv", "file"] as const).flatMap(
      (provider) => {
        const items = preparedRequestItems.flatMap((item) =>
          item.provider === provider ? [item.requestItem] : [],
        );
        return items.length > 0
          ? buildImportStoryRequests({
              items,
              provider,
              sourceDigest: draft.fileHash,
              ...(draft.sourceNamespace
                ? { sourceNamespace: draft.sourceNamespace }
                : {}),
            })
          : [];
      },
    );
    const readyBySourceKey = new Map(
      preparedRequestItems.map((item) => [item.sourceKey, item]),
    );
    for (const request of importRequests) {
      // eslint-disable-next-line no-await-in-loop -- Sequential batches cap load and make progress truthful.
      const response = await importStoriesBatch(request, ctx);
      if (response.error?.message || !response.data) {
        throw new Error(
          response.error?.message || "A batch could not be imported",
        );
      }
      allResults.push(...response.data.items);
      const collaboratorUpdates: Promise<number>[] = [];
      for (const result of response.data.items) {
        const item = readyBySourceKey.get(result.sourceKey);
        if (!item) continue;
        if (result.storyId && !result.error) {
          importedStoryIdsBySourceKey.set(result.sourceKey, result.storyId);
          if (sourceIdCounts.get(item.sourceId) === 1) {
            importedStoryIds.set(item.sourceId, result.storyId);
          }
          if (item.collaboratorIds.length > 0) {
            collaboratorUpdates.push(
              mergeImportStoryCollaborators(
                result.storyId,
                item.collaboratorIds,
                ctx,
              ),
            );
          }
        } else {
          failedStorySourceIds.add(item.sourceId);
        }
      }
      // eslint-disable-next-line no-await-in-loop -- Each story batch must finish collaborator reconciliation before advancing parent-dependent work.
      const appliedCounts = await Promise.all(collaboratorUpdates);
      appliedCollaborators += appliedCounts.reduce(
        (total, count) => total + count,
        0,
      );
      processedStories += request.items.length;
      onProgress(
        preparedTasks.length
          ? 70 + Math.round((processedStories / preparedTasks.length) * 25)
          : 95,
      );
    }
    pendingTasks = pendingTasks.filter(
      ({ sourceKey }) => !readyKeys.has(sourceKey),
    );
  }
  onProgress(95);

  for (const item of preparedTasks) {
    const links = normalizeImportTaskLinks(item.task.links);
    if (links.length === 0) continue;
    const storyId = importedStoryIdsBySourceKey.get(item.sourceKey);
    if (!storyId) {
      unresolvedLinks += links.length;
      continue;
    }

    let existingUrls: Set<string>;
    try {
      // eslint-disable-next-line no-await-in-loop -- Each story must be inspected before link writes so retries cannot duplicate existing links.
      const existingLinks = await getImportStoryLinks(storyId, ctx);
      existingUrls = new Set(
        normalizeImportTaskLinks(
          existingLinks.map((link) => ({
            title: link.title,
            url: link.url,
          })),
        ).map((link) => link.url),
      );
    } catch {
      unresolvedLinks += links.length;
      continue;
    }

    for (const link of links) {
      if (existingUrls.has(link.url)) continue;
      try {
        // eslint-disable-next-line no-await-in-loop -- Link writes are intentionally sequential and exact-URL deduplicated for safe partial retry.
        await createImportStoryLink(
          {
            storyId,
            url: link.url,
            ...(link.title ? { title: link.title } : {}),
          },
          ctx,
        );
        existingUrls.add(link.url);
        createdLinks += 1;
      } catch {
        unresolvedLinks += 1;
      }
    }
  }

  const plannedAssociations = new Map<
    string,
    { fromStoryId: string; toStoryId: string; type: StoryAssociationType }
  >();
  const seenSourceAssociationKeys = new Set<string>();
  for (const item of preparedTasks) {
    for (const association of item.task.associations) {
      const targetSourceId = normalizeSourceId(association.targetSourceId);
      const sourceVertexId =
        sourceIdCounts.get(item.sourceId) === 1
          ? item.sourceId
          : item.sourceKey;
      const sourceAssociation = getCanonicalImportAssociation(
        sourceVertexId,
        targetSourceId,
        association.type,
      );
      const sourceAssociationKey = getImportAssociationKey(sourceAssociation);
      if (seenSourceAssociationKeys.has(sourceAssociationKey)) continue;
      seenSourceAssociationKeys.add(sourceAssociationKey);

      const sourceTask = uniqueTasksBySourceId.get(item.sourceId);
      const targetTask = uniqueTasksBySourceId.get(targetSourceId);
      const sourceStoryId = importedStoryIds.get(item.sourceId);
      const targetStoryId = importedStoryIds.get(targetSourceId);
      if (
        sourceIdCounts.get(item.sourceId) !== 1 ||
        sourceIdCounts.get(targetSourceId) !== 1 ||
        !sourceTask ||
        !targetTask ||
        !sourceStoryId ||
        !targetStoryId ||
        sourceStoryId === targetStoryId
      ) {
        unresolvedAssociations += 1;
        continue;
      }
      if (sourceTask.teamId !== targetTask.teamId) {
        unresolvedAssociations += 1;
        destinationConflicts += 1;
        continue;
      }

      const destinationAssociation = getCanonicalImportAssociation(
        sourceStoryId,
        targetStoryId,
        association.type,
      );
      plannedAssociations.set(getImportAssociationKey(destinationAssociation), {
        fromStoryId: destinationAssociation.fromId,
        toStoryId: destinationAssociation.toId,
        type: destinationAssociation.type,
      });
    }
  }

  const associationStoryIds = new Set<string>();
  for (const association of plannedAssociations.values()) {
    associationStoryIds.add(association.fromStoryId);
    associationStoryIds.add(association.toStoryId);
  }
  const existingAssociationGroups = await Promise.all(
    [...associationStoryIds].map((storyId) =>
      getImportStoryAssociations(storyId, ctx),
    ),
  );
  const existingAssociationKeys = new Set(
    existingAssociationGroups.flatMap((associations) =>
      associations.map((association) =>
        getImportAssociationKey({
          fromId: association.fromStoryId,
          toId: association.toStoryId,
          type: association.type,
        }),
      ),
    ),
  );
  for (const [associationKey, association] of plannedAssociations) {
    if (existingAssociationKeys.has(associationKey)) continue;
    // eslint-disable-next-line no-await-in-loop -- Association writes are ordered so a partial failure is safely discoverable and reusable on retry.
    await createImportStoryAssociation(
      association.fromStoryId,
      { toStoryId: association.toStoryId, type: association.type },
      ctx,
    );
    existingAssociationKeys.add(associationKey);
    createdAssociations += 1;
  }
  onProgress(100);

  const created = allResults.filter((item) => item.created).length;
  const failed = allResults.filter((item) => item.error !== null).length;
  const replayed = allResults.length - created - failed;
  return {
    created,
    failed,
    items: allResults,
    replayed,
    teamId: fallbackTeamId,
    createdTeams,
    createdStrategicPillars,
    createdObjectives,
    createdKeyResults,
    createdSprints,
    createdLabels,
    createdLinks,
    addedMemberships: addedMembershipCount,
    appliedCollaborators,
    createdAssociations,
    alignedObjectives,
    destinationConflicts,
    unresolvedAssociations,
    unresolvedLinks,
    unresolvedPeople: unresolvedPeople.size,
  };
};
