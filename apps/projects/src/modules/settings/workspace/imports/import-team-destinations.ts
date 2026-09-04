import type { RunImportInput } from "./import-run-model";
import { addExistingImportMemberToTeam, createImportTeam } from "./api";
import { deriveImportTeamCode, deriveImportTeamColor } from "./execution";
import { normalizeImportMatch } from "./import-entity-matching";
import { resolveImportSourceTeam, toImportTeamName } from "./import-team-model";

export const prepareImportTeamDestinations = async ({
  actorUserId,
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
  sourceTeamCache,
  structureMode,
}: Pick<
  RunImportInput,
  | "actorUserId"
  | "ctx"
  | "draft"
  | "existingTeams"
  | "fallbackTeamCode"
  | "fallbackTeamCreated"
  | "fallbackTeamId"
  | "fallbackTeamIsPrivate"
  | "fallbackTeamIsNew"
  | "fallbackTeamName"
  | "joinedTeamIds"
  | "onProgress"
  | "sourceTeamCache"
  | "structureMode"
>) => {
  let createdTeams = fallbackTeamCreated ? 1 : 0;
  let destinationConflicts = 0;
  let addedMembershipCount = 0;
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
  return {
    getTargetTeamId,
    targetTeamIds,
    createdTeams,
    destinationConflicts,
    addedMembershipCount,
  };
};

export type ImportTeamDestinations = Awaited<
  ReturnType<typeof prepareImportTeamDestinations>
>;
