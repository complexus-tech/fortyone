"use client";

import { Box, Flex, Text, Tooltip, Avatar, DatePicker } from "ui";
import { differenceInDays, formatISO } from "date-fns";
import { useCallback, useState } from "react";
import Link from "next/link";
import { useQueryClient } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useRouter } from "next/navigation";
import { CalendarPlusIcon } from "icons";
import type { DateRange } from "react-day-picker";
import type { Objective } from "@/modules/objectives/types";
import { useUpdateObjectiveMutation } from "@/modules/objectives/hooks/update-mutation";
import { useTeamMembers } from "@/lib/hooks/team-members";
import { useUserRole } from "@/hooks";
import { objectiveKeys } from "@/modules/objectives/constants";
import { getObjective } from "@/modules/objectives/queries/get-objective";
import { PrioritiesMenu } from "@/components/ui/story/priorities-menu";
import { ObjectiveStatusesMenu } from "@/components/ui/objective-statuses-menu";
import { AssigneesMenu } from "@/components/ui/story/assignees-menu";
import { PriorityIcon } from "./priority-icon";
import { ObjectiveStatusIcon } from "./objective-status-icon";
import { BaseGantt, GanttHeader, type ZoomLevel } from "./base-gantt";

// Individual Objective Row Component
const ObjectiveRow = ({
  objective,
  duration,
  handleUpdate,
}: {
  objective: Objective;
  duration: number | null;
  handleUpdate: (objectiveId: string, data: Partial<Objective>) => void;
}) => {
  // Import userRole directly in this component
  const { userRole } = useUserRole();
  const { data: session } = useSession();
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
            <Text className="line-clamp-1 hover:opacity-90" fontWeight="medium">
              {objective.name}
            </Text>
          </Link>
        </Flex>

        {duration ? (
          <Text className="ml-4 shrink-0" color="muted">
            {duration} day{duration !== 1 ? "s" : ""}
          </Text>
        ) : (
          <DatePicker>
            <Tooltip title="Add dates">
              <span className="mt-1">
                <DatePicker.Trigger>
                  <button type="button">
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
                    endDate: formatISO(range.to, { representation: "date" }),
                  });
                }
              }}
              selected={dates}
            />
          </DatePicker>
        )}
      </Flex>
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
  const router = useRouter();

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

  // Handle bar clicks to navigate to objective page
  const handleBarClick = useCallback(
    (objective: Objective) => {
      router.push(`/teams/${objective.teamId}/objectives/${objective.id}`);
    },
    [router],
  );

  const handleUpdate = useCallback(
    (objectiveId: string, data: Partial<Objective>) => {
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
      zoomLevel: ZoomLevel,
      onZoomChange: (zoom: ZoomLevel) => void,
    ) => {
      return (
        <Box className="sticky left-0 z-20 w-screen shrink-0 border-r-[0.5px] border-gray-200/60 bg-white dark:border-dark-100 dark:bg-dark md:w-136">
          <GanttHeader
            onReset={onReset}
            onZoomChange={onZoomChange}
            zoomLevel={zoomLevel}
          />
          {objectives.map((objective) => {
            const startDate = objective.startDate
              ? new Date(objective.startDate)
              : null;
            const endDate = objective.endDate
              ? new Date(objective.endDate)
              : null;
            const duration =
              startDate && endDate
                ? differenceInDays(endDate, startDate)
                : null;

            return (
              <ObjectiveRow
                duration={duration}
                handleUpdate={handleUpdate}
                key={objective.id}
                objective={objective}
              />
            );
          })}
        </Box>
      );
    },
    [handleUpdate],
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
      className={className}
      items={objectives}
      onBarClick={handleBarClick}
      onDateUpdate={handleDateUpdate}
      renderBarContent={renderBarContent}
      renderSidebar={renderSidebar}
      storageKey="roadmapZoomLevel"
      zoomLevel="months"
    />
  );
};
