import type { AutoSchedulingStatus } from "@/modules/stories/types";

type AutoSchedulingState = {
  assigneeId?: string | null;
  autoSchedulingEnabled?: boolean;
  autoSchedulingLocked?: boolean;
  autoSchedulingStatus?: AutoSchedulingStatus | null;
  estimatedDurationMinutes?: number | null;
};

export const AUTO_SCHEDULING_STATUS_LABELS: Record<
  AutoSchedulingStatus,
  string
> = {
  off: "Off",
  needs_owner: "Needs owner",
  needs_time: "Needs time",
  planning: "Planning",
  scheduled: "Scheduled",
  at_risk: "At risk",
  cannot_fit: "Can't fit",
  locked: "Locked",
};

export const AUTO_SCHEDULING_STATUS_HELPERS: Record<
  AutoSchedulingStatus,
  string
> = {
  off: "Maya will not place this work on a calendar.",
  needs_owner: "Choose an assignee so Maya knows whose calendar to plan.",
  needs_time:
    "Add time needed so Maya knows how much calendar space to reserve.",
  planning: "Maya is looking for the best available focus time.",
  scheduled: "Maya has reserved focus time and will keep it up to date.",
  at_risk: "Some planned work may no longer fit before its deadline.",
  cannot_fit:
    "Maya could not find enough available time in the planning window.",
  locked: "The current Maya blocks stay fixed until you unlock them.",
};

const isActiveServerStatus = (
  status?: AutoSchedulingStatus | null,
): status is Exclude<AutoSchedulingStatus, "off" | "locked"> =>
  Boolean(status && status !== "off" && status !== "locked");

const isActionableServerStatus = (
  status?: AutoSchedulingStatus | null,
): status is Extract<
  AutoSchedulingStatus,
  "needs_owner" | "needs_time" | "at_risk" | "cannot_fit"
> =>
  status === "needs_owner" ||
  status === "needs_time" ||
  status === "at_risk" ||
  status === "cannot_fit";

export const deriveAutoSchedulingStatus = ({
  assigneeId,
  autoSchedulingEnabled = false,
  autoSchedulingLocked = false,
  autoSchedulingStatus,
  estimatedDurationMinutes,
}: AutoSchedulingState): AutoSchedulingStatus => {
  if (!autoSchedulingEnabled) return autoSchedulingLocked ? "locked" : "off";
  if (isActionableServerStatus(autoSchedulingStatus)) {
    return autoSchedulingStatus;
  }
  if (autoSchedulingLocked) return "locked";
  if (isActiveServerStatus(autoSchedulingStatus)) return autoSchedulingStatus;
  if (!assigneeId) return "needs_owner";
  if (!estimatedDurationMinutes) return "needs_time";
  return "planning";
};

export const getAutoSchedulingLabel = (
  status?: AutoSchedulingStatus | null,
  options?: { planningLabel?: string },
) =>
  status === "planning" && options?.planningLabel
    ? options.planningLabel
    : AUTO_SCHEDULING_STATUS_LABELS[status ?? "planning"];

export const getAutoSchedulingHelper = (
  status: AutoSchedulingStatus,
  reason?: string | null,
) => reason?.trim() || AUTO_SCHEDULING_STATUS_HELPERS[status];

export const canLockAutoSchedulingStatus = (status: AutoSchedulingStatus) =>
  status === "scheduled" || status === "locked";

export const canToggleAutoSchedulingLock = (
  status: AutoSchedulingStatus,
  autoSchedulingLocked: boolean,
) => autoSchedulingLocked || canLockAutoSchedulingStatus(status);

export const isMayaAssigneeSelection = (
  assigneeId?: string | null,
  mayaAssigneeId?: string | null,
) => Boolean(assigneeId && mayaAssigneeId && assigneeId === mayaAssigneeId);

export const getNewStoryAutoSchedulingEnabled = ({
  currentEnabled,
  mayaAssigneeId,
  selectedAssigneeId,
}: {
  currentEnabled: boolean;
  mayaAssigneeId?: string | null;
  selectedAssigneeId?: string | null;
}) => {
  if (isMayaAssigneeSelection(selectedAssigneeId, mayaAssigneeId)) return true;
  return currentEnabled;
};
