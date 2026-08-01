"use client";

import type { DragEndEvent, DragStartEvent } from "@dnd-kit/core";
import {
  DndContext,
  DragOverlay,
  PointerSensor,
  useDraggable,
  useDroppable,
  useSensor,
  useSensors,
} from "@dnd-kit/core";
import { format, formatISO } from "date-fns";
import {
  ArrowDownIcon,
  CalendarIcon,
  MoreHorizontalIcon,
  ObjectiveIcon,
  PlusIcon,
} from "icons";
import { cn } from "lib";
import { useMemo, useState } from "react";
import {
  Avatar,
  Box,
  Button,
  DatePicker,
  Flex,
  Popover,
  Text,
  Tooltip,
} from "ui";
import { BodyContainer } from "@/components/shared/body";
import {
  AssigneesMenu,
  ObjectiveHealthIcon,
  PrioritiesMenu,
  PriorityIcon,
} from "@/components/ui";
import { ObjectiveStatusIcon } from "@/components/ui/objective-status-icon";
import { ObjectiveStatusesMenu } from "@/components/ui/objective-statuses-menu";
import { useIsAdminOrOwner } from "@/hooks/owner";
import { useTerminology, useUserRole } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { useMembers } from "@/lib/hooks/members";
import { useObjectiveStatuses } from "@/lib/hooks/objective-statuses";
import { useTeams } from "@/modules/teams/hooks/teams";
import { ObjectiveCard } from "@/modules/objectives/components/card";
import { TableHeader } from "@/modules/objectives/components/heading";
import { useUpdateObjectiveMutation } from "@/modules/objectives/hooks";
import type { Objective, ObjectiveUpdate } from "@/modules/objectives/types";
import type { RoadmapLayoutType } from "@/modules/roadmap/types";
import { hexToRgba } from "@/utils";
import {
  getHiddenObjectiveGroupKeys,
  getObjectiveGroupUpdate,
  groupObjectives,
  hideObjectiveGroup,
  showObjectiveGroup,
  type ObjectiveGroup,
  type ObjectiveViewOptions,
} from "../objective-board-utils";

const GROUP_ID_PREFIX = "objective-group:";

const ObjectiveGroupIdentity = ({
  group,
  groupBy,
}: {
  group: ObjectiveGroup;
  groupBy: ObjectiveViewOptions["groupBy"];
}) => {
  if (groupBy === "status") {
    return (
      <>
        <ObjectiveStatusIcon statusId={group.status?.id} />
        <Text className="max-w-[20ch] truncate">
          {group.status?.name ?? "Unknown status"}
        </Text>
      </>
    );
  }

  if (groupBy === "priority") {
    return (
      <>
        <PriorityIcon priority={group.priority} />
        <Text>{group.priority ?? "No Priority"}</Text>
      </>
    );
  }

  return (
    <>
      <Avatar
        name={group.member?.fullName || group.member?.username}
        size="xs"
        src={group.member?.avatarUrl}
      />
      <Text className="max-w-[20ch] truncate">
        {group.member?.username ?? "Unassigned"}
      </Text>
    </>
  );
};

const ObjectiveBoardCard = ({
  objective,
  teamCode,
  onSelect,
  isOverlay = false,
}: {
  objective: Objective;
  teamCode?: string;
  onSelect: (objective: Objective) => void;
  isOverlay?: boolean;
}) => {
  const { data: session } = useSession();
  const { data: members = [] } = useMembers();
  const { data: statuses = [] } = useObjectiveStatuses();
  const updateMutation = useUpdateObjectiveMutation();
  const { isAdminOrOwner } = useIsAdminOrOwner(objective.createdBy);
  const canUpdate = isAdminOrOwner || session?.user.id === objective.leadUser;
  const lead = members.find(({ id }) => id === objective.leadUser);
  const status = statuses.find(({ id }) => id === objective.statusId);
  const objectiveReference = teamCode
    ? `${teamCode}-${objective.sequenceId}`
    : String(objective.sequenceId);
  const { attributes, isDragging, listeners, setNodeRef } = useDraggable({
    id: objective.id,
    disabled: !canUpdate || isOverlay,
  });

  const handleUpdate = (data: ObjectiveUpdate) => {
    updateMutation.mutate({ objectiveId: objective.id, data });
  };

  return (
    <div
      className={cn(
        "border-border shadow-shadow hover:bg-surface-elevated dark:bg-surface w-[340px] rounded-xl border-[0.5px] bg-white px-4 pb-4 shadow-lg backdrop-blur transition duration-200 ease-linear select-none",
        {
          "rotate-2 shadow-xl": isOverlay,
          "bg-surface-muted opacity-60": isDragging,
        },
      )}
      ref={setNodeRef}
    >
      <Box
        className={cn("cursor-grab pt-3 pb-1.5", {
          "cursor-grabbing": isDragging,
        })}
        {...attributes}
        {...listeners}
      >
        <button
          className="focus-visible:ring-primary flex w-full justify-between gap-2 rounded-sm text-left outline-none focus-visible:ring-1"
          onClick={() => {
            if (!isDragging) onSelect(objective);
          }}
          type="button"
        >
          <Text className="line-clamp-3 text-[1.1rem] leading-[1.4rem]">
            {objective.name}
          </Text>
          {objective.sequenceId > 0 ? (
            <Text
              className="shrink-0 text-[0.95rem] leading-[1.4rem] uppercase"
              color="muted"
            >
              {objectiveReference}
            </Text>
          ) : null}
        </button>
      </Box>
      <Flex align="center" className="mt-1 gap-1.5" wrap>
        <AssigneesMenu>
          <AssigneesMenu.Trigger>
            <Button
              aria-label={lead ? `Lead: ${lead.username}` : "Add lead"}
              asIcon
              className="gap-1 px-1"
              color="tertiary"
              disabled={!canUpdate}
              size="xs"
              type="button"
              variant="outline"
            >
              <Avatar
                name={lead?.fullName || lead?.username}
                rounded="md"
                size="xs"
                src={lead?.avatarUrl}
              />
            </Button>
          </AssigneesMenu.Trigger>
          <AssigneesMenu.Items
            assigneeId={objective.leadUser}
            onAssigneeSelected={(leadUser) => {
              handleUpdate({ leadUser });
            }}
            teamId={objective.teamId}
          />
        </AssigneesMenu>
        <ObjectiveStatusesMenu>
          <ObjectiveStatusesMenu.Trigger>
            <Button
              className="gap-1 pr-2"
              disabled={!canUpdate}
              rounded="md"
              size="xs"
              style={{
                backgroundColor: hexToRgba(status?.color, 0.1),
                borderColor: hexToRgba(status?.color, 0.2),
              }}
              type="button"
              variant="outline"
            >
              <ObjectiveStatusIcon statusId={objective.statusId} />
              {status?.name ?? "Status"}
            </Button>
          </ObjectiveStatusesMenu.Trigger>
          <ObjectiveStatusesMenu.Items
            setStatusId={(statusId) => {
              handleUpdate({ statusId });
            }}
            statusId={objective.statusId}
          />
        </ObjectiveStatusesMenu>
        <PrioritiesMenu>
          <PrioritiesMenu.Trigger>
            <Button
              className="gap-1 pr-2"
              color="tertiary"
              disabled={!canUpdate}
              size="xs"
              type="button"
              variant="outline"
            >
              <PriorityIcon
                className="h-[1.15rem]"
                priority={objective.priority}
              />
              {objective.priority ?? "No Priority"}
            </Button>
          </PrioritiesMenu.Trigger>
          <PrioritiesMenu.Items
            priority={objective.priority}
            setPriority={(priority) => {
              handleUpdate({ priority });
            }}
          />
        </PrioritiesMenu>
        <Button
          className="gap-1 pr-2"
          color="tertiary"
          rounded="md"
          size="xs"
          type="button"
          variant="outline"
        >
          <ObjectiveHealthIcon health={objective.health} />
          {objective.health ?? "No Health"}
        </Button>
        <DatePicker>
          <DatePicker.Trigger>
            <Button
              className="gap-1 pr-2"
              color="tertiary"
              disabled={!canUpdate}
              leftIcon={<CalendarIcon className="h-4" />}
              rounded="md"
              size="xs"
              variant="outline"
            >
              {objective.endDate
                ? format(new Date(objective.endDate), "MMM d")
                : "Target"}
            </Button>
          </DatePicker.Trigger>
          <DatePicker.Calendar
            onDayClick={(day) => {
              handleUpdate({
                endDate: formatISO(day, { representation: "date" }),
              });
            }}
            selected={
              objective.endDate ? new Date(objective.endDate) : undefined
            }
          />
        </DatePicker>
      </Flex>
    </div>
  );
};

const ObjectiveKanbanColumn = ({
  group,
  onObjectiveSelect,
  teamCodeById,
}: {
  group: ObjectiveGroup;
  onObjectiveSelect: (objective: Objective) => void;
  teamCodeById: Map<string, string>;
}) => {
  const { setNodeRef, isOver } = useDroppable({
    id: `${GROUP_ID_PREFIX}${group.key}`,
  });

  return (
    <div
      className={cn(
        "flex h-full w-[340px] flex-col gap-3 overflow-y-auto rounded-md pb-6 transition-colors",
        { "bg-surface-muted": isOver },
      )}
      ref={setNodeRef}
    >
      {group.objectives.map((objective) => (
        <ObjectiveBoardCard
          key={objective.id}
          objective={objective}
          onSelect={onObjectiveSelect}
          teamCode={teamCodeById.get(objective.teamId)}
        />
      ))}
    </div>
  );
};

const ObjectiveGroupHeader = ({
  group,
  groupBy,
  onCreateObjective,
  collapsible = false,
  isCollapsed = false,
  onToggle,
  onHide,
}: {
  group: ObjectiveGroup;
  groupBy: ObjectiveViewOptions["groupBy"];
  onCreateObjective: () => void;
  collapsible?: boolean;
  isCollapsed?: boolean;
  onToggle?: () => void;
  onHide?: () => void;
}) => {
  const { userRole } = useUserRole();
  const { getTermDisplay } = useTerminology();

  return (
    <Flex align="center" gap={2} justify="between">
      <Flex align="center" className="min-w-0" gap={2}>
        <button
          className="focus-visible:ring-primary flex min-w-0 items-center gap-2 rounded-sm outline-none focus-visible:ring-1 disabled:cursor-default"
          disabled={!collapsible}
          onClick={onToggle}
          type="button"
        >
          {collapsible ? (
            <ArrowDownIcon
              className={cn("text-text-muted h-4 w-auto transition", {
                "-rotate-90": isCollapsed,
              })}
              strokeWidth={1}
            />
          ) : null}
          <ObjectiveGroupIdentity group={group} groupBy={groupBy} />
        </button>
        <Tooltip
          title={`Total ${getTermDisplay("objectiveTerm", { variant: "plural" })}`}
        >
          <span>
            <ObjectiveIcon className="text-text-muted h-5 w-auto" />
          </span>
        </Tooltip>
        <Text color="muted">
          {group.objectives.length}{" "}
          {getTermDisplay("objectiveTerm", {
            variant: group.objectives.length === 1 ? "singular" : "plural",
          })}
        </Text>
      </Flex>
      <Flex align="center" gap={1}>
        {onHide ? (
          <Popover>
            <Popover.Trigger asChild>
              <Button
                aria-label="Column options"
                color="tertiary"
                size="sm"
                variant="naked"
              >
                <MoreHorizontalIcon
                  className="h-[1.15rem] w-auto"
                  strokeWidth={4}
                />
              </Button>
            </Popover.Trigger>
            <Popover.Content align="end" className="w-44 p-1.5">
              <Button
                className="justify-start px-2"
                color="tertiary"
                fullWidth
                onClick={onHide}
                size="sm"
                variant="naked"
              >
                Hide column
              </Button>
            </Popover.Content>
          </Popover>
        ) : null}
        <Button
          aria-label={`New ${getTermDisplay("objectiveTerm")}`}
          color="tertiary"
          disabled={userRole === "guest"}
          onClick={onCreateObjective}
          size="sm"
          variant="naked"
        >
          <PlusIcon className="h-[1.2rem] w-auto" />
        </Button>
      </Flex>
    </Flex>
  );
};

const HiddenObjectiveGroups = ({
  groups,
  groupBy,
  onShow,
}: {
  groups: ObjectiveGroup[];
  groupBy: ObjectiveViewOptions["groupBy"];
  onShow: (groupKey: string) => void;
}) => {
  if (groups.length === 0) return null;

  return (
    <Box className="w-[340px] shrink-0">
      <Flex direction="column" gap={3}>
        {groups.map((group) => (
          <Box
            className="border-border bg-surface hover:bg-surface-elevated group flex min-h-14 cursor-pointer items-center justify-between rounded-xl border-[0.5px] px-4 transition duration-200 ease-linear select-none"
            key={group.key}
            onClick={() => {
              onShow(group.key);
            }}
            onKeyDown={(event) => {
              if (event.key === "Enter" || event.key === " ") {
                event.preventDefault();
                onShow(group.key);
              }
            }}
            role="button"
            tabIndex={0}
          >
            <Flex align="center" className="min-w-0" gap={2}>
              <ObjectiveGroupIdentity group={group} groupBy={groupBy} />
            </Flex>
            <Flex align="center" className="shrink-0" gap={1}>
              <Popover>
                <Popover.Trigger asChild>
                  <Button
                    aria-label="Hidden column options"
                    className="opacity-0 transition group-hover:opacity-100 focus-visible:opacity-100"
                    color="tertiary"
                    onClick={(event) => {
                      event.stopPropagation();
                    }}
                    size="sm"
                    variant="naked"
                  >
                    <MoreHorizontalIcon
                      className="h-[1.15rem] w-auto"
                      strokeWidth={4}
                    />
                  </Button>
                </Popover.Trigger>
                <Popover.Content align="end" className="w-40 p-1.5">
                  <Button
                    className="justify-start px-2"
                    color="tertiary"
                    fullWidth
                    onClick={() => {
                      onShow(group.key);
                    }}
                    size="sm"
                    variant="naked"
                  >
                    Show column
                  </Button>
                </Popover.Content>
              </Popover>
              <Text color="muted">{group.objectives.length}</Text>
            </Flex>
          </Box>
        ))}
      </Flex>
    </Box>
  );
};

const ObjectivesKanban = ({
  groups,
  viewOptions,
  setViewOptions,
  teamCodeById,
  onObjectiveSelect,
  onCreateObjective,
}: {
  groups: ObjectiveGroup[];
  viewOptions: ObjectiveViewOptions;
  setViewOptions: (viewOptions: ObjectiveViewOptions) => void;
  teamCodeById: Map<string, string>;
  onObjectiveSelect: (objective: Objective) => void;
  onCreateObjective: () => void;
}) => {
  const hiddenGroupKeys = getHiddenObjectiveGroupKeys(viewOptions);
  const visibleGroups = groups.filter(
    (group) => !hiddenGroupKeys.includes(group.key),
  );
  const hiddenGroups = groups.filter((group) =>
    hiddenGroupKeys.includes(group.key),
  );

  return (
    <BodyContainer className="dark:bg-background h-full overflow-x-auto bg-white">
      <Box className="sticky top-0 z-1 h-14 w-max px-6 backdrop-blur">
        <Flex align="center" className="h-full" gap={6}>
          {visibleGroups.map((group) => (
            <Box className="w-[340px] pl-1" key={group.key}>
              <ObjectiveGroupHeader
                group={group}
                groupBy={viewOptions.groupBy}
                onCreateObjective={onCreateObjective}
                onHide={() => {
                  setViewOptions(hideObjectiveGroup(viewOptions, group.key));
                }}
              />
            </Box>
          ))}
          {hiddenGroups.length > 0 ? (
            <Flex align="center" className="w-[340px] pl-1" gap={2}>
              <Text color="muted" fontWeight="medium">
                Hidden columns
              </Text>
              <Text color="muted">{hiddenGroups.length}</Text>
            </Flex>
          ) : null}
        </Flex>
      </Box>
      <Box className="flex h-[calc(100%-3.5rem)] w-max gap-x-6 px-7">
        {visibleGroups.map((group) => (
          <ObjectiveKanbanColumn
            group={group}
            key={group.key}
            onObjectiveSelect={onObjectiveSelect}
            teamCodeById={teamCodeById}
          />
        ))}
        <HiddenObjectiveGroups
          groupBy={viewOptions.groupBy}
          groups={hiddenGroups}
          onShow={(groupKey) => {
            setViewOptions(showObjectiveGroup(viewOptions, groupKey));
          }}
        />
      </Box>
    </BodyContainer>
  );
};

const ObjectivesGroupedList = ({
  groups,
  viewOptions,
  onObjectiveSelect,
  onCreateObjective,
}: {
  groups: ObjectiveGroup[];
  viewOptions: ObjectiveViewOptions;
  onObjectiveSelect: (objective: Objective) => void;
  onCreateObjective: () => void;
}) => {
  const [collapsedGroups, setCollapsedGroups] = useState<string[]>([]);

  return (
    <BodyContainer className="h-full overflow-x-auto pb-6">
      <Box className="min-w-6xl">
        <TableHeader />
        {groups.map((group) => {
          const isCollapsed = collapsedGroups.includes(group.key);
          return (
            <Box key={group.key}>
              <Box className="border-border bg-surface-muted/85 border-b-[0.5px] px-12 py-[0.4rem] backdrop-blur">
                <ObjectiveGroupHeader
                  collapsible
                  group={group}
                  groupBy={viewOptions.groupBy}
                  isCollapsed={isCollapsed}
                  onCreateObjective={onCreateObjective}
                  onToggle={() => {
                    setCollapsedGroups((current) =>
                      current.includes(group.key)
                        ? current.filter((key) => key !== group.key)
                        : [...current, group.key],
                    );
                  }}
                />
              </Box>
              {isCollapsed
                ? null
                : group.objectives.map((objective) => (
                    <ObjectiveCard
                      key={objective.id}
                      {...objective}
                      onSelect={() => {
                        onObjectiveSelect(objective);
                      }}
                    />
                  ))}
            </Box>
          );
        })}
      </Box>
    </BodyContainer>
  );
};

export const ObjectivesBoard = ({
  objectives,
  layout,
  viewOptions,
  setViewOptions,
  onObjectiveSelect,
  onCreateObjective,
}: {
  objectives: Objective[];
  layout: Extract<RoadmapLayoutType, "kanban" | "list">;
  viewOptions: ObjectiveViewOptions;
  setViewOptions: (viewOptions: ObjectiveViewOptions) => void;
  onObjectiveSelect: (objective: Objective) => void;
  onCreateObjective: () => void;
}) => {
  const { data: statuses = [] } = useObjectiveStatuses();
  const { data: members = [] } = useMembers();
  const { data: teams = [] } = useTeams();
  const updateMutation = useUpdateObjectiveMutation();
  const [activeObjective, setActiveObjective] = useState<Objective | null>(
    null,
  );
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 8 } }),
  );
  const groups = useMemo(
    () => groupObjectives({ objectives, statuses, members, viewOptions }),
    [members, objectives, statuses, viewOptions],
  );
  const teamCodeById = useMemo(
    () => new Map(teams.map((team) => [team.id, team.code])),
    [teams],
  );

  const handleDragStart = ({ active }: DragStartEvent) => {
    setActiveObjective(
      objectives.find(({ id }) => id === active.id.toString()) ?? null,
    );
  };

  const handleDragEnd = ({ active, over }: DragEndEvent) => {
    setActiveObjective(null);
    if (!over) return;

    const overId = over.id.toString();
    if (!overId.startsWith(GROUP_ID_PREFIX)) return;

    const objective = objectives.find(({ id }) => id === active.id.toString());
    if (!objective) return;

    const groupKey = overId.slice(GROUP_ID_PREFIX.length);
    updateMutation.mutate({
      objectiveId: objective.id,
      data: getObjectiveGroupUpdate(viewOptions.groupBy, groupKey),
    });
  };

  if (layout === "list") {
    return (
      <ObjectivesGroupedList
        groups={groups}
        onCreateObjective={onCreateObjective}
        onObjectiveSelect={onObjectiveSelect}
        viewOptions={viewOptions}
      />
    );
  }

  return (
    <DndContext
      onDragCancel={() => {
        setActiveObjective(null);
      }}
      onDragEnd={handleDragEnd}
      onDragStart={handleDragStart}
      sensors={sensors}
    >
      <ObjectivesKanban
        groups={groups}
        onCreateObjective={onCreateObjective}
        onObjectiveSelect={onObjectiveSelect}
        setViewOptions={setViewOptions}
        teamCodeById={teamCodeById}
        viewOptions={viewOptions}
      />
      <DragOverlay>
        {activeObjective ? (
          <ObjectiveBoardCard
            isOverlay
            objective={activeObjective}
            onSelect={() => {}}
            teamCode={teamCodeById.get(activeObjective.teamId)}
          />
        ) : null}
      </DragOverlay>
    </DndContext>
  );
};
