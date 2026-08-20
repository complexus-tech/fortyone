import type { CalendarScheduleIssue } from "@/lib/queries/calendar/types";

const SCHEDULE_ISSUE_DISMISSAL_PREFIX = "fortyone:schedule-issue-dismissal:";
export const SCHEDULE_ISSUE_DISMISSAL_TTL_MS = 24 * 60 * 60 * 1000;
const SCHEDULE_ISSUE_DISMISSAL_VERSION = 1;

type ScheduleIssueDismissal = {
  dismissedUntil: number;
  issueUpdatedAt: string;
  version: typeof SCHEDULE_ISSUE_DISMISSAL_VERSION;
};

const getScheduleIssueDismissalKey = (workspaceSlug: string, storyId: string) =>
  `${SCHEDULE_ISSUE_DISMISSAL_PREFIX}${encodeURIComponent(workspaceSlug)}:${encodeURIComponent(storyId)}`;

export const getScheduleIssueDismissalToken = (issue: CalendarScheduleIssue) =>
  `${issue.storyId}:${issue.updatedAt}`;

const parseScheduleIssueDismissal = (
  value: string,
): ScheduleIssueDismissal | null => {
  try {
    const parsed: unknown = JSON.parse(value);
    if (!parsed || typeof parsed !== "object") return null;

    const dismissal = parsed as Record<string, unknown>;
    if (
      dismissal.version !== SCHEDULE_ISSUE_DISMISSAL_VERSION ||
      typeof dismissal.issueUpdatedAt !== "string" ||
      typeof dismissal.dismissedUntil !== "number" ||
      !Number.isFinite(dismissal.dismissedUntil)
    ) {
      return null;
    }

    return {
      dismissedUntil: dismissal.dismissedUntil,
      issueUpdatedAt: dismissal.issueUpdatedAt,
      version: SCHEDULE_ISSUE_DISMISSAL_VERSION,
    };
  } catch {
    return null;
  }
};

export const isScheduleIssueDismissed = (
  issue: CalendarScheduleIssue,
  workspaceSlug: string,
  now: number,
  sessionDismissals: ReadonlyMap<string, number>,
) => {
  const sessionDismissedUntil = sessionDismissals.get(
    getScheduleIssueDismissalToken(issue),
  );
  if (sessionDismissedUntil && sessionDismissedUntil > now) return true;

  try {
    const storedValue = window.localStorage.getItem(
      getScheduleIssueDismissalKey(workspaceSlug, issue.storyId),
    );
    if (!storedValue) return false;

    const dismissal = parseScheduleIssueDismissal(storedValue);
    return Boolean(
      dismissal &&
        dismissal.issueUpdatedAt === issue.updatedAt &&
        dismissal.dismissedUntil > now,
    );
  } catch {
    return false;
  }
};

export const dismissScheduleIssue = (
  issue: CalendarScheduleIssue,
  workspaceSlug: string,
  dismissedAt: number,
) => {
  const dismissedUntil = dismissedAt + SCHEDULE_ISSUE_DISMISSAL_TTL_MS;
  const dismissal: ScheduleIssueDismissal = {
    dismissedUntil,
    issueUpdatedAt: issue.updatedAt,
    version: SCHEDULE_ISSUE_DISMISSAL_VERSION,
  };

  try {
    window.localStorage.setItem(
      getScheduleIssueDismissalKey(workspaceSlug, issue.storyId),
      JSON.stringify(dismissal),
    );
  } catch {
    // The in-memory dismissal still applies when storage is unavailable.
  }

  return dismissedUntil;
};
