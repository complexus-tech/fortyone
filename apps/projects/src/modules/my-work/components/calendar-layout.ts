import { differenceInMinutes, startOfDay } from "date-fns";

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
  const visibleStart = addMinutes(dayStart, visibleStartHour * 60);
  const visibleEnd = addMinutes(dayStart, visibleEndHour * 60);
  const pixelsPerMinute = hourHeight / 60;

  const timedEvents = events
    .map((event): TimedLayoutEvent | null => {
      const rawStart = new Date(event.startAt);
      const rawEnd = new Date(event.endAt);
      const start = rawStart < visibleStart ? visibleStart : rawStart;
      const end = rawEnd > visibleEnd ? visibleEnd : rawEnd;
      if (end <= start) {
        return null;
      }
      return {
        ...event,
        start,
        end,
        top: Math.max(
          0,
          differenceInMinutes(start, visibleStart) * pixelsPerMinute,
        ),
        height: Math.max(24, differenceInMinutes(end, start) * pixelsPerMinute),
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

const addMinutes = (date: Date, minutes: number) =>
  new Date(date.getTime() + minutes * 60 * 1000);
