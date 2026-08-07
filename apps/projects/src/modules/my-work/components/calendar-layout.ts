import { startOfDay } from "date-fns";

export type CalendarLayoutEvent = {
  id: string;
  startAt: string;
  endAt: string;
};

export type CalendarEventLayout = {
  id: string;
  top: number;
  height: number;
  lane: number;
  laneCount: number;
};

type CalendarVisibleHoursEvent = {
  startAt: string;
  endAt: string;
  isAllDay?: boolean;
};

export const parseCalendarDate = (value?: string) => {
  const dateParts = value?.split("-") ?? [];
  if (dateParts.length !== 3 || !/^\d{4}-\d{2}-\d{2}$/.test(value ?? "")) {
    return null;
  }

  const year = Number(dateParts[0]);
  const month = Number(dateParts[1]);
  const day = Number(dateParts[2]);
  const date = new Date(year, month - 1, day);
  if (
    date.getFullYear() !== year ||
    date.getMonth() !== month - 1 ||
    date.getDate() !== day
  ) {
    return null;
  }
  return date;
};

// Detailed-event snapshots and availability-only snapshots are mutually
// exclusive. The API still returns busy windows beside detailed events so
// clients that predate event rendering continue to respect occupied time.
export const getDisplayBusyWindows = <T>({
  busyWindows,
  events,
}: {
  busyWindows: readonly T[];
  events: readonly unknown[];
}) => (events.length > 0 ? [] : busyWindows);

export const deriveCalendarVisibleHours = ({
  defaultEndHour,
  defaultStartHour,
  events,
}: {
  defaultEndHour: number;
  defaultStartHour: number;
  events: CalendarVisibleHoursEvent[];
}) => {
  let visibleStartHour = defaultStartHour;
  let visibleEndHour = defaultEndHour;

  for (const event of events) {
    if (event.isAllDay) continue;
    const start = new Date(event.startAt);
    const end = new Date(event.endAt);
    if (
      Number.isNaN(start.getTime()) ||
      Number.isNaN(end.getTime()) ||
      end <= start
    ) {
      continue;
    }

    if (start.toDateString() !== end.toDateString()) {
      visibleStartHour = 0;
      visibleEndHour = 24;
      continue;
    }

    visibleStartHour = Math.min(visibleStartHour, start.getHours());
    visibleEndHour = Math.max(
      visibleEndHour,
      Math.ceil(end.getHours() + end.getMinutes() / 60),
    );
  }

  return {
    visibleStartHour: Math.max(0, visibleStartHour),
    visibleEndHour: Math.min(
      24,
      Math.max(visibleStartHour + 1, visibleEndHour),
    ),
  };
};

type BuildCalendarEventLayoutsInput = {
  day: Date;
  events: CalendarLayoutEvent[];
  hourHeight: number;
  visibleStartHour: number;
  visibleEndHour: number;
};

type TimedLayoutEvent = CalendarLayoutEvent & {
  start: Date;
  end: Date;
  top: number;
  height: number;
  lane: number;
  laneCount: number;
};

export const buildCalendarEventLayouts = ({
  day,
  events,
  hourHeight,
  visibleStartHour,
  visibleEndHour,
}: BuildCalendarEventLayoutsInput): CalendarEventLayout[] => {
  const dayStart = startOfDay(day);
  const visibleStart = atLocalHour(dayStart, visibleStartHour);
  const visibleEnd = atLocalHour(dayStart, visibleEndHour);
  const pixelsPerMinute = hourHeight / 60;

  const timedEvents = events
    .map((event): TimedLayoutEvent | null => {
      const rawStart = new Date(event.startAt);
      const rawEnd = new Date(event.endAt);
      const startsAtBoundary = rawStart <= visibleStart;
      const endsAtBoundary = rawEnd >= visibleEnd;
      const start = startsAtBoundary ? visibleStart : rawStart;
      const end = endsAtBoundary ? visibleEnd : rawEnd;
      if (end <= start) {
        return null;
      }
      const startMinute = startsAtBoundary
        ? visibleStartHour * 60
        : wallClockMinute(start);
      return {
        ...event,
        start,
        end,
        top: Math.max(
          0,
          (startMinute - visibleStartHour * 60) * pixelsPerMinute,
        ),
        height: Math.max(
          24,
          ((end.getTime() - start.getTime()) / 60_000) * pixelsPerMinute,
        ),
        lane: 0,
        laneCount: 1,
      };
    })
    .filter((event): event is TimedLayoutEvent => event !== null)
    .sort((first, second) => first.start.getTime() - second.start.getTime());

  const clusters: TimedLayoutEvent[][] = [];
  let currentCluster: TimedLayoutEvent[] = [];
  let clusterEnd = new Date(0);
  for (const event of timedEvents) {
    if (currentCluster.length === 0 || event.start < clusterEnd) {
      currentCluster.push(event);
      if (event.end > clusterEnd) {
        clusterEnd = event.end;
      }
      continue;
    }
    clusters.push(currentCluster);
    currentCluster = [event];
    clusterEnd = event.end;
  }
  if (currentCluster.length > 0) {
    clusters.push(currentCluster);
  }

  for (const cluster of clusters) {
    const laneEnds: Date[] = [];
    for (const event of cluster) {
      const lane = laneEnds.findIndex((end) => end <= event.start);
      if (lane === -1) {
        event.lane = laneEnds.length;
        laneEnds.push(event.end);
      } else {
        event.lane = lane;
        laneEnds[lane] = event.end;
      }
    }
    for (const event of cluster) {
      event.laneCount = laneEnds.length;
    }
  }

  return timedEvents.map(({ id, top, height, lane, laneCount }) => ({
    id,
    top,
    height,
    lane,
    laneCount,
  }));
};

const atLocalHour = (date: Date, hour: number) => {
  const boundary = new Date(date);
  boundary.setHours(hour, 0, 0, 0);
  return boundary;
};

const wallClockMinute = (date: Date) =>
  date.getHours() * 60 + date.getMinutes();
