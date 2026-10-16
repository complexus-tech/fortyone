import type { QueryClient } from "@tanstack/react-query";
import type {
  DetailedStory,
  NewStory,
  StoryUpdate,
} from "@/modules/story/types";
import type { Comment } from "@/types";
import type { Session } from "@/auth";
import { storyKeys } from "@/modules/stories/constants";
import { DEFAULT_ESTIMATE_SCHEME } from "@/lib/estimate";
import type { StoryCommentPayload } from "@/shared/story/comment-payload";
import type { DesktopCacheScope } from "./cache-types";
import type { DesktopEnqueueOperation } from "./outbox-types";
import { DESKTOP_OUTBOX_CHANGED } from "./outbox-types";
import { getDesktopBridge } from "./bridge";
import { bindDesktopSession } from "./session-binding";

const submissionIds = new WeakMap<object, Map<string, string>>();
function submissionId(scope: DesktopCacheScope, payload: object) {
  const key = `${scope.accountId}:${scope.workspaceId}`;
  const ids = submissionIds.get(payload) ?? new Map<string, string>();
  const id = ids.get(key) ?? crypto.randomUUID();
  ids.set(key, id);
  submissionIds.set(payload, ids);
  return id;
}

export const isDesktopPending = (value: unknown): boolean =>
  Boolean(
    value &&
      typeof value === "object" &&
      "desktopPending" in value &&
      value.desktopPending === true,
  );

export function localStory(
  payload: NewStory,
  id: string,
  scope: DesktopCacheScope,
  createdAt: string,
): DetailedStory & { desktopPending: boolean; desktopCreated: boolean } {
  return {
    title: "Untitled",
    description: "",
    descriptionHTML: "",
    parentId: "",
    estimateLabel: null,
    estimateValue: null,
    estimateScheme: DEFAULT_ESTIMATE_SCHEME,
    estimatedDurationMinutes: null,
    minimumFocusBlockMinutes: null,
    autoSchedulingEnabled: false,
    autoSchedulingLocked: false,
    autoSchedulingStatus: "off",
    autoSchedulingReason: null,
    autoSchedulingUpdatedAt: null,
    teamId: "",
    objectiveId: null,
    keyResultId: null,
    statusId: "",
    assigneeId: null,
    collaboratorIds: [],
    collaborators: [],
    collaboratorCount: 0,
    watcherCount: 0,
    watchers: [],
    isWatching: false,
    watchingReason: null,
    reporterId: scope.accountId,
    priority: "No Priority",
    sprintId: null,
    epicId: null,
    startDate: null,
    endDate: null,
    createdAt,
    updatedAt: createdAt,
    deletedAt: null,
    completedAt: null,
    archivedAt: null,
    subStories: [],
    labels: payload.labelIds ?? [],
    associations: [],
    ...payload,
    id,
    workspaceId: scope.workspaceId,
    teamCode: "",
    sequenceId: 0,
    desktopPending: true,
    desktopCreated: true,
  };
}

function cachedStory(
  value: unknown,
  id: string,
): { updatedAt: string; desktopCreated?: boolean } | null {
  if (!value || typeof value !== "object") return null;
  if (Array.isArray(value)) {
    for (const item of value) {
      const found = cachedStory(item, id);
      if (found) return found;
    }
    return null;
  }
  const record = value as Record<string, unknown>;
  if (
    record.id === id &&
    typeof record.title === "string" &&
    typeof record.updatedAt === "string"
  ) {
    return {
      updatedAt: record.updatedAt,
      desktopCreated: record.desktopCreated === true,
    };
  }
  for (const key of ["pages", "groups", "stories", "subStories"]) {
    const found = cachedStory(record[key], id);
    if (found) return found;
  }
  return null;
}

async function enqueue(
  scope: DesktopCacheScope,
  operation: Omit<DesktopEnqueueOperation, "accountId" | "workspaceId">,
) {
  const bridge = getDesktopBridge();
  if (!bridge?.enqueueOperation)
    throw new Error("Update the desktop app to save changes offline.");
  await bindDesktopSession(scope);
  const result = await bridge.enqueueOperation({
    accountId: scope.accountId,
    workspaceId: scope.workspaceId,
    ...operation,
  });
  window.dispatchEvent(
    new CustomEvent(DESKTOP_OUTBOX_CHANGED, { detail: scope }),
  );
  return result;
}

export async function createDesktopStory(
  scope: DesktopCacheScope,
  payload: NewStory,
  consumedDraftKey?: string,
) {
  const operationId = payload.idempotencyKey || submissionId(scope, payload);
  const resourceId = operationId;
  const result = await enqueue(scope, {
    id: operationId,
    kind: "story.create",
    resourceId,
    payload,
    consumedDraftKey,
  });
  const local = localStory(
    payload,
    result.resourceId,
    scope,
    new Date().toISOString(),
  );
  return result.pending
    ? local
    : {
        ...local,
        ...result.resource,
        id: result.resourceId,
        desktopPending: false,
        desktopCreated: false,
      };
}

export async function updateDesktopStory(
  scope: DesktopCacheScope,
  client: QueryClient,
  storyId: string,
  payload: StoryUpdate,
  baseVersion?: string,
) {
  let current = cachedStory(
    client.getQueryData(storyKeys.detail(scope.workspaceSlug, storyId)),
    storyId,
  );
  if (!current) {
    for (const [, value] of client.getQueriesData({
      queryKey: storyKeys.all(scope.workspaceSlug),
    })) {
      current = cachedStory(value, storyId);
      if (current) break;
    }
  }
  if (!current)
    throw new Error(
      "Open this item while connected before editing it offline.",
    );
  const fields = { ...payload };
  delete fields.id;
  await enqueue(scope, {
    id: crypto.randomUUID(),
    kind: "story.update",
    resourceId: storyId,
    payload: fields,
    baseVersion:
      baseVersion ??
      (current.desktopCreated
        ? `pending-create:${storyId}`
        : current.updatedAt),
  });
  return { data: null, desktopPending: true };
}

export async function commentDesktopStory(
  scope: DesktopCacheScope,
  session: Session | null,
  storyId: string,
  payload: StoryCommentPayload,
) {
  const id = payload.creationId ?? submissionId(scope, payload);
  const result = await enqueue(scope, {
    id,
    kind: "comment.create",
    resourceId: id,
    payload: { ...payload, storyId },
  });
  const now = new Date().toISOString();
  const comment: Comment = {
    id: result.resourceId,
    storyId,
    parentId: payload.parentId ?? null,
    userId: scope.accountId,
    user: {
      id: scope.accountId,
      fullName: session?.user.name ?? "",
      username: session?.user.username ?? "",
      avatarUrl: session?.user.image ?? null,
      isActive: true,
      isSystem: false,
    },
    comment: payload.comment,
    createdAt: now,
    updatedAt: now,
    subComments: [],
  };
  return {
    data: { ...comment, ...result.resource, id: result.resourceId },
    desktopPending: result.pending,
  };
}
