"use client";

import { Box, Button, Flex, Menu, Text, Tooltip, Avatar } from "ui";
import { cn } from "lib";
import {
  format,
  eachDayOfInterval,
  addDays,
  differenceInDays,
  formatISO,
  isWeekend,
  subDays,
  getWeek,
  isSameWeek,
  isYesterday,
  eachMonthOfInterval,
  eachQuarterOfInterval,
  startOfMonth,
  endOfMonth,
  startOfQuarter,
  endOfQuarter,
  addMonths,
  addQuarters,
} from "date-fns";
import { useState, useCallback, useEffect, useRef, useMemo } from "react";
import { ArrowDown2Icon } from "icons";
import Link from "next/link";
import { useQueryClient } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import type { Objective } from "@/modules/objectives/types";
import { useUpdateObjectiveMutation } from "@/modules/objectives/hooks/update-mutation";
import { useTeamMembers } from "@/lib/hooks/team-members";
import { useLocalStorage, useUserRole } from "@/hooks";
import { objectiveKeys } from "@/modules/objectives/constants";
import { getObjective } from "@/modules/objectives/queries/get-objective";
import { PrioritiesMenu } from "@/components/ui/story/priorities-menu";
import { ObjectiveStatusesMenu } from "@/components/ui/objective-statuses-menu";
import { AssigneesMenu } from "@/components/ui/story/assignees-menu";
import { PriorityIcon } from "./priority-icon";
import { ObjectiveStatusIcon } from "./objective-status-icon";

type ZoomLevel = "weeks" | "months" | "quarters";

const getTimePeriodsForZoom = (
  dateRange: { start: Date; end: Date },
  zoomLevel: ZoomLevel,
) => {
  switch (zoomLevel) {
    case "weeks":
      return eachDayOfInterval({
        start: dateRange.start,
        end: dateRange.end,
      });
    case "months":
      return eachMonthOfInterval({
        start: startOfMonth(dateRange.start),
        end: endOfMonth(dateRange.end),
      });
    case "quarters":
      return eachQuarterOfInterval({
        start: startOfQuarter(dateRange.start),
        end: endOfQuarter(dateRange.end),
      });
    default:
      return [];
  }
};

const getColumnWidth = (zoomLevel: ZoomLevel) => {
  switch (zoomLevel) {
    case "weeks":
      return 64;
    case "months":
      return 120;
    case "quarters":
      return 180;
    default:
      return 64;
  }
};

const getWeekSpans = (days: Date[]) => {
  if (days.length === 0) return [];

  const spans: {
    week: string;
    month: string;
    startIndex: number;
    span: number;
  }[] = [];
  let startIndex = 0;

  for (let i = 0; i < days.length; i++) {
    const currentDay = days[i];
    const nextDay = days[i + 1];

    const isEndOfWeek =
      i === days.length - 1 ||
      !isSameWeek(currentDay, nextDay, { weekStartsOn: 0 });

    if (isEndOfWeek) {
      const span = i - startIndex + 1;
      const weekStart = days[startIndex];
      const weekNumber = getWeek(weekStart, { weekStartsOn: 0 });
      const monthYear = format(weekStart, "MMM yyyy");

      spans.push({
        week: `Week ${weekNumber}`,
        month: monthYear,
        startIndex,
        span,
      });

      startIndex = i + 1;
    }
  }

  return spans;
};

const getVisibleDateRange = (centerDate: Date, viewportDays = 365) => {
  const halfViewport = Math.floor(viewportDays / 2);
  const start = subDays(centerDate, halfViewport);
  const end = addDays(centerDate, halfViewport);

  start.setHours(0, 0, 0, 0);
  end.setHours(0, 0, 0, 0);

  return { start, end };
};

const Bar = ({
  objective,
  dateRange,
  onDateUpdate,
  containerRef,
  zoomLevel,
}: {
  objective: Objective;
  dateRange: { start: Date; end: Date };
  onDateUpdate: (
    objectiveId: string,
    startDate: string,
    endDate: string,
  ) => void;
  containerRef: React.RefObject<HTMLDivElement | null>;
  zoomLevel: ZoomLevel;
}) => {
  const [isDragging, setIsDragging] = useState(false);
  const [dragStart, setDragStart] = useState<{
    x: number;
    type: "move" | "resize-start" | "resize-end";
    originalStartDate: Date;
    originalEndDate: Date;
  } | null>(null);

  const [dragPosition, setDragPosition] = useState<{
    leftPosition: number;
    width: number;
  } | null>(null);

  const startDate = useMemo(() => {
    const date = objective.startDate
      ? new Date(objective.startDate)
      : new Date();
    date.setHours(0, 0, 0, 0);
    return date;
  }, [objective.startDate]);

  const endDate = useMemo(() => {
    const date = objective.endDate
      ? new Date(objective.endDate)
      : addDays(startDate, 1);
    date.setHours(0, 0, 0, 0);
    return date;
  }, [objective.endDate, startDate]);

  const columnWidth = getColumnWidth(zoomLevel);

  const handleMouseDown = useCallback(
    (e: React.MouseEvent, type: "move" | "resize-start" | "resize-end") => {
      e.preventDefault();
      e.stopPropagation();
      setIsDragging(true);
      setDragStart({
        x: e.clientX,
        type,
        originalStartDate: startDate,
        originalEndDate: endDate,
      });
    },
    [startDate, endDate],
  );

  const handleMouseMove = useCallback(
    (e: MouseEvent) => {
      if (!isDragging || !dragStart) return;

      const deltaX = e.clientX - dragStart.x;
      const container = containerRef.current;
      if (!container) return;

      const deltaUnits = Math.round(deltaX / columnWidth);

      if (deltaUnits === 0) return;

      let newStartDate = dragStart.originalStartDate;
      let newEndDate = dragStart.originalEndDate;

      if (dragStart.type === "move") {
        switch (zoomLevel) {
          case "weeks":
            newStartDate = addDays(dragStart.originalStartDate, deltaUnits);
            newEndDate = addDays(dragStart.originalEndDate, deltaUnits);
            break;
          case "months":
            newStartDate = addMonths(dragStart.originalStartDate, deltaUnits);
            newEndDate = addMonths(dragStart.originalEndDate, deltaUnits);
            break;
          case "quarters":
            newStartDate = addQuarters(dragStart.originalStartDate, deltaUnits);
            newEndDate = addQuarters(dragStart.originalEndDate, deltaUnits);
            break;
        }
      } else if (dragStart.type === "resize-start") {
        switch (zoomLevel) {
          case "weeks":
            newStartDate = addDays(dragStart.originalStartDate, deltaUnits);
            if (newStartDate >= dragStart.originalEndDate) {
              newStartDate = addDays(dragStart.originalEndDate, -1);
            }
            break;
          case "months":
            newStartDate = addMonths(dragStart.originalStartDate, deltaUnits);
            if (newStartDate >= dragStart.originalEndDate) {
              newStartDate = addMonths(dragStart.originalEndDate, -1);
            }
            break;
          case "quarters":
            newStartDate = addQuarters(dragStart.originalStartDate, deltaUnits);
            if (newStartDate >= dragStart.originalEndDate) {
              newStartDate = addQuarters(dragStart.originalEndDate, -1);
            }
            break;
        }
      } else {
        switch (zoomLevel) {
          case "weeks":
            newEndDate = addDays(dragStart.originalEndDate, deltaUnits);
            if (newEndDate <= dragStart.originalStartDate) {
              newEndDate = addDays(dragStart.originalStartDate, 1);
            }
            break;
          case "months":
            newEndDate = addMonths(dragStart.originalEndDate, deltaUnits);
            if (newEndDate <= dragStart.originalStartDate) {
              newEndDate = addMonths(dragStart.originalStartDate, 1);
            }
            break;
          case "quarters":
            newEndDate = addQuarters(dragStart.originalEndDate, deltaUnits);
            if (newEndDate <= dragStart.originalStartDate) {
              newEndDate = addQuarters(dragStart.originalStartDate, 1);
            }
            break;
        }
      }

      let newLeftPosition = 0;
      let newWidth = columnWidth;

      switch (zoomLevel) {
        case "weeks": {
          const newStartOffset = differenceInDays(
            newStartDate,
            dateRange.start,
          );
          const newDuration = differenceInDays(newEndDate, newStartDate) || 1;
          newLeftPosition = newStartOffset * columnWidth;
          newWidth = newDuration * columnWidth;
          break;
        }
        case "months":
        case "quarters": {
          const periods = getTimePeriodsForZoom(dateRange, zoomLevel);

          if (zoomLevel === "months") {
            const newStartMonth = startOfMonth(newStartDate);
            const newEndMonth = startOfMonth(newEndDate);

            const startOffset = periods.findIndex(
              (period) => period.getTime() === newStartMonth.getTime(),
            );
            const endOffset = periods.findIndex(
              (period) => period.getTime() === newEndMonth.getTime(),
            );

            if (startOffset !== -1) {
              newLeftPosition = startOffset * columnWidth;
              const duration =
                endOffset !== -1 ? Math.max(1, endOffset - startOffset + 1) : 1;
              newWidth = duration * columnWidth;
            }
          } else {
            const newStartQuarter = startOfQuarter(newStartDate);
            const newEndQuarter = startOfQuarter(newEndDate);

            const startOffset = periods.findIndex(
              (period) => period.getTime() === newStartQuarter.getTime(),
            );
            const endOffset = periods.findIndex(
              (period) => period.getTime() === newEndQuarter.getTime(),
            );

            if (startOffset !== -1) {
              newLeftPosition = startOffset * columnWidth;
              const duration =
                endOffset !== -1 ? Math.max(1, endOffset - startOffset + 1) : 1;
              newWidth = duration * columnWidth;
            }
          }
          break;
        }
      }

      setDragPosition({
        leftPosition: newLeftPosition,
        width: newWidth,
      });
    },
    [isDragging, dragStart, columnWidth, dateRange, zoomLevel, containerRef],
  );

  const handleMouseUp = useCallback(() => {
    if (!isDragging || !dragStart || !dragPosition) {
      setIsDragging(false);
      setDragStart(null);
      setDragPosition(null);
      return;
    }

    let finalStartDate = dragStart.originalStartDate;
    let finalEndDate = dragStart.originalEndDate;

    switch (zoomLevel) {
      case "weeks": {
        const finalStartOffset = Math.round(
          dragPosition.leftPosition / columnWidth,
        );
        const finalDuration = Math.round(dragPosition.width / columnWidth);

        finalStartDate = addDays(dateRange.start, finalStartOffset);
        finalEndDate = addDays(finalStartDate, Math.max(1, finalDuration));
        break;
      }
      case "months":
      case "quarters": {
        const periods = getTimePeriodsForZoom(dateRange, zoomLevel);
        const startPeriodIndex = Math.round(
          dragPosition.leftPosition / columnWidth,
        );
        const durationPeriods = Math.round(dragPosition.width / columnWidth);

        if (startPeriodIndex >= 0 && startPeriodIndex < periods.length) {
          const startPeriod = periods[startPeriodIndex];

          if (zoomLevel === "months") {
            finalStartDate = startOfMonth(startPeriod);
            if (durationPeriods === 1) {
              finalEndDate = endOfMonth(startPeriod);
            } else {
              const lastMonth = addMonths(startPeriod, durationPeriods - 1);
              finalEndDate = endOfMonth(lastMonth);
            }
          } else {
            finalStartDate = startOfQuarter(startPeriod);
            if (durationPeriods === 1) {
              finalEndDate = endOfQuarter(startPeriod);
            } else {
              const lastQuarter = addQuarters(startPeriod, durationPeriods - 1);
              finalEndDate = endOfQuarter(lastQuarter);
            }
          }
        }
        break;
      }
    }

    const originalStartISO = formatISO(dragStart.originalStartDate);
    const originalEndISO = formatISO(dragStart.originalEndDate);
    const finalStartISO = formatISO(finalStartDate);
    const finalEndISO = formatISO(finalEndDate);

    if (originalStartISO !== finalStartISO || originalEndISO !== finalEndISO) {
      onDateUpdate(objective.id, finalStartISO, finalEndISO);
    }

    setIsDragging(false);
    setDragStart(null);

    setTimeout(() => {
      setDragPosition(null);
    }, 200);
  }, [
    isDragging,
    dragStart,
    dragPosition,
    columnWidth,
    dateRange,
    objective.id,
    onDateUpdate,
    zoomLevel,
  ]);

  useEffect(() => {
    if (isDragging) {
      document.addEventListener("mousemove", handleMouseMove);
      document.addEventListener("mouseup", handleMouseUp);
      return () => {
        document.removeEventListener("mousemove", handleMouseMove);
        document.removeEventListener("mouseup", handleMouseUp);
      };
    }
  }, [isDragging, handleMouseMove, handleMouseUp]);

  const calculatePosition = () => {
    const periods = getTimePeriodsForZoom(dateRange, zoomLevel);
    const columnWidth = getColumnWidth(zoomLevel);

    let startOffset = 0;
    let duration = 1;

    switch (zoomLevel) {
      case "weeks": {
        startOffset = differenceInDays(startDate, dateRange.start);
        duration = differenceInDays(endDate, startDate) || 1;
        break;
      }
      case "months": {
        const startMonth = startOfMonth(startDate);
        const endMonth = startOfMonth(endDate);

        startOffset = periods.findIndex(
          (period) => period.getTime() === startMonth.getTime(),
        );

        const endOffset = periods.findIndex(
          (period) => period.getTime() === endMonth.getTime(),
        );

        if (startOffset === -1) startOffset = 0;
        if (endOffset === -1) {
          duration = 1;
        } else {
          duration = Math.max(1, endOffset - startOffset + 1);
        }
        break;
      }
      case "quarters": {
        const startQuarter = startOfQuarter(startDate);
        const endQuarter = startOfQuarter(endDate);

        startOffset = periods.findIndex(
          (period) => period.getTime() === startQuarter.getTime(),
        );

        const endOffset = periods.findIndex(
          (period) => period.getTime() === endQuarter.getTime(),
        );

        if (startOffset === -1) startOffset = 0;
        if (endOffset === -1) {
          duration = 1;
        } else {
          duration = Math.max(1, endOffset - startOffset + 1);
        }
        break;
      }
    }

    return {
      leftPosition: startOffset * columnWidth,
      width: duration * columnWidth,
    };
  };

  const { leftPosition: calculatedLeft, width: calculatedWidth } =
    calculatePosition();

  const finalLeftPosition = dragPosition?.leftPosition ?? calculatedLeft;
  const finalWidth = dragPosition?.width ?? calculatedWidth;

  if (!objective.startDate || !objective.endDate) return null;
  if (finalWidth <= 0) return null;

  return (
    <Box
      className={cn(
        "group absolute z-0 h-10 cursor-pointer overflow-hidden rounded-lg border-[0.5px] border-gray-200/60 bg-gray-100 transition-colors dark:border-dark-50/80 dark:bg-dark-200",
        {
          "shadow-lg ring-2": isDragging,
        },
      )}
      onFocus={(e) => {
        e.preventDefault();
        e.currentTarget.blur();
      }}
      onKeyDown={(e) => {
        e.preventDefault();
      }}
      onMouseDown={(e) => {
        handleMouseDown(e, "move");
      }}
      style={{
        left: `${finalLeftPosition}px`,
        width: `${finalWidth}px`,
        top: "6px",
      }}
      tabIndex={-1}
    >
      <Box
        className="absolute left-0 top-0 h-full w-1 cursor-ew-resize"
        onMouseDown={(e) => {
          e.stopPropagation();
          handleMouseDown(e, "resize-start");
        }}
      />

      <Box
        className="absolute right-0 top-0 h-full w-1 cursor-ew-resize"
        onMouseDown={(e) => {
          e.stopPropagation();
          handleMouseDown(e, "resize-end");
        }}
      />

      <Text
        className="line-clamp-1 block truncate px-3 leading-10"
        fontWeight="medium"
      >
        {objective.name}
      </Text>
    </Box>
  );
};

const TimelineHeader = ({
  dateRange,
  zoomLevel,
}: {
  dateRange: { start: Date; end: Date };
  zoomLevel: ZoomLevel;
}) => {
  const periods = getTimePeriodsForZoom(dateRange, zoomLevel);
  const columnWidth = getColumnWidth(zoomLevel);

  const timelineMinWidth = periods.length * columnWidth;

  const today = new Date();
  today.setHours(0, 0, 0, 0);

  const renderPeriodHeader = () => {
    switch (zoomLevel) {
      case "weeks":
        return (
          <>
            <Box className="border-b-[0.5px] border-gray-100 dark:border-dark-100">
              <Flex>
                {getWeekSpans(periods).map(({ week, month, span }, index) => (
                  <Box
                    className="border-r-[0.5px] border-gray-100 px-2 py-1.5 text-left dark:border-dark-100"
                    key={`${month}-${week}-${index}`}
                    style={{ width: `${(span / periods.length) * 100}%` }}
                  >
                    <Flex
                      align="center"
                      className="h-5 min-h-0"
                      justify="between"
                    >
                      <Text
                        className="text-[0.9rem]"
                        color="muted"
                        fontWeight="semibold"
                      >
                        {month}
                      </Text>
                      <Text
                        className="text-[0.9rem] opacity-60"
                        color="muted"
                        fontWeight="semibold"
                      >
                        {week}
                      </Text>
                    </Flex>
                  </Box>
                ))}
              </Flex>
            </Box>

            <Flex>
              {periods.map((day) => {
                const isToday = day.getTime() === today.getTime();

                return (
                  <Box
                    className={cn(
                      "h-[calc(2rem-1px)] min-w-16 flex-1 border-r-[0.5px] border-gray-100 px-1 py-1 text-center dark:border-dark-100",
                      {
                        "bg-gray-50 dark:bg-dark-200/30":
                          isWeekend(day) && !isToday,
                        "border-primary bg-primary dark:border-primary":
                          isToday,
                      },
                    )}
                    key={day.getTime()}
                    style={{ minWidth: `${columnWidth}px` }}
                  >
                    <Flex align="center" className="px-1" justify="between">
                      {isToday ? (
                        <Text color="white" fontSize="sm" fontWeight="medium">
                          Today
                        </Text>
                      ) : (
                        <>
                          <Text color="muted" fontSize="sm">
                            {format(day, "d")}
                          </Text>
                          <Text color="muted" fontSize="sm">
                            {format(day, "eeeee")}
                          </Text>
                        </>
                      )}
                    </Flex>
                  </Box>
                );
              })}
            </Flex>
          </>
        );
      case "months":
        return (
          <>
            <Box className="border-b-[0.5px] border-gray-100 dark:border-dark-100">
              <Flex>
                {periods.map((month) => (
                  <Box
                    className="border-r-[0.5px] border-gray-100 px-2 py-1.5 text-left dark:border-dark-100"
                    key={month.getTime()}
                    style={{ minWidth: `${columnWidth}px` }}
                  >
                    <Flex
                      align="center"
                      className="h-5 min-h-0"
                      justify="between"
                    >
                      <Text
                        className="text-[0.9rem]"
                        color="muted"
                        fontWeight="semibold"
                      >
                        {format(month, "MMM")}
                      </Text>
                      <Text
                        className="text-[0.9rem] opacity-60"
                        color="muted"
                        fontWeight="semibold"
                      >
                        {format(month, "yyyy")}
                      </Text>
                    </Flex>
                  </Box>
                ))}
              </Flex>
            </Box>

            <Flex>
              {periods.map((month) => (
                <Box
                  className="h-[calc(2rem-1px)] min-w-16 flex-1 border-r-[0.5px] border-gray-100 px-1 py-1 text-center dark:border-dark-100"
                  key={month.getTime()}
                  style={{ minWidth: `${columnWidth}px` }}
                >
                  <Flex align="center" className="px-1" justify="between">
                    <Text color="muted" fontSize="sm">
                      {format(month, "d")}
                    </Text>
                    <Text color="muted" fontSize="sm">
                      {format(endOfMonth(month), "d")}
                    </Text>
                  </Flex>
                </Box>
              ))}
            </Flex>
          </>
        );
      case "quarters":
        return (
          <>
            <Box className="border-b-[0.5px] border-gray-100 dark:border-dark-100">
              <Flex>
                {periods.map((quarter) => (
                  <Box
                    className="border-r-[0.5px] border-gray-100 px-2 py-1.5 text-left dark:border-dark-100"
                    key={quarter.getTime()}
                    style={{ minWidth: `${columnWidth}px` }}
                  >
                    <Flex
                      align="center"
                      className="h-5 min-h-0"
                      justify="between"
                    >
                      <Text
                        className="text-[0.9rem]"
                        color="muted"
                        fontWeight="semibold"
                      >
                        Q{Math.ceil((quarter.getMonth() + 1) / 3)}
                      </Text>
                      <Text
                        className="text-[0.9rem] opacity-60"
                        color="muted"
                        fontWeight="semibold"
                      >
                        {format(quarter, "yyyy")}
                      </Text>
                    </Flex>
                  </Box>
                ))}
              </Flex>
            </Box>

            <Flex>
              {periods.map((quarter) => {
                const quarterStart = startOfQuarter(quarter);
                const quarterEnd = endOfQuarter(quarter);

                return (
                  <Box
                    className="h-[calc(2rem-1px)] min-w-16 flex-1 border-r-[0.5px] border-gray-100 px-1 py-1 text-center dark:border-dark-100"
                    key={quarter.getTime()}
                    style={{ minWidth: `${columnWidth}px` }}
                  >
                    <Flex align="center" className="px-1" justify="between">
                      <Text color="muted" fontSize="sm">
                        {format(quarterStart, "MMM")}
                      </Text>
                      <Text color="muted" fontSize="sm">
                        {format(quarterEnd, "MMM")}
                      </Text>
                    </Flex>
                  </Box>
                );
              })}
            </Flex>
          </>
        );
      default:
        return null;
    }
  };

  return (
    <Box
      className="sticky top-0 z-10 h-16 border-b-[0.5px] border-gray-200/60 bg-white dark:border-dark-100 dark:bg-dark"
      style={{ minWidth: `${timelineMinWidth}px` }}
    >
      <Box className="h-8 w-full">{renderPeriodHeader()}</Box>
    </Box>
  );
};

const Header = ({
  onReset,
  zoomLevel,
  onZoomChange,
}: {
  onReset: () => void;
  zoomLevel: ZoomLevel;
  onZoomChange: (zoom: ZoomLevel) => void;
}) => {
  const getZoomLabel = (zoom: ZoomLevel) => {
    switch (zoom) {
      case "weeks":
        return "Weeks";
      case "months":
        return "Months";
      case "quarters":
        return "Quarters";
      default:
        return "Weeks";
    }
  };

  return (
    <Flex
      align="center"
      className="sticky top-0 z-10 h-16 border-b-[0.5px] border-gray-200/60 bg-white px-6 py-2.5 dark:border-dark-100 dark:bg-dark"
      justify="between"
    >
      <Button color="tertiary" onClick={onReset} size="sm">
        Today
      </Button>
      <Flex align="center" gap={2}>
        <Text color="muted" fontWeight="medium">
          Zoom:
        </Text>
        <Menu>
          <Menu.Button>
            <Button
              color="tertiary"
              rightIcon={<ArrowDown2Icon className="h-4" />}
              size="sm"
            >
              {getZoomLabel(zoomLevel)}
            </Button>
          </Menu.Button>
          <Menu.Items className="w-40">
            <Menu.Group>
              <Menu.Item
                onSelect={() => {
                  onZoomChange("weeks");
                }}
              >
                Weeks
              </Menu.Item>
              <Menu.Item
                onSelect={() => {
                  onZoomChange("months");
                }}
              >
                Months
              </Menu.Item>
              <Menu.Item
                onSelect={() => {
                  onZoomChange("quarters");
                }}
              >
                Quarters
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>
    </Flex>
  );
};

const Objectives = ({
  objectives,
  onReset,
  zoomLevel,
  onZoomChange,
}: {
  objectives: Objective[];
  onReset: () => void;
  zoomLevel: ZoomLevel;
  onZoomChange: (zoom: ZoomLevel) => void;
}) => {
  const { data: session } = useSession();
  const queryClient = useQueryClient();
  const { userRole } = useUserRole();
  const { mutate } = useUpdateObjectiveMutation();

  const firstObjectiveTeamId = objectives[0]?.teamId;
  const { data: members = [] } = useTeamMembers(firstObjectiveTeamId);

  const handleUpdate = useCallback(
    (objectiveId: string, data: Partial<Objective>) => {
      mutate({
        objectiveId,
        data,
      });
    },
    [mutate],
  );

  return (
    <Box className="sticky left-0 z-20 min-h-full w-[34rem] shrink-0 border-r-[0.5px] border-gray-200/60 bg-white dark:border-dark-100 dark:bg-dark">
      <Header
        onReset={onReset}
        onZoomChange={(zoom: ZoomLevel) => {
          onZoomChange(zoom);
        }}
        zoomLevel={zoomLevel}
      />
      {objectives.map((objective) => {
        const startDate = objective.startDate
          ? new Date(objective.startDate)
          : null;
        const endDate = objective.endDate ? new Date(objective.endDate) : null;
        const duration =
          startDate && endDate ? differenceInDays(endDate, startDate) : null;

        const selectedLead = members.find(
          (member) => member.id === objective.leadUser,
        );

        return (
          <Box
            key={objective.id}
            onMouseEnter={() => {
              if (session) {
                queryClient.prefetchQuery({
                  queryKey: objectiveKeys.objective(objective.id),
                  queryFn: () => getObjective(objective.id, session),
                });
              }
            }}
          >
            <Flex
              align="center"
              className="group h-14 border-b-[0.5px] border-gray-100 px-6 transition-colors hover:bg-gray-50 dark:border-dark-100 dark:hover:bg-dark-300"
              justify="between"
            >
              <Flex align="center" className="min-w-0 flex-1 gap-3">
                <AssigneesMenu>
                  <Tooltip
                    className="py-2.5"
                    title={
                      <Box>
                        <Flex gap={2}>
                          <Avatar
                            className="mt-0.5"
                            name={selectedLead?.fullName}
                            size="sm"
                            src={selectedLead?.avatarUrl}
                          />
                          <Box>
                            <Text fontSize="md" fontWeight="medium">
                              {selectedLead?.fullName}
                            </Text>
                            <Text color="muted" fontSize="md">
                              ({selectedLead?.username})
                            </Text>
                          </Box>
                        </Flex>
                      </Box>
                    }
                  >
                    <span>
                      <AssigneesMenu.Trigger>
                        <button
                          className="flex"
                          disabled={userRole === "guest"}
                          type="button"
                        >
                          <Avatar
                            name={
                              selectedLead?.fullName || selectedLead?.username
                            }
                            size="xs"
                            src={selectedLead?.avatarUrl}
                          />
                        </button>
                      </AssigneesMenu.Trigger>
                    </span>
                  </Tooltip>
                  <AssigneesMenu.Items
                    assigneeId={selectedLead?.id}
                    onAssigneeSelected={(assigneeId) => {
                      handleUpdate(objective.id, {
                        leadUser: assigneeId || undefined,
                      });
                    }}
                    teamId={objective.teamId}
                  />
                </AssigneesMenu>

                {objective.priority ? (
                  <PrioritiesMenu>
                    <PrioritiesMenu.Trigger>
                      <button
                        className="flex shrink-0 select-none items-center gap-1 disabled:cursor-not-allowed disabled:opacity-50"
                        disabled={userRole === "guest"}
                        type="button"
                      >
                        <PriorityIcon priority={objective.priority} />
                        <span className="sr-only">{objective.priority}</span>
                      </button>
                    </PrioritiesMenu.Trigger>
                    <PrioritiesMenu.Items
                      priority={objective.priority}
                      setPriority={(priority) => {
                        handleUpdate(objective.id, { priority });
                      }}
                    />
                  </PrioritiesMenu>
                ) : null}

                <ObjectiveStatusesMenu>
                  <ObjectiveStatusesMenu.Trigger>
                    <button
                      className="flex items-center gap-1 disabled:cursor-not-allowed disabled:opacity-50"
                      disabled={userRole === "guest"}
                      type="button"
                    >
                      <ObjectiveStatusIcon statusId={objective.statusId} />
                      <span className="sr-only">Objective status</span>
                    </button>
                  </ObjectiveStatusesMenu.Trigger>
                  <ObjectiveStatusesMenu.Items
                    setStatusId={(statusId) => {
                      handleUpdate(objective.id, { statusId });
                    }}
                    statusId={objective.statusId}
                  />
                </ObjectiveStatusesMenu>

                <Link
                  className="flex min-w-0 flex-1 items-center gap-1.5"
                  href={`/teams/${objective.teamId}/objectives/${objective.id}`}
                >
                  <Text
                    className="line-clamp-1 hover:opacity-90"
                    fontWeight="medium"
                  >
                    {objective.name}
                  </Text>
                </Link>
              </Flex>

              {duration ? (
                <Text className="ml-4 shrink-0" color="muted">
                  {duration} day{duration !== 1 ? "s" : ""}
                </Text>
              ) : null}
            </Flex>
          </Box>
        );
      })}
    </Box>
  );
};

const Chart = ({
  objectives,
  dateRange,
  onDateUpdate,
  containerRef,
  zoomLevel,
}: {
  objectives: Objective[];
  dateRange: { start: Date; end: Date };
  onDateUpdate: (
    objectiveId: string,
    startDate: string,
    endDate: string,
  ) => void;
  containerRef: React.RefObject<HTMLDivElement | null>;
  zoomLevel: ZoomLevel;
}) => {
  const periods = getTimePeriodsForZoom(dateRange, zoomLevel);
  const columnWidth = getColumnWidth(zoomLevel);
  const isContainerScrollable =
    containerRef.current &&
    containerRef.current.scrollHeight > containerRef.current.clientHeight;

  const timelineMinWidth = periods.length * columnWidth;

  const today = new Date();
  today.setHours(0, 0, 0, 0);

  return (
    <Box className="flex-1" style={{ minWidth: `${timelineMinWidth}px` }}>
      <TimelineHeader dateRange={dateRange} zoomLevel={zoomLevel} />
      {objectives.map((objective, idx) => (
        <Box className="relative h-14" key={objective.id}>
          <Flex className="absolute inset-0">
            {periods.map((period) => {
              let isToday = false;
              let dayIsYesterday = false;

              if (zoomLevel === "weeks") {
                isToday = period.getTime() === today.getTime();
                dayIsYesterday = isYesterday(period);
              }

              return (
                <Box
                  className={cn(
                    "min-w-16 flex-1 border-r-[0.5px] border-gray-100 dark:border-dark-100",
                    {
                      "bg-gray-50/50 dark:bg-dark-200/20":
                        zoomLevel === "weeks" && isWeekend(period) && !isToday,
                      "border-primary/50 bg-primary/10 dark:border-primary/50":
                        isToday,
                      "border-primary/50 dark:border-primary/50":
                        dayIsYesterday,
                    },
                  )}
                  key={period.getTime()}
                  style={{
                    minWidth: `${columnWidth}px`,
                    height:
                      idx === objectives.length - 1 && !isContainerScrollable
                        ? `calc(100dvh - 8rem - ${3.5 * (objectives.length - 1)}rem)`
                        : "100%",
                  }}
                />
              );
            })}
          </Flex>
          <Box className="z-5 relative h-full px-2">
            <Bar
              containerRef={containerRef}
              dateRange={dateRange}
              objective={objective}
              onDateUpdate={onDateUpdate}
              zoomLevel={zoomLevel}
            />
          </Box>
        </Box>
      ))}
    </Box>
  );
};

type RoadmapGanttBoardProps = {
  objectives: Objective[];
  className?: string;
};

export const RoadmapGanttBoard = ({
  objectives,
  className,
}: RoadmapGanttBoardProps) => {
  const { mutate } = useUpdateObjectiveMutation();
  const containerRef = useRef<HTMLDivElement>(null);
  const hasScrolledRef = useRef(false);
  const [zoomLevel, setZoomLevel] = useLocalStorage<ZoomLevel>(
    "roadmapZoomLevel",
    "months",
  );

  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const dateRange = getVisibleDateRange(today, 365);

  const handleDateUpdate = useCallback(
    (objectiveId: string, startDate: string, endDate: string) => {
      mutate({
        objectiveId,
        data: {
          startDate,
          endDate,
        },
      });
    },
    [mutate],
  );

  const scrollToToday = useCallback(() => {
    const container = containerRef.current;
    if (!container) return;

    requestAnimationFrame(() => {
      const now = new Date();
      const periods = getTimePeriodsForZoom(dateRange, zoomLevel);
      const columnWidth = getColumnWidth(zoomLevel);

      let periodOffset = 0;
      switch (zoomLevel) {
        case "weeks":
          periodOffset = differenceInDays(now, dateRange.start);
          break;
        case "months": {
          const currentMonth = periods.findIndex(
            (period) =>
              period.getMonth() === now.getMonth() &&
              period.getFullYear() === now.getFullYear(),
          );
          periodOffset = currentMonth >= 0 ? currentMonth : 0;
          break;
        }
        case "quarters": {
          const currentQuarter = Math.floor(now.getMonth() / 3);
          const currentQuarterIndex = periods.findIndex(
            (period) =>
              Math.floor(period.getMonth() / 3) === currentQuarter &&
              period.getFullYear() === now.getFullYear(),
          );
          periodOffset = currentQuarterIndex >= 0 ? currentQuarterIndex : 0;
          break;
        }
      }

      const stickyColumnsWidth = 34 * 16;
      const currentPeriodPixelPosition = periodOffset * columnWidth;

      const viewportWidth = container.clientWidth;
      const scrollPosition =
        stickyColumnsWidth + currentPeriodPixelPosition - viewportWidth / 2;

      container.scrollLeft = Math.max(0, scrollPosition);
    });
  }, [dateRange, zoomLevel]);

  useEffect(() => {
    if (!hasScrolledRef.current) {
      scrollToToday();
      hasScrolledRef.current = true;
    }
  }, [scrollToToday]);

  const objectivesWithDates = objectives.filter(
    (objective) => objective.startDate || objective.endDate,
  );

  return (
    <div
      className={cn(
        "relative left-px overflow-x-auto overflow-y-auto",
        className,
      )}
      ref={containerRef}
    >
      <Flex className="min-h-[calc(100dvh-4rem)] min-w-max">
        <Objectives
          objectives={objectivesWithDates}
          onReset={scrollToToday}
          onZoomChange={(zoom: ZoomLevel) => {
            setZoomLevel(zoom);
          }}
          zoomLevel={zoomLevel}
        />
        <Chart
          containerRef={containerRef}
          dateRange={dateRange}
          objectives={objectivesWithDates}
          onDateUpdate={handleDateUpdate}
          zoomLevel={zoomLevel}
        />
      </Flex>
    </div>
  );
};
