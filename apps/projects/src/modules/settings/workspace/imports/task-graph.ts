import {
  normalizeImportTaskLinks,
  type ImportMapping,
  type ImportTask,
} from "./schema";

export type MergeAnalyzedTaskGraphOptions = {
  /**
   * Keeps only the values rebuilt from user-edited column controls
   * authoritative. AI can still enrich untouched semantics and relationships.
   */
  authoritativeFields?: ReadonlySet<keyof ImportMapping>;
  /**
   * Treats analyzed tasks as sparse semantic patches. Required null and empty
   * schema values cannot erase exact values retained by a deterministic parser.
   */
  enrichmentOnly?: boolean;
};

const IMPORT_TASK_REFERENCE_LIMIT = 100;

const normalizeImportTitle = (value: string) =>
  value.normalize("NFKC").trim().toLocaleLowerCase().replace(/\s+/g, " ");

const countBy = <T>(items: T[], getKey: (item: T) => string) => {
  const counts = new Map<string, number>();
  for (const item of items) {
    const key = getKey(item);
    if (!key) continue;
    counts.set(key, (counts.get(key) ?? 0) + 1);
  }
  return counts;
};

const indexUniqueBy = <T>(items: T[], getKey: (item: T) => string) => {
  const index = new Map<string, T | null>();
  for (const item of items) {
    const key = getKey(item);
    if (!key) continue;
    index.set(key, index.has(key) ? null : item);
  }
  return index;
};

const mergeUniqueStrings = (first: string[], second: string[]) =>
  [...new Set([...first, ...second])].slice(0, IMPORT_TASK_REFERENCE_LIMIT);

const mergeTaskAssociations = (
  first: ImportTask["associations"],
  second: ImportTask["associations"],
) => {
  const seen = new Set<string>();
  return [...first, ...second]
    .filter((association) => {
      const key = `${association.type}\u0000${association.targetSourceId}`;
      if (seen.has(key)) return false;
      seen.add(key);
      return true;
    })
    .slice(0, IMPORT_TASK_REFERENCE_LIMIT);
};

const preferAnalyzedValue = <T>(
  sourceValue: T | null,
  analyzedValue: T | null,
  enrichmentOnly: boolean,
) => (enrichmentOnly ? analyzedValue ?? sourceValue : analyzedValue);

const mergeAnalyzedTask = (
  task: ImportTask,
  analyzedTask: ImportTask | undefined,
  options: MergeAnalyzedTaskGraphOptions,
): ImportTask => {
  if (!analyzedTask) return task;
  const enrichmentOnly = options.enrichmentOnly === true;
  let statusCategory = preferAnalyzedValue(
    task.statusCategory,
    analyzedTask.statusCategory,
    enrichmentOnly,
  );
  if (options.authoritativeFields?.has("status")) {
    statusCategory = task.statusCategory;
  }
  let priority = analyzedTask.priority;
  if (
    options.authoritativeFields?.has("priority") ||
    (enrichmentOnly && analyzedTask.priority === "No Priority")
  ) {
    priority = task.priority;
  }
  let assigneeName = preferAnalyzedValue(
    task.assigneeName,
    analyzedTask.assigneeName,
    enrichmentOnly,
  );
  let assigneePersonSourceId = preferAnalyzedValue(
    task.assigneePersonSourceId,
    analyzedTask.assigneePersonSourceId,
    enrichmentOnly,
  );
  if (options.authoritativeFields?.has("assigneeEmail")) {
    assigneeName = task.assigneeName;
    assigneePersonSourceId = task.assigneePersonSourceId;
  }
  const collaboratorPersonSourceIds = (
    enrichmentOnly
      ? mergeUniqueStrings(
          task.collaboratorPersonSourceIds,
          analyzedTask.collaboratorPersonSourceIds,
        )
      : analyzedTask.collaboratorPersonSourceIds
  ).filter((sourceId) => sourceId !== assigneePersonSourceId);
  const taskReferencesAreSafe = !options.authoritativeFields?.has("sourceId");
  let parentSourceId: string | null = null;
  let associations: ImportTask["associations"] = [];
  if (taskReferencesAreSafe) {
    parentSourceId = preferAnalyzedValue(
      task.parentSourceId,
      analyzedTask.parentSourceId,
      enrichmentOnly,
    );
    associations = enrichmentOnly
      ? mergeTaskAssociations(task.associations, analyzedTask.associations)
      : analyzedTask.associations;
  }
  return {
    ...task,
    statusCategory,
    priority,
    estimateValue: preferAnalyzedValue(
      task.estimateValue,
      analyzedTask.estimateValue,
      enrichmentOnly,
    ),
    estimatedDurationMinutes: preferAnalyzedValue(
      task.estimatedDurationMinutes,
      analyzedTask.estimatedDurationMinutes,
      enrichmentOnly,
    ),
    minimumFocusBlockMinutes: preferAnalyzedValue(
      task.minimumFocusBlockMinutes,
      analyzedTask.minimumFocusBlockMinutes,
      enrichmentOnly,
    ),
    assigneeName,
    assigneePersonSourceId,
    collaboratorPersonSourceIds,
    teamSourceId: preferAnalyzedValue(
      task.teamSourceId,
      analyzedTask.teamSourceId,
      enrichmentOnly,
    ),
    parentSourceId,
    objectiveSourceId: preferAnalyzedValue(
      task.objectiveSourceId,
      analyzedTask.objectiveSourceId,
      enrichmentOnly,
    ),
    keyResultSourceId: preferAnalyzedValue(
      task.keyResultSourceId,
      analyzedTask.keyResultSourceId,
      enrichmentOnly,
    ),
    sprintSourceId: preferAnalyzedValue(
      task.sprintSourceId,
      analyzedTask.sprintSourceId,
      enrichmentOnly,
    ),
    labelSourceIds: enrichmentOnly
      ? mergeUniqueStrings(task.labelSourceIds, analyzedTask.labelSourceIds)
      : analyzedTask.labelSourceIds,
    associations,
    links: normalizeImportTaskLinks(
      enrichmentOnly
        ? [...task.links, ...analyzedTask.links]
        : analyzedTask.links,
    ),
  };
};

export const mergeAnalyzedTaskGraph = (
  tasks: ImportTask[],
  analyzedTasks: ImportTask[],
  options: MergeAnalyzedTaskGraphOptions = {},
) => {
  const sourceIdCounts = countBy(tasks, (task) => task.sourceId);
  const titleCounts = countBy(tasks, (task) =>
    normalizeImportTitle(task.title),
  );
  const analyzedBySourceId = indexUniqueBy(
    analyzedTasks,
    (task) => task.sourceId,
  );
  const analyzedByTitle = indexUniqueBy(analyzedTasks, (task) =>
    normalizeImportTitle(task.title),
  );

  return tasks.map((task) => {
    const exactMatch =
      sourceIdCounts.get(task.sourceId) === 1
        ? analyzedBySourceId.get(task.sourceId)
        : undefined;
    if (exactMatch) return mergeAnalyzedTask(task, exactMatch, options);
    if (options.enrichmentOnly) return task;

    const normalizedTitle = normalizeImportTitle(task.title);
    const titleMatch =
      titleCounts.get(normalizedTitle) === 1
        ? analyzedByTitle.get(normalizedTitle)
        : undefined;
    return mergeAnalyzedTask(task, titleMatch ?? undefined, options);
  });
};
