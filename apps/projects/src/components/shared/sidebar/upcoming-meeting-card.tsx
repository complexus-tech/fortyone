"use client";

import type { ReactNode } from "react";
import { useEffect, useRef, useState, useSyncExternalStore } from "react";
import { addDays, format, startOfDay } from "date-fns";
import Link from "next/link";
import {
  ChevronLeftIcon,
  ChevronRightIcon,
  CloseIcon,
  ExternalLinkIcon,
  Video02Icon,
  WarningIcon,
} from "icons";
import { Box, Button, Flex, Popover, Text } from "ui";
import { useTerminology, useWorkspacePath } from "@/hooks";
import { ScheduleIssueDialog } from "@/components/ui/story/schedule-issue-dialog";
import { useSession } from "@/lib/auth/client";
import {
  useCalendarIntegration,
  useCalendarSchedule,
  useOverrideCalendarScheduleIssue,
} from "@/lib/hooks/calendar";
import { formatTimeNeeded } from "@/lib/time-needed";
import type { CalendarScheduleIssue } from "@/lib/queries/calendar/types";
import { useObjectiveStatuses } from "@/lib/hooks/objective-statuses";
import {
  getObjectiveForecastRiskCopy,
  isObjectiveForecastAtRisk,
} from "@/modules/objectives/components/objective-forecast-risk-utils";
import { useObjectives } from "@/modules/objectives/hooks/use-objectives";
import type { Objective, ObjectiveStatus } from "@/modules/objectives/types";
import { getStoryPath } from "@/modules/story/utils/story-url";
import {
  dismissMeetingUntilEnd,
  getUpcomingMeetings,
  isMeetingDismissed,
} from "./upcoming-meeting";
import type { UpcomingMeeting } from "./upcoming-meeting";
import {
  dismissScheduleIssue,
  getScheduleIssueDismissalToken,
  isScheduleIssueDismissed,
} from "./schedule-issue-dismissal";

const CLOCK_REFRESH_INTERVAL_MS = 30 * 1000;
const OBJECTIVE_RISK_DISMISSAL_PREFIX = "fortyone:objective-forecast-risk:";
const EMPTY_OBJECTIVES: Objective[] = [];
const EMPTY_OBJECTIVE_STATUSES: ObjectiveStatus[] = [];
const subscribeToHydration = () => () => undefined;
let clockSnapshot = typeof window === "undefined" ? 0 : Date.now();
let clockInterval: number | undefined;
const clockSubscribers = new Set<() => void>();

const refreshClock = () => {
  clockSnapshot = Date.now();
  clockSubscribers.forEach((subscriber) => {
    subscriber();
  });
};

const subscribeToClock = (subscriber: () => void) => {
  clockSubscribers.add(subscriber);
  if (clockSubscribers.size === 1) {
    refreshClock();
    clockInterval = window.setInterval(refreshClock, CLOCK_REFRESH_INTERVAL_MS);
    window.addEventListener("focus", refreshClock);
  }

  return () => {
    clockSubscribers.delete(subscriber);
    if (clockSubscribers.size === 0) {
      if (clockInterval) window.clearInterval(clockInterval);
      clockInterval = undefined;
      window.removeEventListener("focus", refreshClock);
    }
  };
};

const getClockSnapshot = () => clockSnapshot;
const getServerClockSnapshot = () => 0;

type SidebarAssistantItem =
  | { id: string; kind: "meeting"; meeting: UpcomingMeeting }
  | { id: string; issue: CalendarScheduleIssue; kind: "schedule-issue" }
  | { id: string; kind: "objective-risk"; objective: Objective };

const getDismissLabel = (item: SidebarAssistantItem) => {
  switch (item.kind) {
    case "meeting":
      return "Dismiss meeting";
    case "schedule-issue":
      return "Dismiss scheduling message";
    case "objective-risk":
      return "Dismiss objective forecast risk";
  }
};

const getCollapsedIndicatorLabel = (item: SidebarAssistantItem) => {
  switch (item.kind) {
    case "meeting":
      return "Open upcoming meeting";
    case "schedule-issue":
      return "Open Maya scheduling message";
    case "objective-risk":
      return "Open objective forecast risk";
  }
};

const getObjectiveRiskVersion = (objective: Objective) =>
  `${objective.forecastEndDate ?? "unknown"}:${objective.forecastDaysDelta}`;
const getObjectiveRiskToken = (objective: Objective) =>
  `${objective.id}:${getObjectiveRiskVersion(objective)}`;

const isObjectiveRiskDismissed = (objective: Objective) => {
  try {
    return (
      window.localStorage.getItem(
        `${OBJECTIVE_RISK_DISMISSAL_PREFIX}${objective.id}`,
      ) === getObjectiveRiskVersion(objective)
    );
  } catch {
    return false;
  }
};

const dismissObjectiveRisk = (objective: Objective) => {
  try {
    window.localStorage.setItem(
      `${OBJECTIVE_RISK_DISMISSAL_PREFIX}${objective.id}`,
      getObjectiveRiskVersion(objective),
    );
  } catch {
    // The in-memory dismissal still applies when storage is unavailable.
  }
};

const clearObjectiveRiskDismissal = (objectiveId: string) => {
  try {
    window.localStorage.removeItem(
      `${OBJECTIVE_RISK_DISMISSAL_PREFIX}${objectiveId}`,
    );
  } catch {
    // Storage is optional; there is nothing else to clear.
  }
};

const getMeetingTitle = (title?: string) => title?.trim() || "Upcoming meeting";

const MeetingContent = ({ meeting }: { meeting: UpcomingMeeting }) => {
  const startAt = new Date(meeting.event.startAt);
  const endAt = new Date(meeting.event.endAt);
  const statusLabel =
    meeting.status === "in-progress"
      ? "Happening now"
      : `Starts in ${meeting.minutesUntilStart} min`;

  return (
    <>
      <Flex align="center" className="gap-1.5 pr-6">
        <span className="bg-primary relative flex size-2 rounded-full">
          <span className="bg-primary absolute inset-0 animate-ping rounded-full opacity-40 motion-reduce:animate-none" />
        </span>
        <Text
          className="text-[0.72rem] tracking-[0.08em] uppercase"
          color="muted"
          fontWeight="semibold"
        >
          Meeting
        </Text>
        <Text
          className="bg-state-selected ml-auto rounded-md px-1.5 py-0.5 text-[0.78rem] tabular-nums"
          fontWeight="medium"
        >
          {statusLabel}
        </Text>
      </Flex>

      <Text className="mt-2 line-clamp-1 leading-snug" fontWeight="medium">
        {getMeetingTitle(meeting.event.title)}
      </Text>
      <Text className="mt-1 tabular-nums" color="muted" fontSize="sm">
        {format(startAt, "h:mm a")}–{format(endAt, "h:mm a")}
      </Text>

      <a
        className="bg-background-inverse text-foreground-inverse focus-visible:ring-ring mt-3 flex h-8.5 w-full items-center justify-center gap-1.5 rounded-lg px-2.5 text-[0.94rem] font-medium transition-opacity hover:opacity-90 focus-visible:ring-2 focus-visible:outline-none"
        href={meeting.meetingUrl}
        rel="noreferrer noopener"
        target="_blank"
      >
        <Video02Icon
          aria-hidden="true"
          className="text-foreground-inverse dark:text-foreground-inverse h-4 w-auto"
          strokeWidth={2}
        />
        Join meeting
        <ExternalLinkIcon
          aria-hidden="true"
          className="text-foreground-inverse dark:text-foreground-inverse h-3.5 w-auto"
          strokeWidth={2}
        />
      </a>
    </>
  );
};

const ScheduleIssueContent = ({
  issue,
  onChooseTime,
}: {
  issue: CalendarScheduleIssue;
  onChooseTime: () => void;
}) => {
  const { withWorkspace } = useWorkspacePath();
  const remainingMinutes =
    issue.remainingDurationMinutes ?? issue.estimatedDurationMinutes;
  const remainingDuration = remainingMinutes
    ? formatTimeNeeded(remainingMinutes)
    : null;
  const progress = remainingDuration
    ? `${remainingDuration} left to schedule.`
    : null;
  const schedulingReason = issue.autoSchedulingReason?.trim();
  const isProgressOnlyReason =
    Boolean(schedulingReason) &&
    schedulingReason?.toLowerCase() === progress?.toLowerCase();
  let description = schedulingReason;
  if (!description || isProgressOnlyReason) {
    description = remainingDuration
      ? `${remainingDuration} remains. Choose a time or let Maya try again.`
      : "Choose a time or let Maya try again.";
  }
  const descriptionTitle =
    schedulingReason && !isProgressOnlyReason && progress
      ? `${progress} ${description}`
      : description;

  return (
    <>
      <Flex align="center" className="gap-1.5 pr-6">
        <span className="bg-warning flex size-2 rounded-full" />
        <Text
          className="text-[0.72rem] tracking-[0.08em] uppercase"
          color="muted"
          fontWeight="semibold"
        >
          Maya needs your help
        </Text>
      </Flex>

      <Link
        className="focus-visible:ring-ring mt-2 line-clamp-1 rounded-sm leading-snug font-medium outline-none hover:underline focus-visible:ring-2"
        href={withWorkspace(getStoryPath({ id: issue.storyId }))}
        title={issue.storyTitle}
      >
        {issue.storyTitle}
      </Link>
      <Text
        className="mt-1.5 leading-snug"
        color="muted"
        fontSize="sm"
        title={descriptionTitle}
      >
        {description}
      </Text>
      <Button
        align="center"
        className="mt-3 w-full"
        color="invert"
        onClick={onChooseTime}
        size="sm"
      >
        Choose time
      </Button>
    </>
  );
};

const ObjectiveRiskContent = ({ objective }: { objective: Objective }) => {
  const { getTermDisplay } = useTerminology();
  const { withWorkspace } = useWorkspacePath();
  const copy = getObjectiveForecastRiskCopy(
    objective,
    getTermDisplay("storyTerm", { capitalize: true }),
  );
  if (!copy) return null;

  const objectiveHref = withWorkspace(
    `/teams/${objective.teamId}/objectives/${objective.id}?tab=overview`,
  );

  return (
    <>
      <Flex align="center" className="gap-1.5 pr-6">
        <span className="bg-danger flex size-2 rounded-full" />
        <Text
          className="text-[0.72rem] tracking-[0.08em] uppercase"
          color="muted"
          fontWeight="semibold"
        >
          Forecast risk
        </Text>
        <Text
          className="bg-primary/10 text-primary rounded-md px-1.5 py-0.5 text-[0.78rem] tabular-nums"
          fontWeight="medium"
        >
          +{objective.forecastDaysDelta}d
        </Text>
      </Flex>

      <Link
        className="focus-visible:ring-ring mt-2 line-clamp-1 rounded-sm leading-snug font-medium outline-none hover:underline focus-visible:ring-2"
        href={objectiveHref}
        title={objective.name}
      >
        {objective.name}
      </Link>
      <Text
        className="mt-1.5 line-clamp-2 leading-snug"
        color="muted"
        fontSize="sm"
        title={copy.description}
      >
        {copy.description}
      </Text>
      <Link
        aria-label={`Review objective: ${objective.name}`}
        className="bg-background-inverse text-foreground-inverse focus-visible:ring-ring mt-3 flex h-[2.1rem] w-full items-center justify-center rounded-xl px-2 text-center font-medium transition-opacity hover:opacity-90 focus-visible:ring-2 focus-visible:outline-none"
        href={objectiveHref}
      >
        Review objective
      </Link>
    </>
  );
};

export const SidebarAssistantCards = ({
  fallback = null,
  isCollapsed = false,
}: {
  fallback?: ReactNode;
  isCollapsed?: boolean;
}) => {
  const now = useSyncExternalStore(
    subscribeToClock,
    getClockSnapshot,
    getServerClockSnapshot,
  );
  const [activeIndex, setActiveIndex] = useState(0);
  const [dismissedMeetingIds, setDismissedMeetingIds] = useState<Set<string>>(
    () => new Set(),
  );
  const [dismissedIssueTokens, setDismissedIssueTokens] = useState<
    Map<string, number>
  >(() => new Map());
  const [dismissedObjectiveRiskTokens, setDismissedObjectiveRiskTokens] =
    useState<Set<string>>(() => new Set());
  const [selectedIssue, setSelectedIssue] =
    useState<CalendarScheduleIssue | null>(null);
  const [isCollapsedPopoverOpen, setIsCollapsedPopoverOpen] = useState(false);
  const collapsedPopoverCloseTimer = useRef<number | undefined>(undefined);
  const integrationQuery = useCalendarIntegration();
  const { workspaceSlug } = useWorkspacePath();
  const { data: session } = useSession();
  const { data: objectives = EMPTY_OBJECTIVES } = useObjectives();
  const { data: objectiveStatuses = EMPTY_OBJECTIVE_STATUSES } =
    useObjectiveStatuses();
  const connection = integrationQuery.data?.connections[0];
  const rangeStart = startOfDay(now);
  const scheduleQuery = useCalendarSchedule(
    {
      startAt: rangeStart.toISOString(),
      endAt: addDays(rangeStart, 2).toISOString(),
    },
    { enabled: Boolean(connection) },
  );
  const overrideIssue = useOverrideCalendarScheduleIssue();
  const isHydrated = useSyncExternalStore(
    subscribeToHydration,
    () => true,
    () => false,
  );

  useEffect(() => {
    if (!isHydrated) return;
    objectives.forEach((objective) => {
      if (!isObjectiveForecastAtRisk(objective)) {
        clearObjectiveRiskDismissal(objective.id);
      }
    });
  }, [isHydrated, objectives]);

  useEffect(
    () => () => {
      if (collapsedPopoverCloseTimer.current !== undefined) {
        window.clearTimeout(collapsedPopoverCloseTimer.current);
      }
    },
    [],
  );

  const meetings = isHydrated
    ? getUpcomingMeetings(scheduleQuery.data?.events ?? [], now).filter(
        (meeting) =>
          !dismissedMeetingIds.has(meeting.event.id) &&
          !isMeetingDismissed(meeting.event.id),
      )
    : [];
  const issues = (scheduleQuery.data?.scheduleIssues ?? []).filter(
    (issue) =>
      !isHydrated ||
      !isScheduleIssueDismissed(
        issue,
        workspaceSlug,
        now,
        dismissedIssueTokens,
      ),
  );
  const activeObjectiveStatusIds = new Set(
    objectiveStatuses
      .filter(
        (status) =>
          status.category !== "completed" && status.category !== "cancelled",
      )
      .map((status) => status.id),
  );
  const objectiveRisks = isHydrated
    ? objectives.filter(
        (objective) =>
          objective.leadUser === session?.user.id &&
          activeObjectiveStatusIds.has(objective.statusId) &&
          isObjectiveForecastAtRisk(objective) &&
          !dismissedObjectiveRiskTokens.has(getObjectiveRiskToken(objective)) &&
          !isObjectiveRiskDismissed(objective),
      )
    : [];
  const items: SidebarAssistantItem[] = [
    ...meetings.map((meeting) => ({
      id: `meeting:${meeting.event.id}`,
      kind: "meeting" as const,
      meeting,
    })),
    ...issues.map((issue) => ({
      id: `schedule-issue:${issue.storyId}`,
      issue,
      kind: "schedule-issue" as const,
    })),
    ...objectiveRisks.map((objective) => ({
      id: `objective-risk:${objective.id}`,
      kind: "objective-risk" as const,
      objective,
    })),
  ];
  if (items.length === 0) return fallback;

  const safeActiveIndex = activeIndex % items.length;
  const activeItem = items[safeActiveIndex];

  const cancelCollapsedPopoverClose = () => {
    if (collapsedPopoverCloseTimer.current !== undefined) {
      window.clearTimeout(collapsedPopoverCloseTimer.current);
      collapsedPopoverCloseTimer.current = undefined;
    }
  };
  const openCollapsedPopover = () => {
    cancelCollapsedPopoverClose();
    setIsCollapsedPopoverOpen(true);
  };
  const scheduleCollapsedPopoverClose = () => {
    cancelCollapsedPopoverClose();
    collapsedPopoverCloseTimer.current = window.setTimeout(() => {
      setIsCollapsedPopoverOpen(false);
      collapsedPopoverCloseTimer.current = undefined;
    }, 150);
  };

  const dismissActiveItem = () => {
    if (activeItem.kind === "meeting") {
      dismissMeetingUntilEnd(activeItem.meeting);
      setDismissedMeetingIds((current) => {
        const next = new Set(current);
        next.add(activeItem.meeting.event.id);
        return next;
      });
    } else if (activeItem.kind === "schedule-issue") {
      const dismissedUntil = dismissScheduleIssue(
        activeItem.issue,
        workspaceSlug,
        now,
      );
      setDismissedIssueTokens((current) => {
        const next = new Map(current);
        next.set(
          getScheduleIssueDismissalToken(activeItem.issue),
          dismissedUntil,
        );
        return next;
      });
    } else {
      dismissObjectiveRisk(activeItem.objective);
      setDismissedObjectiveRiskTokens((current) => {
        const next = new Set(current);
        next.add(getObjectiveRiskToken(activeItem.objective));
        return next;
      });
    }
    setActiveIndex(0);
  };

  const dismissLabel = getDismissLabel(activeItem);

  let activeContent: ReactNode;
  if (activeItem.kind === "meeting") {
    activeContent = <MeetingContent meeting={activeItem.meeting} />;
  } else if (activeItem.kind === "schedule-issue") {
    activeContent = (
      <ScheduleIssueContent
        issue={activeItem.issue}
        onChooseTime={() => {
          setSelectedIssue(activeItem.issue);
          setIsCollapsedPopoverOpen(false);
        }}
      />
    );
  } else {
    activeContent = <ObjectiveRiskContent objective={activeItem.objective} />;
  }

  const assistantCard = (
    <Box aria-live="polite" className="relative pb-1.5">
      {items.length > 1 ? (
        <Box
          aria-hidden="true"
          className="border-border bg-surface-elevated absolute inset-x-2 bottom-0 h-4 rounded-b-xl border-[0.5px] opacity-70"
        />
      ) : null}
      <Box className="border-border bg-surface-elevated shadow-shadow relative rounded-xl border-[0.5px] px-3.5 pt-3 pb-3.5 shadow-lg">
        <button
          aria-label={dismissLabel}
          className="text-foreground hover:bg-state-hover focus-visible:ring-ring absolute top-2 right-1.5 z-1 grid size-6 place-items-center rounded-md transition-colors outline-none focus-visible:ring-2"
          onClick={dismissActiveItem}
          type="button"
        >
          <CloseIcon
            aria-hidden="true"
            className="h-3.5 w-auto"
            strokeWidth={2.5}
          />
        </button>

        <Box>{activeContent}</Box>

        {items.length > 1 ? (
          <Flex
            align="center"
            className="border-border mt-3 border-t pt-2"
            justify="between"
          >
            <Text className="tabular-nums" color="muted" fontSize="xs">
              {safeActiveIndex + 1} of {items.length}
            </Text>
            <Flex className="gap-1">
              <button
                aria-label="Previous message"
                className="hover:bg-state-hover focus-visible:ring-ring grid size-6 place-items-center rounded-md outline-none focus-visible:ring-2"
                onClick={() => {
                  setActiveIndex(
                    (safeActiveIndex - 1 + items.length) % items.length,
                  );
                }}
                type="button"
              >
                <ChevronLeftIcon aria-hidden="true" className="h-3.5" />
              </button>
              <button
                aria-label="Next message"
                className="hover:bg-state-hover focus-visible:ring-ring grid size-6 place-items-center rounded-md outline-none focus-visible:ring-2"
                onClick={() => {
                  setActiveIndex((safeActiveIndex + 1) % items.length);
                }}
                type="button"
              >
                <ChevronRightIcon aria-hidden="true" className="h-3.5" />
              </button>
            </Flex>
          </Flex>
        ) : null}
      </Box>
    </Box>
  );

  const isWarningItem = activeItem.kind !== "meeting";
  const CollapsedIndicatorIcon = isWarningItem ? WarningIcon : Video02Icon;
  const collapsedIndicatorLabel = getCollapsedIndicatorLabel(activeItem);
  const collapsedIndicatorClass = isWarningItem
    ? "border-warning/35 bg-warning/10 text-primary hover:bg-warning/15 dark:text-primary"
    : "border-primary/25 bg-primary/10 text-primary hover:bg-primary/15 dark:text-primary";

  return (
    <>
      {isCollapsed ? (
        <Popover
          onOpenChange={setIsCollapsedPopoverOpen}
          open={isCollapsedPopoverOpen}
        >
          <Box
            aria-live="polite"
            className="relative w-full"
            onMouseEnter={openCollapsedPopover}
            onMouseLeave={scheduleCollapsedPopoverClose}
          >
            <Popover.Trigger asChild>
              <button
                aria-expanded={isCollapsedPopoverOpen}
                aria-label={collapsedIndicatorLabel}
                className={`focus-visible:ring-ring relative grid h-10 w-full place-items-center rounded-xl border transition-colors outline-none focus-visible:ring-2 ${collapsedIndicatorClass}`}
                type="button"
              >
                <CollapsedIndicatorIcon
                  aria-hidden="true"
                  className="text-primary dark:text-primary h-5 w-auto"
                  strokeWidth={2.1}
                />
                {items.length > 9 ? (
                  <span
                    aria-hidden="true"
                    className="bg-primary absolute -top-0.5 -right-0.5 size-2.5 rounded-full"
                    data-testid="collapsed-overflow-pin"
                  >
                    <span className="bg-primary absolute inset-0 animate-ping rounded-full opacity-75 motion-reduce:animate-none" />
                  </span>
                ) : null}
                {items.length > 1 && items.length <= 9 ? (
                  <span className="bg-background-inverse text-foreground-inverse absolute -top-1 -right-1 grid min-w-4 place-items-center rounded-full px-1 text-[0.625rem] leading-4 font-semibold tabular-nums">
                    {items.length}
                  </span>
                ) : null}
              </button>
            </Popover.Trigger>
          </Box>
          <Popover.Content
            align="end"
            className="m-0 w-[calc(var(--sidebar-width)-1.75rem)] max-w-[calc(100vw-5rem)] min-w-0 border-0 bg-transparent p-0 shadow-none backdrop-blur-none dark:bg-transparent"
            onCloseAutoFocus={(event) => {
              event.preventDefault();
            }}
            onMouseEnter={openCollapsedPopover}
            onMouseLeave={scheduleCollapsedPopoverClose}
            onOpenAutoFocus={(event) => {
              event.preventDefault();
            }}
            side="right"
            sideOffset={10}
          >
            {assistantCard}
          </Popover.Content>
        </Popover>
      ) : (
        assistantCard
      )}

      {selectedIssue ? (
        <ScheduleIssueDialog
          isSaving={overrideIssue.isPending}
          issue={selectedIssue}
          now={now}
          onClose={() => {
            setSelectedIssue(null);
          }}
          onSubmit={(startAt) => {
            overrideIssue.mutate(
              {
                storyId: selectedIssue.storyId,
                startAt,
                timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
              },
              {
                onSuccess: () => {
                  setSelectedIssue(null);
                },
              },
            );
          }}
        />
      ) : null}
    </>
  );
};

// Retain the existing export for nearby consumers while the component now owns
// the full sidebar assistant stack, not just meeting reminders.
export const UpcomingMeetingCard = SidebarAssistantCards;
