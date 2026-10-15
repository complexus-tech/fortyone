import type { QueryClient } from "@tanstack/react-query";
import { storyKeys } from "@/modules/stories/constants";
import type { DesktopCacheScope } from "./cache-types";
import type { DesktopOperation } from "./outbox-types";
import { isRecord, setChanged } from "./outbox-records";
import { findCachedStory, projectStory } from "./outbox-story-projection";

function removeResource(value: unknown, id: string): unknown {
  if (Array.isArray(value))
    return value
      .filter((item) => !isRecord(item) || item.id !== id)
      .map((item) => removeResource(item, id));
  if (!isRecord(value)) return value;
  const result = { ...value };
  for (const key of [
    "pages",
    "groups",
    "stories",
    "subStories",
    "comments",
    "subComments",
  ])
    if (Array.isArray(value[key])) result[key] = removeResource(value[key], id);
  if (Array.isArray(value.stories) && Array.isArray(result.stories)) {
    const removed = value.stories.length - result.stories.length;
    if (typeof value.totalCount === "number")
      result.totalCount = Math.max(0, value.totalCount - removed);
    if (typeof value.loadedCount === "number")
      result.loadedCount = Math.max(0, value.loadedCount - removed);
  }
  return result;
}

/** Call only after native discard succeeds, before projecting the remaining outbox. */
export function discardDesktopProjection(
  client: QueryClient,
  scope: DesktopCacheScope,
  operation: DesktopOperation,
) {
  if (operation.kind === "attachment.upload") {
    const attachmentId = operation.payload.attachmentId;
    if (typeof attachmentId !== "string") return;
    const key = storyKeys.attachments(
      scope.workspaceSlug,
      operation.resourceId,
    );
    const current = client.getQueryData(key);
    if (current) setChanged(client, key, removeResource(current, attachmentId));
    return;
  }
  if (operation.kind === "story.update") {
    if (!operation.serverValue)
      throw new Error(
        "Reconnect and load the server version before discarding this change.",
      );
    const current = findCachedStory(client, scope, operation.resourceId);
    if (!current) return;
    const restored = {
      ...current,
      ...operation.serverValue,
      id: operation.resourceId,
      desktopPending: false,
      desktopCreated: false,
    };
    setChanged(
      client,
      storyKeys.detail(scope.workspaceSlug, operation.resourceId),
      restored,
    );
    projectStory(client, scope, restored);
    return;
  }
  for (const [key, data] of client.getQueriesData({
    queryKey: storyKeys.all(scope.workspaceSlug),
  }))
    setChanged(client, key, removeResource(data, operation.resourceId));
  if (operation.kind === "story.create")
    client.removeQueries({
      queryKey: storyKeys.detail(scope.workspaceSlug, operation.resourceId),
    });
}
