import type {
  AutoSchedulingStatus,
  StoryPriority,
} from "@/modules/stories/types";

type StoryWorkspaceChanges = {
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

  if (isNonEmptyString(value.entityId) && isNonEmptyString(value.entityType)) {
    return {
      kind: "notification",
      entityId: value.entityId,
      entityType: value.entityType,
    };
  }

  throw new WorkspaceRealtimeContractError(
    "Workspace realtime event has an unsupported shape",
  );
};
