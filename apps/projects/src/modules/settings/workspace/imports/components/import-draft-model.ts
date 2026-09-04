import type { ImportAnalysis, ImportDraft, ImportMapping } from "../schema";
import { mapRowsToImportTasks, sanitizeAIImportMapping } from "../csv";
import { mergeAnalyzedTaskGraph } from "../task-graph";

export const DO_NOT_IMPORT_VALUE = "__do_not_import__";
const TRELLO_SOURCE_NAMESPACE_PREFIX = "trello:board:";
const normalizeImportColumnName = (value: string) =>
  value.trim().toLocaleLowerCase().replace(/[_-]+/g, " ").replace(/\s+/g, " ");

const isDeterministicTrelloDraft = (draft: ImportDraft | null) =>
  Boolean(
    draft?.sourceType === "json" &&
      (draft.sourceMetadata?.platform === "trello" ||
        draft.sourceNamespace?.startsWith(TRELLO_SOURCE_NAMESPACE_PREFIX)),
  );

export const getTrelloArchivedTaskSourceIds = (draft: ImportDraft | null) => {
  if (!isDeterministicTrelloDraft(draft) || !draft) return new Set<string>();
  if (draft.sourceMetadata?.platform === "trello") {
    return new Set(
      draft.sourceMetadata.archivedTaskSourceIds.flatMap((sourceId) => {
        const normalizedSourceId = sourceId.trim();
        return normalizedSourceId ? [normalizedSourceId] : [];
      }),
    );
  }
  const sourceIdColumn =
    draft.mapping?.sourceId ??
    draft.columns.find((column) => normalizeImportColumnName(column) === "id");
  const closedColumn = draft.columns.find(
    (column) => normalizeImportColumnName(column) === "closed",
  );
  if (!sourceIdColumn || !closedColumn) return new Set<string>();

  return new Set(
    draft.rows.flatMap((row) => {
      const sourceId = (row[sourceIdColumn] ?? "").trim();
      const closed =
        (row[closedColumn] ?? "").trim().toLocaleLowerCase() === "true";
      return sourceId && closed ? [sourceId] : [];
    }),
  );
};

export const getTaskIndexesBySourceId = (
  draft: ImportDraft | null,
  sourceIds: ReadonlySet<string>,
) =>
  new Set(
    draft?.tasks.flatMap((task, index) =>
      sourceIds.has(task.sourceId) ? [index] : [],
    ) ?? [],
  );

const mergeDeterministicImportEntities = <T extends { sourceId: string }>(
  deterministicEntities: readonly T[],
  analyzedEntities: readonly T[],
) => {
  const analyzedBySourceId = new Map(
    analyzedEntities.map((entity) => [entity.sourceId, entity]),
  );
  const merged = deterministicEntities.map((entity) => {
    const analyzedEntity = analyzedBySourceId.get(entity.sourceId);
    return analyzedEntity ? { ...analyzedEntity, ...entity } : entity;
  });
  const sourceIds = new Set(
    deterministicEntities.map((entity) => entity.sourceId),
  );
  for (const entity of analyzedEntities) {
    if (sourceIds.has(entity.sourceId)) continue;
    sourceIds.add(entity.sourceId);
    merged.push(entity);
  }
  return merged;
};

export const mergeCompletedImportAnalysis = (
  current: ImportDraft | null,
  completedAnalysis: ImportAnalysis,
  {
    fileHash,
    fileName,
    mappingEdited,
    mappingOverrideFields,
  }: {
    fileHash: string;
    fileName: string;
    mappingEdited: boolean;
    mappingOverrideFields: ReadonlySet<keyof ImportMapping>;
  },
): ImportDraft => {
  const usesDeterministicRowMapping =
    current?.sourceType === "csv" || current?.sourceType === "jira_csv";
  const preservesDeterministicTrelloGraph = isDeterministicTrelloDraft(current);
  const preservesDeterministicTaskSet =
    usesDeterministicRowMapping || preservesDeterministicTrelloGraph;
  const canMergeDeterministicAnalysis = Boolean(
    current &&
      preservesDeterministicTaskSet &&
      (current.rows.length > 0 || preservesDeterministicTrelloGraph),
  );
  if (current && canMergeDeterministicAnalysis) {
    let mapping = completedAnalysis.mapping;
    if (usesDeterministicRowMapping) {
      mapping =
        !mappingEdited && completedAnalysis.mapping
          ? sanitizeAIImportMapping(completedAnalysis.mapping, current.columns)
          : current.mapping;
    }
    const mappedTasks =
      usesDeterministicRowMapping && mapping
        ? mapRowsToImportTasks(current.rows, mapping)
        : current.tasks;
    return {
      ...current,
      teams: preservesDeterministicTrelloGraph
        ? mergeDeterministicImportEntities(
            current.teams,
            completedAnalysis.teams,
          )
        : completedAnalysis.teams,
      people: preservesDeterministicTrelloGraph
        ? mergeDeterministicImportEntities(
            current.people,
            completedAnalysis.people,
          )
        : completedAnalysis.people,
      labels: preservesDeterministicTrelloGraph
        ? mergeDeterministicImportEntities(
            current.labels,
            completedAnalysis.labels,
          )
        : completedAnalysis.labels,
      strategicPillars: completedAnalysis.strategicPillars,
      objectives: completedAnalysis.objectives,
      keyResults: completedAnalysis.keyResults,
      sprints: completedAnalysis.sprints,
      mapping,
      sourceNamespace:
        current.sourceNamespace ?? completedAnalysis.sourceNamespace,
      summary: preservesDeterministicTrelloGraph
        ? current.summary
        : completedAnalysis.summary,
      tasks: mergeAnalyzedTaskGraph(mappedTasks, completedAnalysis.tasks, {
        authoritativeFields: mappingOverrideFields,
        enrichmentOnly: preservesDeterministicTrelloGraph,
      }),
      warnings: preservesDeterministicTrelloGraph
        ? [
            ...new Set([...current.warnings, ...completedAnalysis.warnings]),
          ].slice(0, 50)
        : completedAnalysis.warnings,
    };
  }
  if (current) {
    return {
      ...current,
      ...completedAnalysis,
      sourceNamespace:
        current.sourceNamespace ?? completedAnalysis.sourceNamespace,
    };
  }
  return {
    ...completedAnalysis,
    columns: [],
    fileHash,
    fileName,
    rows: [],
  };
};
