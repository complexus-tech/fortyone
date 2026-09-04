import type { RunImportInput } from "./import-run-model";
import type { ImportTeamDestinations } from "./import-team-destinations";
import {
  getImportObjectiveStatuses,
  getImportWorkspaceMembers,
  getImportWorkspaceLabels,
  getImportTeamStatuses,
  getImportTeamMembers,
  getImportTeamObjectives,
  getImportTeamLabels,
  getImportTeamSprints,
} from "./api";

export const loadImportDestinationContext = async (
  { ctx, draft }: Pick<RunImportInput, "ctx" | "draft">,
  { targetTeamIds }: Pick<ImportTeamDestinations, "targetTeamIds">,
) => {
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
  return {
    getTeamContext,
    teamContexts,
    objectiveStatuses,
    workspaceMembers,
    workspaceLabels,
  };
};

export type ImportDestinationContext = Awaited<
  ReturnType<typeof loadImportDestinationContext>
>;
