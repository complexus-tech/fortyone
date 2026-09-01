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
};

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

const mergeAnalyzedTask = (
  task: ImportTask,
  analyzedTask: ImportTask | undefined,
  options: MergeAnalyzedTaskGraphOptions,
): ImportTask => {
  if (!analyzedTask) return task;
  const statusCategory = options.authoritativeFields?.has("status")
    ? task.statusCategory
    : analyzedTask.statusCategory;
  const priority = options.authoritativeFields?.has("priority")
    ? task.priority
    : analyzedTask.priority;
  const assigneeName = options.authoritativeFields?.has("assigneeEmail")
    ? task.assigneeName
    : analyzedTask.assigneeName;
  const assigneePersonSourceId = options.authoritativeFields?.has(
    "assigneeEmail",
  )
    ? task.assigneePersonSourceId
    : analyzedTask.assigneePersonSourceId;
  const taskReferencesAreSafe = !options.authoritativeFields?.has("sourceId");
  return {
    ...task,
    statusCategory,
    priority,
    estimateValue: analyzedTask.estimateValue,
    estimatedDurationMinutes: analyzedTask.estimatedDurationMinutes,
    minimumFocusBlockMinutes: analyzedTask.minimumFocusBlockMinutes,
    assigneeName,
    assigneePersonSourceId,
    collaboratorPersonSourceIds: analyzedTask.collaboratorPersonSourceIds,
    teamSourceId: analyzedTask.teamSourceId,
    parentSourceId: taskReferencesAreSafe ? analyzedTask.parentSourceId : null,
    objectiveSourceId: analyzedTask.objectiveSourceId,
    keyResultSourceId: analyzedTask.keyResultSourceId,
    sprintSourceId: analyzedTask.sprintSourceId,
    labelSourceIds: analyzedTask.labelSourceIds,
    associations: taskReferencesAreSafe ? analyzedTask.associations : [],
    links: normalizeImportTaskLinks(analyzedTask.links),
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

    const normalizedTitle = normalizeImportTitle(task.title);
    const titleMatch =
      titleCounts.get(normalizedTitle) === 1
        ? analyzedByTitle.get(normalizedTitle)
        : undefined;
    return mergeAnalyzedTask(task, titleMatch ?? undefined, options);
  });
};
