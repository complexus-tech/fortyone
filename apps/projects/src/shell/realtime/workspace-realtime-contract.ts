import type {
  AutoSchedulingStatus,
  StoryPriority,
} from "@/modules/stories/types";

type StoryWorkspaceChanges = {
  checklistChanged?: true;
  statusId?: string;
  completedAt?: string | null;
  assigneeId?: string | null;
  priority?: StoryPriority;
  title?: string;
  autoSchedulingEnabled?: boolean;
  autoSchedulingLocked?: boolean;
  autoSchedulingStatus?: AutoSchedulingStatus;
  autoSchedulingReason?: string | null;
  autoSchedulingUpdatedAt?: string | null;
};

export type WorkspaceRealtimeEvent =
  | { kind: "calendar-updated" }
  | {
      kind: "chat-updated";
      workspaceId: string;
      userId: string;
      conversationId: string;
    }
  | {
      kind: "chat-typing";
      workspaceId: string;
      userId: string;
      conversationId: string;
      actorId: string;
      fullName: string;
      threadRootId: string | null;
      isTyping: boolean;
      expiresAt: string;
    }
  | {
      kind: "story-updated";
      storyId: string;
      changes: StoryWorkspaceChanges;
    }
  | {
      kind: "notification";
      entityId: string;
      entityType: string;
    };

export class WorkspaceRealtimeContractError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "WorkspaceRealtimeContractError";
  }
}

const AUTO_SCHEDULING_STATUSES = new Set<AutoSchedulingStatus>([
  "off",
  "needs_owner",
  "needs_time",
  "planning",
  "scheduled",
  "at_risk",
  "cannot_fit",
  "locked",
]);

const STORY_PRIORITIES = new Set<StoryPriority>([
  "No Priority",
  "Urgent",
  "High",
  "Medium",
  "Low",
]);

const STORY_CHANGE_KEYS: ReadonlySet<string> = new Set([
  "checklistChanged",
  "statusId",
  "completedAt",
  "assigneeId",
  "priority",
  "title",
  "autoSchedulingEnabled",
  "autoSchedulingLocked",
  "autoSchedulingStatus",
  "autoSchedulingReason",
  "autoSchedulingUpdatedAt",
]);

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === "object" && value !== null && !Array.isArray(value);

const isNonEmptyString = (value: unknown): value is string =>
  typeof value === "string" && value.trim().length > 0;

const isNullableString = (value: unknown): value is string | null =>
  value === null || typeof value === "string";

const isUUID = (value: unknown): value is string =>
  typeof value === "string" &&
  /^[\da-f]{8}-[\da-f]{4}-[\da-f]{4}-[\da-f]{4}-[\da-f]{12}$/i.test(value);

const decodeStoryChanges = (value: unknown): StoryWorkspaceChanges => {
  if (!isRecord(value)) {
    throw new WorkspaceRealtimeContractError(
      "Story workspace update changes must be an object",
    );
  }

  const keys = Object.keys(value);
  if (keys.length === 0 || keys.some((key) => !STORY_CHANGE_KEYS.has(key))) {
    throw new WorkspaceRealtimeContractError(
      "Story workspace update contains unsupported changes",
    );
  }

  const changes: StoryWorkspaceChanges = {};

  if ("checklistChanged" in value) {
    if (value.checklistChanged !== true) {
      throw new WorkspaceRealtimeContractError("Invalid checklist update");
    }
    changes.checklistChanged = true;
  }

  if ("statusId" in value) {
    if (!isNonEmptyString(value.statusId)) {
      throw new WorkspaceRealtimeContractError("Invalid story status update");
    }
    changes.statusId = value.statusId;
  }
  if ("completedAt" in value) {
    if (!isNullableString(value.completedAt)) {
      throw new WorkspaceRealtimeContractError(
        "Invalid story completion update",
      );
    }
    changes.completedAt = value.completedAt;
  }
  if ("assigneeId" in value) {
    if (!isNullableString(value.assigneeId)) {
      throw new WorkspaceRealtimeContractError("Invalid story assignee update");
    }
    changes.assigneeId = value.assigneeId;
  }
  if ("priority" in value) {
    if (
      typeof value.priority !== "string" ||
      !STORY_PRIORITIES.has(value.priority as StoryPriority)
    ) {
      throw new WorkspaceRealtimeContractError("Invalid story priority update");
    }
    changes.priority = value.priority as StoryPriority;
  }
  if ("title" in value) {
    if (!isNonEmptyString(value.title)) {
      throw new WorkspaceRealtimeContractError("Invalid story title update");
    }
    changes.title = value.title;
  }
  if ("autoSchedulingEnabled" in value) {
    if (typeof value.autoSchedulingEnabled !== "boolean") {
      throw new WorkspaceRealtimeContractError(
        "Invalid auto-scheduling enabled update",
      );
    }
    changes.autoSchedulingEnabled = value.autoSchedulingEnabled;
  }
  if ("autoSchedulingLocked" in value) {
    if (typeof value.autoSchedulingLocked !== "boolean") {
      throw new WorkspaceRealtimeContractError(
        "Invalid auto-scheduling lock update",
      );
    }
    changes.autoSchedulingLocked = value.autoSchedulingLocked;
  }
  if ("autoSchedulingStatus" in value) {
    if (
      typeof value.autoSchedulingStatus !== "string" ||
      !AUTO_SCHEDULING_STATUSES.has(
        value.autoSchedulingStatus as AutoSchedulingStatus,
      )
    ) {
      throw new WorkspaceRealtimeContractError(
        "Invalid auto-scheduling status update",
      );
    }
    changes.autoSchedulingStatus =
      value.autoSchedulingStatus as AutoSchedulingStatus;
  }
  if ("autoSchedulingReason" in value) {
    if (!isNullableString(value.autoSchedulingReason)) {
      throw new WorkspaceRealtimeContractError(
        "Invalid auto-scheduling reason update",
      );
    }
    changes.autoSchedulingReason = value.autoSchedulingReason;
  }
  if ("autoSchedulingUpdatedAt" in value) {
    if (!isNullableString(value.autoSchedulingUpdatedAt)) {
      throw new WorkspaceRealtimeContractError(
        "Invalid auto-scheduling timestamp update",
      );
    }
    changes.autoSchedulingUpdatedAt = value.autoSchedulingUpdatedAt;
  }

  return changes;
};

export const parseWorkspaceRealtimeEvent = (
  serializedEvent: unknown,
): WorkspaceRealtimeEvent => {
  if (typeof serializedEvent !== "string") {
    throw new WorkspaceRealtimeContractError(
      "Workspace realtime event must be a JSON string",
    );
  }

  let value: unknown;
  try {
    value = JSON.parse(serializedEvent) as unknown;
  } catch {
    throw new WorkspaceRealtimeContractError(
      "Workspace realtime event is not valid JSON",
    );
  }

  if (!isRecord(value)) {
    throw new WorkspaceRealtimeContractError(
      "Workspace realtime event must be an object",
    );
  }

  if (value.type === "calendar.updated") {
    return { kind: "calendar-updated" };
  }

  if (value.type === "chat.updated") {
    if (
      ![
        value.eventId,
        value.workspaceId,
        value.userId,
        value.conversationId,
      ].every(isUUID)
    ) {
      throw new WorkspaceRealtimeContractError(
        "Chat update requires scoped IDs",
      );
    }
    return {
      kind: "chat-updated",
      workspaceId: value.workspaceId as string,
      userId: value.userId as string,
      conversationId: value.conversationId as string,
    };
  }

  if (value.type === "chat.typing") {
    if (
      !isUUID(value.workspaceId) ||
      !isUUID(value.userId) ||
      !isUUID(value.conversationId) ||
      !isUUID(value.actorId) ||
      !(value.threadRootId === null || isUUID(value.threadRootId)) ||
      !isNonEmptyString(value.fullName) ||
      value.fullName.length > 200 ||
      typeof value.isTyping !== "boolean" ||
      typeof value.expiresAt !== "string" ||
      !Number.isFinite(Date.parse(value.expiresAt))
    ) {
      throw new WorkspaceRealtimeContractError(
        "Invalid scoped chat typing activity",
      );
    }
    return {
      kind: "chat-typing",
      workspaceId: value.workspaceId,
      userId: value.userId,
      conversationId: value.conversationId,
      actorId: value.actorId,
      fullName: value.fullName,
      threadRootId: value.threadRootId,
      isTyping: value.isTyping,
      expiresAt: value.expiresAt,
    };
  }

  if (value.type === "story.workspace_update") {
    if (!isNonEmptyString(value.storyId)) {
      throw new WorkspaceRealtimeContractError(
        "Story workspace update requires a story ID",
      );
    }

    return {
      kind: "story-updated",
      storyId: value.storyId,
      changes: decodeStoryChanges(value.changes),
    };
  }

  const entityId = value.entityId ?? value.entity_id;
  const entityType = value.entityType ?? value.entity_type;
  if (isNonEmptyString(entityId) && isNonEmptyString(entityType)) {
    return {
      kind: "notification",
      entityId,
      entityType,
    };
  }

  throw new WorkspaceRealtimeContractError(
    "Workspace realtime event has an unsupported shape",
  );
};
