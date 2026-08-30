"use client";

import { format } from "date-fns";
import { ChevronLeftIcon, ChevronRightIcon } from "icons";
import { Button, Flex, Text } from "ui";
import {
  calculateGanttPosition,
  getGanttOffscreenDirection,
  type ZoomLevel,
} from "./base-gantt-utils";

export const BaseGanttOffscreenIndicator = <
  T extends {
    startDate?: string | null;
    endDate?: string | null;
  },
>({
  dateRange,
  item,
  onScrollToItem,
  viewport,
  zoomLevel,
}: {
  dateRange: { start: Date; end: Date };
  item: T;
  onScrollToItem: (item: T) => void;
  viewport: { scrollLeft: number; visibleWidth: number };
  zoomLevel: ZoomLevel;
}) => {
  if (!item.startDate || !item.endDate) return null;

  const position = calculateGanttPosition({
    start: new Date(item.startDate),
    end: new Date(item.endDate),
    dateRange,
    zoomLevel,
  });
  const direction = getGanttOffscreenDirection(position, viewport);
  if (!direction || viewport.visibleWidth <= 0) return null;

  const indicatorPosition =
    direction === "left"
      ? viewport.scrollLeft + 12
      : viewport.scrollLeft + viewport.visibleWidth - 12;
  const dateRangeLabel = `${format(
    new Date(item.startDate),
    "MMM yyyy",
  )} - ${format(new Date(item.endDate), "MMM yyyy")}`;

  return (
    <Flex
      align="center"
      className="absolute top-3 z-20 gap-2"
      style={{
        left: indicatorPosition,
        transform: direction === "right" ? "translateX(-100%)" : undefined,
      }}
    >
      {direction === "right" ? (
        <Text
          className="shrink-0 text-[0.95rem] whitespace-nowrap dark:opacity-70"
          color="muted"
        >
          {dateRangeLabel}
        </Text>
      ) : null}
      <Button
        aria-label={`Scroll ${direction} to item`}
        asIcon
        className="border-border dark:border-border/70 dark:bg-surface/80 h-8 w-8 bg-white/80 p-0 backdrop-blur-2xl"
        color="tertiary"
        leftIcon={
          direction === "left" ? (
            <ChevronLeftIcon className="h-3.5 w-auto" />
          ) : (
            <ChevronRightIcon className="h-3.5 w-auto" />
          )
        }
        onClick={() => {
          onScrollToItem(item);
        }}
        rounded="full"
        size="sm"
        variant="outline"
      />
      {direction === "left" ? (
        <Text
          className="shrink-0 text-[0.95rem] whitespace-nowrap dark:opacity-70"
          color="muted"
        >
          {dateRangeLabel}
        </Text>
      ) : null}
    </Flex>
  );
};
