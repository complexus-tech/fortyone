"use client";

import { Box, Flex, Text, Tooltip, Avatar, Button, DatePicker } from "ui";
import { differenceInDays, format, formatISO } from "date-fns";
import {
  useCallback,
  useMemo,
  useState,
  type CSSProperties,
  type ReactNode,
} from "react";
import { cn } from "lib";
import { useQueryClient } from "@tanstack/react-query";
import { CalendarPlusIcon } from "icons";
import type { DateRange } from "react-day-picker";
import { useSession } from "@/lib/auth/client";
import type { Objective, ObjectiveUpdate } from "@/modules/objectives/types";
import { useUpdateObjectiveMutation } from "@/modules/objectives/hooks/update-mutation";
import { useTeamMembers } from "@/lib/hooks/team-members";
import { useObjectiveStatuses } from "@/lib/hooks/objective-statuses";
import { useWorkspacePath } from "@/hooks";
import { objectiveKeys } from "@/modules/objectives/constants";
import { getObjective } from "@/modules/objectives/queries/get-objective";
import { ObjectiveHealthEditor } from "@/modules/objectives/components/objective-health-editor";
import { useCanUpdateObjective } from "@/modules/objectives/hooks";
import { PrioritiesMenu } from "@/components/ui/story/priorities-menu";
import { ObjectiveStatusesMenu } from "@/components/ui/objective-statuses-menu";
import { AssigneesMenu } from "@/components/ui/story/assignees-menu";
import { ObjectiveHealthIcon } from "@/components/ui/objective-health-icon";
import { hexToRgba } from "@/utils";
import { PriorityIcon } from "./priority-icon";
import { ObjectiveStatusIcon } from "./objective-status-icon";
import { BaseGantt, GanttControls, type ZoomLevel } from "./base-gantt";

const ROADMAP_STICKY_COLUMNS_WIDTH = 640;
const ROADMAP_ROW_HEIGHT = "3.5rem";
const ROADMAP_COLUMNS =
  "grid-cols-[2rem_minmax(0,7rem)_2rem_2rem_minmax(0,1fr)_5.25rem]";
const DAYS_PER_DISPLAY_MONTH = 30;

const formatObjectiveDuration = (days: number) => {
  if (days < DAYS_PER_DISPLAY_MONTH) {
    return `${days} day${days === 1 ? "" : "s"}`;
  }

  const months = Math.max(1, Math.round(days / DAYS_PER_DISPLAY_MONTH));
  return `${months} month${months === 1 ? "" : "s"}`;
};

type RoadmapGanttItem = {
  id: string;
  startDate: string | null;
  endDate: string | null;
  objective: Objective;
};

const getTimelineDate = (
  dates: (string | null | undefined)[],
  mode: "earliest" | "latest",
) => {
  let timestamp: number | null = null;
  for (const value of dates) {
    if (!value) continue;
    const candidate = new Date(value).getTime();
    if (Number.isNaN(candidate)) continue;

    if (
      timestamp === null ||
      (mode === "earliest" ? candidate < timestamp : candidate > timestamp)
    ) {
      timestamp = candidate;
    }
  }

  if (timestamp === null) return null;
  return formatISO(new Date(timestamp), { representation: "date" });
};

// Individual Objective Row Component
const ObjectiveRow = ({
  objective,
  duration,
  handleUpdate,
  statusName,
  statusColor,
  isSelected,
  onObjectiveSelect,
}: {
  objective: Objective;
  duration: number | null;
  handleUpdate: (objectiveId: string, data: ObjectiveUpdate) => void;
  statusName: string;
  statusColor?: string;
  isSelected: boolean;
  onObjectiveSelect: (objective: Objective) => void;
}) => {
  const canUpdate = useCanUpdateObjective();
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const queryClient = useQueryClient();
  const [dates, setDates] = useState<DateRange | undefined>(undefined);

  // Get team members for this specific objective's team
  const { data: members = [] } = useTeamMembers(objective.teamId);

  const selectedAssignee = members.find(
    (member) => member.id === objective.leadUser,
  );
  const rowStyle = {
    "--objective-row-background": hexToRgba(
      objective.color,
      isSelected ? 0.09 : 0.045,
    ),
    "--objective-row-background-dark": hexToRgba(
      objective.color,
      isSelected ? 0.11 : 0.06,
    ),
    "--objective-row-hover-background": hexToRgba(
      objective.color,
      isSelected ? 0.13 : 0.085,
    ),
    "--objective-row-hover-background-dark": hexToRgba(
      objective.color,
      isSelected ? 0.15 : 0.11,
    ),
  } as CSSProperties;
  let scheduleCell: ReactNode;

  if (objective.scheduleStatus === "at_risk") {
    scheduleCell = (
      <Tooltip
        className="pointer-events-none max-w-72"
        title={`Forecast ${
          objective.forecastEndDate
            ? format(new Date(objective.forecastEndDate), "MMM d, yyyy")
            : "is late"
        }${
          objective.forecastCauseStory
            ? `. Driven by task ${objective.forecastCauseStory.sequenceId}: ${objective.forecastCauseStory.title}`
            : ""
        }`}
      >
        <Text className="shrink-0 text-[0.95rem]" color="danger">
          +{objective.forecastDaysDelta}d
        </Text>
      </Tooltip>
    );
  } else if (duration !== null) {
    const durationLabel = formatObjectiveDuration(duration);
    scheduleCell = (
      <Tooltip
        className="pointer-events-none"
        title={`${duration} calendar day${duration === 1 ? "" : "s"}`}
      >
        <Text
          aria-label={`${durationLabel}, ${duration} calendar day${duration === 1 ? "" : "s"}`}
          className="shrink-0 text-[0.95rem]"
          color="muted"
        >
          {durationLabel}
        </Text>
      </Tooltip>
    );
  } else {
    scheduleCell = (
      <DatePicker>
        <Tooltip title="Add dates">
          <span className="mt-1">
            <DatePicker.Trigger>
              <button aria-label="Add objective dates" type="button">
                <CalendarPlusIcon />
              </button>
            </DatePicker.Trigger>
          </span>
        </Tooltip>
        <DatePicker.Calendar
          mode="range"
          numberOfMonths={2}
          onSelect={(range) => {
            setDates(range);
            if (range?.from && range.to) {
              handleUpdate(objective.id, {
                startDate: formatISO(range.from, {
                  representation: "date",
                }),
                endDate: formatISO(range.to, {
                  representation: "date",
                }),
              });
            }
          }}
          selected={dates}
        />
      </DatePicker>
    );
  }

  return (
    <Box
      onMouseEnter={() => {
        if (session) {
          const ctx = { session, workspaceSlug };
          queryClient.prefetchQuery({
            queryKey: objectiveKeys.objective(workspaceSlug, objective.id),
            queryFn: () => getObjective(objective.id, ctx),
          });
        }
      }}
    >
      <Box
        className={cn(
          "group border-border dark:border-border/70 grid h-14 items-center gap-4 border-b-[0.5px] bg-[var(--objective-row-background)] px-4 transition-colors duration-150 hover:bg-[var(--objective-row-hover-background)] dark:bg-[var(--objective-row-background-dark)] dark:hover:bg-[var(--objective-row-hover-background-dark)]",
          ROADMAP_COLUMNS,
        )}
        style={rowStyle}
      >
        <Flex align="center" justify="center">
          <AssigneesMenu>
            <Tooltip
              className="pointer-events-none py-2.5"
              title={
                selectedAssignee ? (
                  <Box>
                    <Flex gap={2}>
                      <Avatar
                        className="mt-0.5"
                        name={selectedAssignee.fullName}
                        size="sm"
                        src={selectedAssignee.avatarUrl}
                      />
                      <Box>
                        <Text fontSize="md" fontWeight="medium">
                          {selectedAssignee.fullName}
                        </Text>
                        <Text color="muted" fontSize="md">
                          ({selectedAssignee.username})
                        </Text>
                      </Box>
                    </Flex>
                  </Box>
                ) : null
              }
            >
              <span>
                <AssigneesMenu.Trigger>
                  <button className="flex" disabled={!canUpdate} type="button">
                    <Avatar
                      name={
                        selectedAssignee?.fullName || selectedAssignee?.username
                      }
                      size="xs"
                      src={selectedAssignee?.avatarUrl}
                    />
                  </button>
                </AssigneesMenu.Trigger>
              </span>
            </Tooltip>
            <AssigneesMenu.Items
              assigneeId={selectedAssignee?.id}
              onAssigneeSelected={(assigneeId) => {
                handleUpdate(objective.id, {
                  leadUser: assigneeId || undefined,
                });
              }}
              teamId={objective.teamId}
            />
          </AssigneesMenu>
        </Flex>
        <Flex align="center" className="min-w-0">
          <ObjectiveStatusesMenu>
            <Tooltip className="pointer-events-none" title={statusName}>
              <span className="block w-full min-w-0">
                <ObjectiveStatusesMenu.Trigger>
                  <Button
                    aria-label={`Change status: ${statusName}`}
                    className="w-max max-w-full min-w-0 gap-1 pr-2"
                    color="tertiary"
                    disabled={!canUpdate}
                    rounded="md"
                    size="xs"
                    style={{
                      backgroundColor: hexToRgba(statusColor, 0.1),
                      borderColor: hexToRgba(statusColor, 0.2),
                    }}
                    type="button"
                    variant="outline"
                  >
                    <ObjectiveStatusIcon statusId={objective.statusId} />
                    <span className="min-w-0 truncate">{statusName}</span>
                  </Button>
                </ObjectiveStatusesMenu.Trigger>
              </span>
            </Tooltip>
            <ObjectiveStatusesMenu.Items
              setStatusId={(statusId) => {
                handleUpdate(objective.id, { statusId });
              }}
              statusId={objective.statusId}
            />
          </ObjectiveStatusesMenu>
        </Flex>
        <Flex align="center" justify="center">
          <Tooltip
            className="pointer-events-none"
            title={objective.health ?? "No health"}
          >
            <span>
              <ObjectiveHealthEditor
                health={objective.health}
                objectiveId={objective.id}
              >
                <button
                  aria-label={`Change health: ${objective.health ?? "No health"}`}
                  className="focus-visible:ring-primary flex rounded-sm outline-none focus-visible:ring-1 disabled:cursor-not-allowed disabled:opacity-50"
                  disabled={!canUpdate}
                  type="button"
                >
                  <ObjectiveHealthIcon health={objective.health} />
                </button>
              </ObjectiveHealthEditor>
            </span>
          </Tooltip>
        </Flex>
        <Flex align="center" justify="center">
          <PrioritiesMenu>
            <Tooltip
              className="pointer-events-none"
              title={objective.priority ?? "No priority"}
            >
              <span>
                <PrioritiesMenu.Trigger>
                  <button
                    aria-label={`Change priority: ${objective.priority ?? "No priority"}`}
                    className="focus-visible:ring-primary flex rounded-sm outline-none select-none focus-visible:ring-1 disabled:cursor-not-allowed disabled:opacity-50"
                    disabled={!canUpdate}
                    type="button"
                  >
                    <PriorityIcon priority={objective.priority} />
                  </button>
                </PrioritiesMenu.Trigger>
              </span>
            </Tooltip>
            <PrioritiesMenu.Items
              priority={objective.priority}
              setPriority={(priority) => {
                handleUpdate(objective.id, { priority });
              }}
            />
          </PrioritiesMenu>
        </Flex>
        <Flex align="center" className="min-w-0 pr-3">
          <button
            className="focus-visible:ring-primary min-w-0 flex-1 rounded-sm text-left outline-none focus-visible:ring-1"
            onClick={() => {
              onObjectiveSelect(objective);
            }}
            type="button"
          >
            <Flex align="center" className="min-w-0" gap={2}>
              <span
                aria-hidden
                className="size-2.5 shrink-0 rounded-sm"
                style={{ backgroundColor: objective.color }}
              />
              <Text
                className="line-clamp-1 hover:opacity-90"
                fontWeight="medium"
              >
                {objective.name}
              </Text>
            </Flex>
          </button>
        </Flex>
        <Flex justify="end">{scheduleCell}</Flex>
      </Box>
    </Box>
  );
};

type RoadmapGanttBoardProps = {
  objectives: Objective[];
  className?: string;
  zoomLevel: ZoomLevel;
  onZoomLevelChange: (zoomLevel: ZoomLevel) => void;
  onObjectiveSelect: (objective: Objective) => void;
  selectedObjectiveId?: string | null;
};

export const RoadmapGanttBoard = ({
  objectives,
  className,
  zoomLevel,
  onZoomLevelChange,
  onObjectiveSelect,
  selectedObjectiveId,
}: RoadmapGanttBoardProps) => {
  const { mutate } = useUpdateObjectiveMutation();
  const { data: statuses = [] } = useObjectiveStatuses();
  const ganttItems = useMemo<RoadmapGanttItem[]>(
    () =>
      objectives.map((objective) => ({
        id: objective.id,
        startDate: getTimelineDate(
          [objective.startDate, objective.forecastStartDate],
          "earliest",
        ),
        endDate: getTimelineDate(
          [objective.endDate, objective.forecastEndDate],
          "latest",
        ),
        objective,
      })),
    [objectives],
  );

  // Handle date updates from drag operations
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

  const handleUpdate = useCallback(
    (objectiveId: string, data: ObjectiveUpdate) => {
      mutate({
        objectiveId,
        data,
      });
    },
    [mutate],
  );

  // Render sidebar for objectives
  const renderSidebar = useCallback(
    (
      items: RoadmapGanttItem[],
      onReset: () => void,
      sidebarZoomLevel: ZoomLevel,
      onZoomChange: (zoom: ZoomLevel) => void,
    ) => {
      return (
        <Box className="border-border bg-background dark:border-border/60 sticky left-0 z-40 w-screen shrink-0 border-r-[0.5px] md:w-160">
          <Box className="border-border bg-background sticky top-0 z-10 hidden h-16 items-center border-b-[0.5px] px-4 md:flex">
            <GanttControls
              className="w-full justify-between"
              onReset={onReset}
              onZoomChange={onZoomChange}
              zoomLevel={sidebarZoomLevel}
            />
          </Box>
          {items.map(({ objective }) => {
            const startDate = objective.startDate
              ? new Date(objective.startDate)
              : null;
            const endDate = objective.endDate
              ? new Date(objective.endDate)
              : null;
            const duration =
              startDate && endDate
                ? differenceInDays(endDate, startDate) + 1
                : null;
            const status = statuses.find(
              (status) => status.id === objective.statusId,
            );

            return (
              <ObjectiveRow
                duration={duration}
                handleUpdate={handleUpdate}
                isSelected={selectedObjectiveId === objective.id}
                key={objective.id}
                objective={objective}
                onObjectiveSelect={onObjectiveSelect}
                statusColor={status?.color}
                statusName={status?.name ?? "No status"}
              />
            );
          })}
        </Box>
      );
    },
    [handleUpdate, onObjectiveSelect, selectedObjectiveId, statuses],
  );

  // Render bar content
  const renderBarContent = useCallback((item: RoadmapGanttItem) => {
    const { objective } = item;
    const isAtRisk = objective.scheduleStatus === "at_risk";
    const timelineStart = item.startDate ? new Date(item.startDate) : null;
    const timelineEnd = item.endDate ? new Date(item.endDate) : null;
    const targetEnd = objective.endDate ? new Date(objective.endDate) : null;
    const timelineDays =
      timelineStart && timelineEnd
        ? Math.max(1, differenceInDays(timelineEnd, timelineStart))
        : 1;
    const targetPosition =
      timelineStart && targetEnd
        ? Math.min(
            100,
            Math.max(
              0,
              (differenceInDays(targetEnd, timelineStart) / timelineDays) * 100,
            ),
          )
        : 100;

    return (
      <Box
        className="absolute inset-0 overflow-hidden rounded-[inherit]"
        style={{
          background: isAtRisk
            ? `linear-gradient(to right, ${hexToRgba(
                objective.color,
                0.28,
              )} 0%, ${hexToRgba(
                objective.color,
                0.28,
              )} ${targetPosition}%, ${hexToRgba(
                objective.color,
                0.1,
              )} ${targetPosition}%, ${hexToRgba(objective.color, 0.1)} 100%)`
            : hexToRgba(objective.color, 0.22),
        }}
      >
        {isAtRisk && targetEnd ? (
          <span
            aria-label={`Target date ${format(targetEnd, "MMM d, yyyy")}`}
            className="absolute inset-y-0 w-px bg-current opacity-35"
            style={{ left: `${targetPosition}%` }}
          />
        ) : null}
        <Flex align="center" className="h-full min-w-0 px-3" gap={2}>
          <Text className="line-clamp-1 min-w-0" fontWeight="medium">
            {objective.name}
          </Text>
          {isAtRisk ? (
            <Text
              as="span"
              className="shrink-0 text-[0.8rem]"
              color="danger"
              fontWeight="semibold"
            >
              +{objective.forecastDaysDelta}d
            </Text>
          ) : null}
        </Flex>
      </Box>
    );
  }, []);

  return (
    <BaseGantt
      barClassName="hover:border-border-strong dark:hover:border-border-strong"
      className={className}
      controlledZoomLevel={zoomLevel}
      items={ganttItems}
      onBarClick={(item) => {
        onObjectiveSelect(item.objective);
      }}
      onDateUpdate={handleDateUpdate}
      onZoomLevelChange={onZoomLevelChange}
      renderBarContent={renderBarContent}
      renderSidebar={renderSidebar}
      rowHeight={ROADMAP_ROW_HEIGHT}
      stickyColumnsWidth={ROADMAP_STICKY_COLUMNS_WIDTH}
      storageKey="roadmapZoomLevel"
      zoomLevel="months"
    />
  );
};
