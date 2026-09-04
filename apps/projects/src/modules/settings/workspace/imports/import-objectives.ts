import type { RunImportInput } from "./import-run-model";
import type { ImportSelection } from "./import-selection";
import type { ImportTeamDestinations } from "./import-team-destinations";
import type { ImportDestinationContext } from "./import-destination-context";
import type { ImportPeople } from "./import-people";
import {
  getImportStrategyMap,
  createImportStrategicPillar,
  createImportObjective,
  alignImportObjectiveToPillar,
  getImportObjectiveKeyResults,
  createImportKeyResults,
} from "./api";
import {
  resolveImportEntityNameMatch,
  resolveImportStatus,
  isValidImportDateRange,
} from "./execution";
import { normalizeImportMatch } from "./import-entity-matching";

const chunkImportItems = <T>(items: T[], size: number) => {
  const chunks: T[][] = [];
  for (let index = 0; index < items.length; index += size) {
    chunks.push(items.slice(index, index + size));
  }
  return chunks;
};

const toOptionalImportEntityColor = (value: string | null) =>
  value && /^#[0-9A-Fa-f]{6}$/.test(value) ? value.toUpperCase() : undefined;

export const importObjectives = async (
  {
    ctx,
    onProgress,
    selectedStrategicPillarSourceIds,
    sourceObjectiveCache,
  }: Pick<
    RunImportInput,
    | "ctx"
    | "onProgress"
    | "selectedStrategicPillarSourceIds"
    | "sourceObjectiveCache"
  >,
  {
    selectedObjectives,
    selectedStrategicPillars,
    importableKeyResults,
  }: Pick<
    ImportSelection,
    "selectedObjectives" | "selectedStrategicPillars" | "importableKeyResults"
  >,
  { getTargetTeamId }: Pick<ImportTeamDestinations, "getTargetTeamId">,
  {
    getTeamContext,
    objectiveStatuses,
  }: Pick<ImportDestinationContext, "getTeamContext" | "objectiveStatuses">,
  {
    peopleBySourceId,
    resolveReviewedPerson,
  }: Pick<ImportPeople, "peopleBySourceId" | "resolveReviewedPerson">,
) => {
  let createdStrategicPillars = 0;
  let createdObjectives = 0;
  let createdKeyResults = 0;
  let alignedObjectives = 0;
  let destinationConflicts = 0;
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

  return {
    objectiveMappings,
    keyResultMappings,
    createdStrategicPillars,
    createdObjectives,
    createdKeyResults,
    alignedObjectives,
    destinationConflicts,
  };
};

export type ImportedObjectives = Awaited<ReturnType<typeof importObjectives>>;
