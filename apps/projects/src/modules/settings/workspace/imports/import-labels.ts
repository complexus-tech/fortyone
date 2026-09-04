import type { ImportDraft } from "./schema";
import type { RunImportInput } from "./import-run-model";
import type { ImportSelection } from "./import-selection";
import type { ImportTeamDestinations } from "./import-team-destinations";
import type { ImportDestinationContext } from "./import-destination-context";
import { createImportLabel } from "./api";
import { resolveImportEntityNameMatch } from "./execution";
import { normalizeImportMatch } from "./import-entity-matching";

const toImportEntityColor = (value: string | null) =>
  value && /^#[0-9A-Fa-f]{6}$/.test(value) ? value.toUpperCase() : "#697386";

export const importLabels = async (
  {
    ctx,
    draft,
    onProgress,
  }: Pick<RunImportInput, "ctx" | "draft" | "onProgress">,
  { selectedTasks }: Pick<ImportSelection, "selectedTasks">,
  { getTargetTeamId }: Pick<ImportTeamDestinations, "getTargetTeamId">,
  {
    getTeamContext,
    teamContexts,
    workspaceLabels,
  }: Pick<
    ImportDestinationContext,
    "getTeamContext" | "teamContexts" | "workspaceLabels"
  >,
) => {
  let createdLabels = 0;
  let destinationConflicts = 0;
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

  return {
    labelMappings,
    labelsBySourceId,
    getLabelMappingKey,
    createdLabels,
    destinationConflicts,
  };
};

export type ImportedLabels = Awaited<ReturnType<typeof importLabels>>;
