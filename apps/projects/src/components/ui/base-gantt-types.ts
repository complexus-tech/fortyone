import type { ReactNode } from "react";
import type { ZoomLevel } from "./base-gantt-utils";
import type { GanttRowWindow } from "./base-gantt-row-window";

export type GanttDateRange = {
  start: Date;
  end: Date;
};

export type GanttViewport = {
  scrollLeft: number;
  visibleWidth: number;
};

export type GanttItem = {
  id: string;
  startDate?: string | null;
  endDate?: string | null;
};

export type BaseGanttProps<T extends GanttItem> = {
  items: T[];
  className?: string;
  storageKey: string;
  zoomLevel?: ZoomLevel;
  controlledZoomLevel?: ZoomLevel;
  onZoomLevelChange?: (zoom: ZoomLevel) => void;
  scrollToTodayRequest?: number;
  stickyColumnsWidth?: number;
  rowHeight?: number | string;
  barClassName?: string;
  virtualizeRows?: boolean;
  pinnedItemIds?: readonly string[];
  onDateUpdate: (itemId: string, startDate: string, endDate: string) => void;
  onBarClick?: (item: T) => void;
  renderSidebar: (
    items: T[],
    onReset: () => void,
    zoomLevel: ZoomLevel,
    onZoomChange: (zoom: ZoomLevel) => void,
    rowWindow: GanttRowWindow,
  ) => ReactNode;
  renderBarContent: (item: T) => ReactNode;
};
