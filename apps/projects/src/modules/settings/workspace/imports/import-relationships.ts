import type { StoryAssociationType } from "@/shared/story/types";
import type { RunImportInput } from "./import-run-model";
import type { ImportedStories } from "./import-stories";
import { normalizeImportTaskLinks } from "./schema";
import {
  getCanonicalImportAssociation,
  getImportAssociationKey,
} from "./import-association-model";
import {
  getImportStoryLinks,
  createImportStoryLink,
  getImportStoryAssociations,
  createImportStoryAssociation,
} from "./api";

export const importRelationships = async (
  { ctx, onProgress }: Pick<RunImportInput, "ctx" | "onProgress">,
  {
    preparedTasks,
    normalizeSourceId,
    sourceIdCounts,
    uniqueTasksBySourceId,
    importedStoryIds,
    importedStoryIdsBySourceKey,
  }: Pick<
    ImportedStories,
    | "preparedTasks"
    | "normalizeSourceId"
    | "sourceIdCounts"
    | "uniqueTasksBySourceId"
    | "importedStoryIds"
    | "importedStoryIdsBySourceKey"
  >,
) => {
  let createdLinks = 0;
  let createdAssociations = 0;
  let destinationConflicts = 0;
  let unresolvedAssociations = 0;
  let unresolvedLinks = 0;
  for (const item of preparedTasks) {
    const links = normalizeImportTaskLinks(item.task.links);
    if (links.length === 0) continue;
    const storyId = importedStoryIdsBySourceKey.get(item.sourceKey);
    if (!storyId) {
      unresolvedLinks += links.length;
      continue;
    }

    let existingUrls: Set<string>;
    try {
      // eslint-disable-next-line no-await-in-loop -- Each story must be inspected before link writes so retries cannot duplicate existing links.
      const existingLinks = await getImportStoryLinks(storyId, ctx);
      existingUrls = new Set(
        normalizeImportTaskLinks(
          existingLinks.map((link) => ({
            title: link.title,
            url: link.url,
          })),
        ).map((link) => link.url),
      );
    } catch {
      unresolvedLinks += links.length;
      continue;
    }

    for (const link of links) {
      if (existingUrls.has(link.url)) continue;
      try {
        // eslint-disable-next-line no-await-in-loop -- Link writes are intentionally sequential and exact-URL deduplicated for safe partial retry.
        await createImportStoryLink(
          {
            storyId,
            url: link.url,
            ...(link.title ? { title: link.title } : {}),
          },
          ctx,
        );
        existingUrls.add(link.url);
        createdLinks += 1;
      } catch {
        unresolvedLinks += 1;
      }
    }
  }

  const plannedAssociations = new Map<
    string,
    { fromStoryId: string; toStoryId: string; type: StoryAssociationType }
  >();
  const seenSourceAssociationKeys = new Set<string>();
  for (const item of preparedTasks) {
    for (const association of item.task.associations) {
      const targetSourceId = normalizeSourceId(association.targetSourceId);
      const sourceVertexId =
        sourceIdCounts.get(item.sourceId) === 1
          ? item.sourceId
          : item.sourceKey;
      const sourceAssociation = getCanonicalImportAssociation(
        sourceVertexId,
        targetSourceId,
        association.type,
      );
      const sourceAssociationKey = getImportAssociationKey(sourceAssociation);
      if (seenSourceAssociationKeys.has(sourceAssociationKey)) continue;
      seenSourceAssociationKeys.add(sourceAssociationKey);

      const sourceTask = uniqueTasksBySourceId.get(item.sourceId);
      const targetTask = uniqueTasksBySourceId.get(targetSourceId);
      const sourceStoryId = importedStoryIds.get(item.sourceId);
      const targetStoryId = importedStoryIds.get(targetSourceId);
      if (
        sourceIdCounts.get(item.sourceId) !== 1 ||
        sourceIdCounts.get(targetSourceId) !== 1 ||
        !sourceTask ||
        !targetTask ||
        !sourceStoryId ||
        !targetStoryId ||
        sourceStoryId === targetStoryId
      ) {
        unresolvedAssociations += 1;
        continue;
      }
      if (sourceTask.teamId !== targetTask.teamId) {
        unresolvedAssociations += 1;
        destinationConflicts += 1;
        continue;
      }

      const destinationAssociation = getCanonicalImportAssociation(
        sourceStoryId,
        targetStoryId,
        association.type,
      );
      plannedAssociations.set(getImportAssociationKey(destinationAssociation), {
        fromStoryId: destinationAssociation.fromId,
        toStoryId: destinationAssociation.toId,
        type: destinationAssociation.type,
      });
    }
  }

  const associationStoryIds = new Set<string>();
  for (const association of plannedAssociations.values()) {
    associationStoryIds.add(association.fromStoryId);
    associationStoryIds.add(association.toStoryId);
  }
  const existingAssociationGroups = await Promise.all(
    [...associationStoryIds].map((storyId) =>
      getImportStoryAssociations(storyId, ctx),
    ),
  );
  const existingAssociationKeys = new Set(
    existingAssociationGroups.flatMap((associations) =>
      associations.map((association) =>
        getImportAssociationKey({
          fromId: association.fromStoryId,
          toId: association.toStoryId,
          type: association.type,
        }),
      ),
    ),
  );
  for (const [associationKey, association] of plannedAssociations) {
    if (existingAssociationKeys.has(associationKey)) continue;
    // eslint-disable-next-line no-await-in-loop -- Association writes are ordered so a partial failure is safely discoverable and reusable on retry.
    await createImportStoryAssociation(
      association.fromStoryId,
      { toStoryId: association.toStoryId, type: association.type },
      ctx,
    );
    existingAssociationKeys.add(associationKey);
    createdAssociations += 1;
  }
  onProgress(100);

  return {
    createdLinks,
    createdAssociations,
    destinationConflicts,
    unresolvedAssociations,
    unresolvedLinks,
  };
};
