import type { RunImportInput } from "./import-run-model";
import type { ImportSelection } from "./import-selection";
import type { ImportTeamDestinations } from "./import-team-destinations";
import type { ImportDestinationContext } from "./import-destination-context";
import type { ImportedObjectives } from "./import-objectives";
import { createImportSprint } from "./api";
import { resolveImportEntityNameMatch } from "./execution";
import { normalizeImportMatch } from "./import-entity-matching";

export const importSprints = async (
  { ctx, onProgress }: Pick<RunImportInput, "ctx" | "onProgress">,
  { importableSprints }: Pick<ImportSelection, "importableSprints">,
  { getTargetTeamId }: Pick<ImportTeamDestinations, "getTargetTeamId">,
  { getTeamContext }: Pick<ImportDestinationContext, "getTeamContext">,
  { objectiveMappings }: Pick<ImportedObjectives, "objectiveMappings">,
) => {
  let createdSprints = 0;
  let destinationConflicts = 0;
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

  return { sprintMappings, createdSprints, destinationConflicts };
};

export type ImportedSprints = Awaited<ReturnType<typeof importSprints>>;
