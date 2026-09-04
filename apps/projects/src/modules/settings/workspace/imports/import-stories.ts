import type { WorkspaceCtx } from "@/lib/http";
import type { ImportTask } from "./schema";
import type { ImportStoryResult } from "./api";
import type { RunImportInput } from "./import-run-model";
import type { ImportSelection } from "./import-selection";
import type { ImportTeamDestinations } from "./import-team-destinations";
import type { ImportDestinationContext } from "./import-destination-context";
import type { ImportPeople } from "./import-people";
import type { ImportedObjectives } from "./import-objectives";
import type { ImportedLabels } from "./import-labels";
import type { ImportedSprints } from "./import-sprints";
import { JIRA_ISSUE_KEY_PATTERN } from "./schema";
import {
  getImportStoryCollaboratorIds,
  updateImportStoryCollaborators,
  buildImportStoryRequests,
  importStoriesBatch,
} from "./api";
import {
  getBoundedImportSourceKey,
  isValidImportDateRange,
  toImportStoryPayload,
} from "./execution";
import { getTaskImportPerson } from "./import-people";

const mergeImportStoryCollaborators = async (
  storyId: string,
  importedCollaboratorIds: string[],
  ctx: WorkspaceCtx,
) => {
  const existingCollaboratorIds = await getImportStoryCollaboratorIds(
    storyId,
    ctx,
  );
  const existingIds = new Set(existingCollaboratorIds);
  const addedIds = [...new Set(importedCollaboratorIds)].filter(
    (memberId) => !existingIds.has(memberId),
  );
  if (addedIds.length === 0) return 0;

  await updateImportStoryCollaborators(
    storyId,
    [...existingIds, ...addedIds],
    ctx,
  );
  return addedIds.length;
};

export const importStories = async (
  {
    ctx,
    draft,
    selectedTaskIndexes,
    onProgress,
  }: Pick<
    RunImportInput,
    "ctx" | "draft" | "selectedTaskIndexes" | "onProgress"
  >,
  { selectedTasks }: Pick<ImportSelection, "selectedTasks">,
  { getTargetTeamId }: Pick<ImportTeamDestinations, "getTargetTeamId">,
  { getTeamContext }: Pick<ImportDestinationContext, "getTeamContext">,
  {
    peopleBySourceId,
    resolveReviewedPerson,
  }: Pick<ImportPeople, "peopleBySourceId" | "resolveReviewedPerson">,
  {
    objectiveMappings,
    keyResultMappings,
  }: Pick<ImportedObjectives, "objectiveMappings" | "keyResultMappings">,
  {
    labelMappings,
    labelsBySourceId,
    getLabelMappingKey,
  }: Pick<
    ImportedLabels,
    "labelMappings" | "labelsBySourceId" | "getLabelMappingKey"
  >,
  { sprintMappings }: Pick<ImportedSprints, "sprintMappings">,
) => {
  let appliedCollaborators = 0;
  let destinationConflicts = 0;
  const allResults: ImportStoryResult[] = [];
  const normalizeSourceId = (sourceId: string) => {
    const value = sourceId.trim();
    const uppercaseValue = value.toUpperCase();
    return draft.sourceType === "jira_csv" &&
      JIRA_ISSUE_KEY_PATTERN.test(uppercaseValue)
      ? uppercaseValue
      : value;
  };
  const sourceIdCounts = selectedTasks.reduce<Map<string, number>>(
    (counts, task) => {
      const sourceId = normalizeSourceId(task.sourceId);
      counts.set(sourceId, (counts.get(sourceId) ?? 0) + 1);
      return counts;
    },
    new Map(),
  );
  const getTaskProvider = (task: ImportTask) => {
    const sourceId = normalizeSourceId(task.sourceId);
    return draft.sourceType === "jira_csv" &&
      JIRA_ISSUE_KEY_PATTERN.test(sourceId) &&
      sourceIdCounts.get(sourceId) === 1
      ? ("jira_csv" as const)
      : ("file" as const);
  };
  const preparedTasks = await Promise.all(
    draft.tasks
      .map((task, taskIndex) => ({ task, taskIndex }))
      .filter(({ taskIndex }) => selectedTaskIndexes.has(taskIndex))
      .map(async ({ task, taskIndex }) => {
        const sourceId = normalizeSourceId(task.sourceId);
        const provider = getTaskProvider(task);
        const sourceKey = await getBoundedImportSourceKey(
          provider === "jira_csv" || sourceIdCounts.get(sourceId) === 1
            ? sourceId
            : `${sourceId}#row-${taskIndex + 1}`,
        );
        return {
          provider,
          sourceId,
          sourceKey,
          parentSourceId: task.parentSourceId
            ? normalizeSourceId(task.parentSourceId)
            : null,
          task,
          teamId: getTargetTeamId(task.teamSourceId),
        };
      }),
  );
  const resolvedPreparedTasks = preparedTasks.flatMap((item) =>
    item.teamId ? [{ ...item, teamId: item.teamId }] : [],
  );
  const uniqueTasksBySourceId = new Map(
    resolvedPreparedTasks
      .filter(({ sourceId }) => sourceIdCounts.get(sourceId) === 1)
      .map((item) => [item.sourceId, item]),
  );
  const crossTeamParentSourceKeys = new Set(
    resolvedPreparedTasks.flatMap((item) => {
      if (!item.parentSourceId) return [];
      const parent = uniqueTasksBySourceId.get(item.parentSourceId);
      return parent && parent.teamId !== item.teamId ? [item.sourceKey] : [];
    }),
  );
  destinationConflicts += crossTeamParentSourceKeys.size;
  const unresolvedStorySprintSourceKeys = new Set(
    resolvedPreparedTasks.flatMap((item) => {
      if (!item.task.sprintSourceId) return [];
      const sprint = sprintMappings.get(item.task.sprintSourceId);
      return !sprint || sprint.teamId !== item.teamId ? [item.sourceKey] : [];
    }),
  );
  destinationConflicts += unresolvedStorySprintSourceKeys.size;
  const importedStoryIds = new Map<string, string>();
  const importedStoryIdsBySourceKey = new Map<string, string>();
  const failedStorySourceIds = new Set<string>();
  const unresolvedTeamTasks = preparedTasks.filter((item) => !item.teamId);
  for (const item of unresolvedTeamTasks) {
    allResults.push({
      sourceKey: item.sourceKey,
      storyId: null,
      created: false,
      error: {
        code: "destination_team_conflict",
        message: "The destination team match is ambiguous.",
      },
    });
    failedStorySourceIds.add(item.sourceId);
  }
  let processedStories = unresolvedTeamTasks.length;
  let pendingTasks = [...resolvedPreparedTasks];

  while (pendingTasks.length > 0) {
    const blocked = pendingTasks.filter(({ parentSourceId, teamId }) => {
      if (!parentSourceId) return false;
      const parent = uniqueTasksBySourceId.get(parentSourceId);
      return (
        parent?.teamId === teamId && failedStorySourceIds.has(parentSourceId)
      );
    });
    const blockedKeys = new Set(blocked.map(({ sourceKey }) => sourceKey));
    for (const item of blocked) {
      allResults.push({
        sourceKey: item.sourceKey,
        storyId: null,
        created: false,
        error: {
          code: "parent_import_failed",
          message: "The parent work item could not be imported.",
        },
      });
      failedStorySourceIds.add(item.sourceId);
    }
    pendingTasks = pendingTasks.filter(
      ({ sourceKey }) => !blockedKeys.has(sourceKey),
    );
    processedStories += blocked.length;

    const ready = pendingTasks.filter(({ parentSourceId, teamId }) => {
      if (!parentSourceId) return true;
      const parent = uniqueTasksBySourceId.get(parentSourceId);
      if (!parent || parent.teamId !== teamId) return true;
      return importedStoryIds.has(parentSourceId);
    });
    if (ready.length === 0 && pendingTasks.length > 0) {
      for (const item of pendingTasks) {
        allResults.push({
          sourceKey: item.sourceKey,
          storyId: null,
          created: false,
          error: {
            code: "parent_cycle",
            message: "The source contains a circular parent relationship.",
          },
        });
      }
      processedStories += pendingTasks.length;
      pendingTasks = [];
      break;
    }

    const readyKeys = new Set(ready.map(({ sourceKey }) => sourceKey));
    const preparedRequestItems = ready.map(
      ({ parentSourceId, provider, sourceId, sourceKey, task, teamId }) => {
        const context = getTeamContext(teamId);
        const identity = getTaskImportPerson(task, peopleBySourceId);
        const assignee = resolveReviewedPerson(
          identity,
          teamId,
          task.assigneePersonSourceId,
        );
        const collaboratorIds = [
          ...new Set(
            task.collaboratorPersonSourceIds.flatMap((personSourceId) => {
              const collaborator = resolveReviewedPerson(
                peopleBySourceId.get(personSourceId),
                teamId,
                personSourceId,
              );
              return collaborator && collaborator.id !== assignee?.id
                ? [collaborator.id]
                : [];
            }),
          ),
        ];
        const objective = task.objectiveSourceId
          ? objectiveMappings.get(task.objectiveSourceId)
          : undefined;
        const keyResult = task.keyResultSourceId
          ? keyResultMappings.get(task.keyResultSourceId)
          : undefined;
        const sprint = task.sprintSourceId
          ? sprintMappings.get(task.sprintSourceId)
          : undefined;
        const labelIds = task.labelSourceIds.flatMap((labelSourceId) => {
          const label = labelsBySourceId.get(labelSourceId);
          if (!label) return [];
          const labelId = labelMappings.get(getLabelMappingKey(label, teamId));
          return labelId ? [labelId] : [];
        });
        const parent = parentSourceId
          ? uniqueTasksBySourceId.get(parentSourceId)
          : undefined;
        const parentId =
          parentSourceId && parent?.teamId === teamId
            ? importedStoryIds.get(parentSourceId)
            : undefined;
        const linkedObjective = keyResult
          ? {
              id: keyResult.objectiveId,
              teamId: keyResult.teamId,
            }
          : objective;
        const taskWithSafeDates = isValidImportDateRange(
          task.startDate,
          task.endDate,
        )
          ? task
          : { ...task, startDate: null, endDate: null };
        return {
          collaboratorIds,
          provider,
          sourceId,
          sourceKey,
          requestItem: {
            sourceKey,
            story: toImportStoryPayload({
              allowAutomaticAssigneeResolution: false,
              ...(assignee ? { assigneeId: assignee.id } : {}),
              ...(linkedObjective ? { objectiveId: linkedObjective.id } : {}),
              ...(keyResult ? { keyResultId: keyResult.id } : {}),
              ...(sprint?.teamId === teamId ? { sprintId: sprint.id } : {}),
              ...(parentId ? { parentId } : {}),
              labelIds,
              members: context.members,
              statuses: context.statuses,
              task: taskWithSafeDates,
              teamId,
            }),
          },
        };
      },
    );
    const importRequests = (["jira_csv", "file"] as const).flatMap(
      (provider) => {
        const items = preparedRequestItems.flatMap((item) =>
          item.provider === provider ? [item.requestItem] : [],
        );
        return items.length > 0
          ? buildImportStoryRequests({
              items,
              provider,
              sourceDigest: draft.fileHash,
              ...(draft.sourceNamespace
                ? { sourceNamespace: draft.sourceNamespace }
                : {}),
            })
          : [];
      },
    );
    const readyBySourceKey = new Map(
      preparedRequestItems.map((item) => [item.sourceKey, item]),
    );
    for (const request of importRequests) {
      // eslint-disable-next-line no-await-in-loop -- Sequential batches cap load and make progress truthful.
      const response = await importStoriesBatch(request, ctx);
      if (response.error?.message || !response.data) {
        throw new Error(
          response.error?.message || "A batch could not be imported",
        );
      }
      allResults.push(...response.data.items);
      const collaboratorUpdates: Promise<number>[] = [];
      for (const result of response.data.items) {
        const item = readyBySourceKey.get(result.sourceKey);
        if (!item) continue;
        if (result.storyId && !result.error) {
          importedStoryIdsBySourceKey.set(result.sourceKey, result.storyId);
          if (sourceIdCounts.get(item.sourceId) === 1) {
            importedStoryIds.set(item.sourceId, result.storyId);
          }
          if (item.collaboratorIds.length > 0) {
            collaboratorUpdates.push(
              mergeImportStoryCollaborators(
                result.storyId,
                item.collaboratorIds,
                ctx,
              ),
            );
          }
        } else {
          failedStorySourceIds.add(item.sourceId);
        }
      }
      // eslint-disable-next-line no-await-in-loop -- Each story batch must finish collaborator reconciliation before advancing parent-dependent work.
      const appliedCounts = await Promise.all(collaboratorUpdates);
      appliedCollaborators += appliedCounts.reduce(
        (total, count) => total + count,
        0,
      );
      processedStories += request.items.length;
      onProgress(
        preparedTasks.length
          ? 70 + Math.round((processedStories / preparedTasks.length) * 25)
          : 95,
      );
    }
    pendingTasks = pendingTasks.filter(
      ({ sourceKey }) => !readyKeys.has(sourceKey),
    );
  }
  onProgress(95);

  return {
    allResults,
    preparedTasks,
    normalizeSourceId,
    sourceIdCounts,
    uniqueTasksBySourceId,
    importedStoryIds,
    importedStoryIdsBySourceKey,
    appliedCollaborators,
    destinationConflicts,
  };
};

export type ImportedStories = Awaited<ReturnType<typeof importStories>>;
