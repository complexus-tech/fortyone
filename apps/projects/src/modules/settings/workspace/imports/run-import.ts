import type { ImportRunResult, RunImportInput } from "./import-run-model";
import { selectImportWork } from "./import-selection";
import { prepareImportTeamDestinations } from "./import-team-destinations";
import { loadImportDestinationContext } from "./import-destination-context";
import { prepareImportPeople } from "./import-people";
import { importObjectives } from "./import-objectives";
import { importLabels } from "./import-labels";
import { importSprints } from "./import-sprints";
import { importStories } from "./import-stories";
import { importRelationships } from "./import-relationships";

export type {
  ImportRunResult,
  RunImportInput,
  ImportStructureMode,
} from "./import-run-model";
export {
  getImportSourceTeamDestination,
  resolveImportSourceTeam,
} from "./import-team-model";
export {
  getCanonicalImportAssociation,
  getImportAssociationKey,
} from "./import-association-model";

// Each phase completes before its dependent entities are created. Destination
// reads and collaborator updates retain their own bounded parallel work.
export const runImport = async (
  input: RunImportInput,
): Promise<ImportRunResult> => {
  const selection = selectImportWork(input);
  const teams = await prepareImportTeamDestinations(input);
  const context = await loadImportDestinationContext(input, teams);
  const people = await prepareImportPeople(input, selection, teams, context);
  const objectives = await importObjectives(
    input,
    selection,
    teams,
    context,
    people,
  );
  const labels = await importLabels(input, selection, teams, context);
  const sprints = await importSprints(
    input,
    selection,
    teams,
    context,
    objectives,
  );
  const stories = await importStories(
    input,
    selection,
    teams,
    context,
    people,
    objectives,
    labels,
    sprints,
  );
  const relationships = await importRelationships(input, stories);

  const created = stories.allResults.filter((item) => item.created).length;
  const failed = stories.allResults.filter(
    (item) => item.error !== null,
  ).length;
  return {
    created,
    failed,
    items: stories.allResults,
    replayed: stories.allResults.length - created - failed,
    teamId: input.fallbackTeamId,
    createdTeams: teams.createdTeams,
    createdStrategicPillars: objectives.createdStrategicPillars,
    createdObjectives: objectives.createdObjectives,
    createdKeyResults: objectives.createdKeyResults,
    createdSprints: sprints.createdSprints,
    createdLabels: labels.createdLabels,
    createdLinks: relationships.createdLinks,
    addedMemberships: teams.addedMembershipCount + people.addedMembershipCount,
    appliedCollaborators: stories.appliedCollaborators,
    createdAssociations: relationships.createdAssociations,
    alignedObjectives: objectives.alignedObjectives,
    destinationConflicts:
      teams.destinationConflicts +
      objectives.destinationConflicts +
      labels.destinationConflicts +
      sprints.destinationConflicts +
      stories.destinationConflicts +
      relationships.destinationConflicts,
    unresolvedAssociations: relationships.unresolvedAssociations,
    unresolvedLinks: relationships.unresolvedLinks,
    unresolvedPeople: people.unresolvedPeople.size,
  };
};
