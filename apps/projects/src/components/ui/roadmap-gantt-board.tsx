"use client";

import { Box, Flex, Text, Tooltip, Avatar, DatePicker } from "ui";
import { differenceInDays, formatISO } from "date-fns";
import { useCallback, useState } from "react";
import { cn } from "lib";
import { useQueryClient } from "@tanstack/react-query";
import { CalendarPlusIcon } from "icons";
import type { DateRange } from "react-day-picker";
import { useSession } from "@/lib/auth/client";
import type { Objective, ObjectiveUpdate } from "@/modules/objectives/types";
import { useUpdateObjectiveMutation } from "@/modules/objectives/hooks/update-mutation";
import { useTeamMembers } from "@/lib/hooks/team-members";
import { useObjectiveStatuses } from "@/lib/hooks/objective-statuses";
import { useUserRole, useWorkspacePath } from "@/hooks";
import { objectiveKeys } from "@/modules/objectives/constants";
import { getObjective } from "@/modules/objectives/queries/get-objective";
import { PrioritiesMenu } from "@/components/ui/story/priorities-menu";
import { ObjectiveStatusesMenu } from "@/components/ui/objective-statuses-menu";
import { AssigneesMenu } from "@/components/ui/story/assignees-menu";
import { ObjectiveHealthIcon } from "@/components/ui/objective-health-icon";
import { PriorityIcon } from "./priority-icon";
import { ObjectiveStatusIcon } from "./objective-status-icon";
import { BaseGantt, GanttControls, type ZoomLevel } from "./base-gantt";

const ROADMAP_STICKY_COLUMNS_WIDTH = 640;
const ROADMAP_ROW_HEIGHT = "3.5rem";
const ROADMAP_COLUMNS = "grid-cols-[2rem_2rem_2rem_2rem_minmax(0,1fr)_5.25rem]";

// Individual Objective Row Component
const ObjectiveRow = ({
  objective,
  duration,
  handleUpdate,
  statusName,
  isSelected,
  onObjectiveSelect,
}: {
  objective: Objective;
  duration: number | null;
  handleUpdate: (objectiveId: string, data: ObjectiveUpdate) => void;
  statusName: string;
  isSelected: boolean;
  onObjectiveSelect: (objective: Objective) => void;
}) => {
  // Import userRole directly in this component
  const { userRole } = useUserRole();
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const queryClient = useQueryClient();
  const [dates, setDates] = useState<DateRange | undefined>(undefined);

  // Get team members for this specific objective's team
  const { data: members = [] } = useTeamMembers(objective.teamId);

  const selectedAssignee = members.find(
    (member) => member.id === objective.leadUser,
  );

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
          "group border-border hover:bg-state-hover grid h-14 items-center gap-4 border-b-[0.5px] px-4 transition-colors dark:hover:bg-white/[0.025]",
          ROADMAP_COLUMNS,
          { "bg-state-active/50 dark:bg-white/[0.03]": isSelected },
        )}
      >
        <Flex align="center" justify="center">
          <AssigneesMenu>
            <Tooltip
              className="py-2.5"
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
                  <button
                    className="flex"
                    disabled={userRole === "guest"}
                    type="button"
                  >
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
        <Flex align="center" justify="center">
          <ObjectiveStatusesMenu>
            <Tooltip title={statusName}>
              <span>
                <ObjectiveStatusesMenu.Trigger>
                  <button
                    aria-label={`Change status: ${statusName}`}
                    className="focus-visible:ring-primary flex rounded-sm outline-none focus-visible:ring-1 disabled:cursor-not-allowed disabled:opacity-50"
                    disabled={userRole === "guest"}
                    type="button"
                  >
                    <ObjectiveStatusIcon statusId={objective.statusId} />
                  </button>
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
          <Tooltip title={objective.health ?? "No health"}>
            <span className="flex">
              <ObjectiveHealthIcon health={objective.health} />
            </span>
          </Tooltip>
        </Flex>
        <Flex align="center" justify="center">
          <PrioritiesMenu>
            <Tooltip title={objective.priority ?? "No priority"}>
              <span>
                <PrioritiesMenu.Trigger>
                  <button
                    aria-label={`Change priority: ${objective.priority ?? "No priority"}`}
                    className="focus-visible:ring-primary flex rounded-sm outline-none select-none focus-visible:ring-1 disabled:cursor-not-allowed disabled:opacity-50"
                    disabled={userRole === "guest"}
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
            <Text className="line-clamp-1 hover:opacity-90" fontWeight="medium">
              {objective.name}
            </Text>
          </button>
        </Flex>
        <Flex justify="end">
          {duration !== null ? (
            <Text className="shrink-0 text-[0.95rem]" color="muted">
              {duration} day{duration !== 1 ? "s" : ""}
            </Text>
          ) : (
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
          )}
        </Flex>
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
      objectives: Objective[],
      onReset: () => void,
      sidebarZoomLevel: ZoomLevel,
      onZoomChange: (zoom: ZoomLevel) => void,
    ) => {
      return (
        <Box className="border-border/60 bg-background sticky left-0 z-20 w-screen shrink-0 border-r-[0.5px] md:w-160">
          <Box className="border-border bg-background sticky top-0 z-10 hidden h-16 items-center border-b-[0.5px] px-4 md:flex">
            <GanttControls
              className="w-full justify-between"
              onReset={onReset}
              onZoomChange={onZoomChange}
              zoomLevel={sidebarZoomLevel}
            />
          </Box>
          {objectives.map((objective) => {
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

            return (
              <ObjectiveRow
                duration={duration}
                handleUpdate={handleUpdate}
                isSelected={selectedObjectiveId === objective.id}
                key={objective.id}
                objective={objective}
                onObjectiveSelect={onObjectiveSelect}
                statusName={
                  statuses.find((status) => status.id === objective.statusId)
                    ?.name ?? "No status"
                }
              />
            );
          })}
        </Box>
      );
    },
    [handleUpdate, onObjectiveSelect, selectedObjectiveId, statuses],
  );

  // Render bar content
  const renderBarContent = useCallback(
    (objective: Objective) => (
      <Text className="line-clamp-1" fontWeight="medium">
        {objective.name}
      </Text>
    ),
    [],
  );

  return (
    <BaseGantt
      barClassName="dark:bg-white/[0.07] dark:hover:bg-white/[0.1]"
      className={className}
      controlledZoomLevel={zoomLevel}
      items={objectives}
      onBarClick={onObjectiveSelect}
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
