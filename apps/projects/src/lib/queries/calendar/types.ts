import type { AutoSchedulingStatus } from "@/modules/stories/types";

export type CalendarBusyWindow = {
  id: string;
  provider: string;
  title?: string;
  startAt: string;
  endAt: string;
  status: "busy";
  isPrivate: boolean;
  createdAt: string;
  updatedAt: string;
};

export type CalendarEventPerson = {
  displayName?: string;
  email?: string;
};

export type CalendarEventAttendee = CalendarEventPerson & {
  responseStatus?: string;
  optional: boolean;
  organizer: boolean;
  self: boolean;
};

export type CalendarEventSummary = {
  id: string;
  provider: string;
  calendarId?: string;
  title?: string;
  location?: string;
  meetingUrl?: string;
  htmlLink?: string;
  startAt: string;
  endAt: string;
  isAllDay: boolean;
  startDate?: string;
  endDate?: string;
  isPrivate: boolean;
};

export type CalendarEventDetail = CalendarEventSummary & {
  description?: string;
  organizer?: CalendarEventPerson;
  attendees: CalendarEventAttendee[];
  attendeesOmitted: boolean;
};

export type CalendarScheduleBlock = {
  id: string;
  storyId?: string;
  storyTitle?: string;
  storyCode?: string;
  storyStatusColor?: string;
  teamId?: string;
  teamName?: string;
  teamCode?: string;
  blockType: "work" | "focus";
  title: string;
  startAt: string;
  endAt: string;
  hasConflict: boolean;
  isLocked: boolean;
  isCrossWorkspace?: boolean;
  source: "user" | "maya" | "other_workspace";
  autoSchedulingStatus?: AutoSchedulingStatus;
  autoSchedulingReason?: string | null;
  createdAt: string;
  updatedAt: string;
  manualOverrideAt?: string;
  manualOverrideBy?: string;
};

export type CalendarScheduleIssue = {
  storyId: string;
  storyTitle: string;
  storyCode: string;
  teamId: string;
  teamName: string;
  teamCode: string;
  estimatedDurationMinutes: number | null;
  scheduledDurationMinutes?: number;
  remainingDurationMinutes?: number;
  autoSchedulingStatus: "cannot_fit";
  autoSchedulingReason?: string | null;
  updatedAt: string;
};

export type CalendarSchedule = {
  startAt: string;
  endAt: string;
  events: CalendarEventSummary[];
  busyWindows: CalendarBusyWindow[];
  blocks: CalendarScheduleBlock[];
  scheduleIssues: CalendarScheduleIssue[];
};

export type CalendarScheduleBlockInput = {
  storyId?: string | null;
  blockType: "work" | "focus";
  title: string;
  startAt: string;
  endAt: string;
  isLocked?: boolean;
};

export type CalendarManualScheduleBlockInput = {
  startAt: string;
  endAt: string;
  expectedUpdatedAt?: string;
  timezone: string;
  change: "move" | "resize";
  clientMutationId: string;
};
