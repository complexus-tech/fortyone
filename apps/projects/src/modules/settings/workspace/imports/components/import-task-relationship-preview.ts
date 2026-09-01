import type { ImportTask } from "../schema";

export type ImportTaskRelationshipLookups = {
  keyResults: ReadonlyMap<string, string>;
  labels: ReadonlyMap<string, string>;
  people: ReadonlyMap<string, string | null>;
  sprints: ReadonlyMap<string, string>;
  tasks: ReadonlyMap<string, string>;
};

export type ImportTaskRelationshipTerminology = {
  keyResult: string;
  sprint: string;
};

const getSourceLabel = (
  sourceId: string,
  labels: ReadonlyMap<string, string | null>,
) => labels.get(sourceId)?.trim() || sourceId;

const summarizeReferences = (
  label: string,
  sourceIds: readonly string[],
  labels: ReadonlyMap<string, string | null>,
  limit = 4,
) => {
  if (sourceIds.length === 0) return null;
  const values = sourceIds
    .slice(0, limit)
    .map((sourceId) => getSourceLabel(sourceId, labels));
  const remaining = sourceIds.length - values.length;
  return `${label}: ${values.join(", ")}${remaining > 0 ? ` +${remaining}` : ""}`;
};

const ASSOCIATION_LABELS = {
  blocked_by: "Blocked by",
  blocks: "Blocks",
  duplicate: "Duplicates",
  related: "Related to",
} as const;

export const getImportTaskRelationshipPreview = (
  task: ImportTask,
  lookups: ImportTaskRelationshipLookups,
  terminology: ImportTaskRelationshipTerminology,
) => {
  const relationships: string[] = [];
  if (task.parentSourceId) {
    relationships.push(
      `Parent: ${getSourceLabel(task.parentSourceId, lookups.tasks)}`,
    );
  }
  if (task.keyResultSourceId) {
    relationships.push(
      `${terminology.keyResult}: ${getSourceLabel(task.keyResultSourceId, lookups.keyResults)}`,
    );
  }
  if (task.sprintSourceId) {
    relationships.push(
      `${terminology.sprint}: ${getSourceLabel(task.sprintSourceId, lookups.sprints)}`,
    );
  }

  const labelSummary = summarizeReferences(
    "Labels",
    task.labelSourceIds,
    lookups.labels,
  );
  if (labelSummary) relationships.push(labelSummary);
  const collaboratorSummary = summarizeReferences(
    "Collaborators",
    task.collaboratorPersonSourceIds,
    lookups.people,
  );
  if (collaboratorSummary) relationships.push(collaboratorSummary);

  for (const link of task.links.slice(0, 3)) {
    relationships.push(`Link: ${link.title?.trim() || link.url}`);
  }
  if (task.links.length > 3) {
    relationships.push(`${task.links.length - 3} more links`);
  }

  for (const association of task.associations.slice(0, 6)) {
    relationships.push(
      `${ASSOCIATION_LABELS[association.type]}: ${getSourceLabel(
        association.targetSourceId,
        lookups.tasks,
      )}`,
    );
  }
  if (task.associations.length > 6) {
    relationships.push(`${task.associations.length - 6} more relationships`);
  }
  return [...new Set(relationships)];
};
