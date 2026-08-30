import type {
  CalendarBusyWindow,
  CalendarEventSummary,
  CalendarScheduleBlock,
} from "@/lib/queries/calendar/types";
import type { CalendarDragKind } from "./calendar-drag";

export type CalendarItem =
  | {
      kind: "event";
      id: string;
      startAt: string;
      endAt: string;
      event: CalendarEventSummary;
    }
  | {
      kind: "busy";
      id: string;
      startAt: string;
      endAt: string;
      window: CalendarBusyWindow;
    }
  | {
      kind: "block";
      id: string;
      startAt: string;
      endAt: string;
      block: CalendarScheduleBlock;
    };

export type CalendarDragData = {
  kind: CalendarDragKind;
  block: CalendarScheduleBlock;
};

export type CalendarManualChange = {
  block: CalendarScheduleBlock;
  change: CalendarDragKind;
  startAt: Date;
  endAt: Date;
};

export type CalendarStoryOption = {
  id: string;
  sequenceId: number;
  team?: { code: string } | null;
  title: string;
};
