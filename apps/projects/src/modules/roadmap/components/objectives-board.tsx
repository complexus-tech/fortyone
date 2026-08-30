"use client";

import type { DragEndEvent, DragStartEvent } from "@dnd-kit/core";
import {
  DndContext,
  DragOverlay,
  PointerSensor,
  useSensor,
  useSensors,
} from "@dnd-kit/core";
import { format, formatISO } from "date-fns";
import { CalendarIcon } from "icons";
import { useMemo, useState } from "react";
import { Avatar, Button, DatePicker, Flex } from "ui";
import type {
  KeyResult,
  Objective,
  ObjectiveUpdate,
} from "@/modules/objectives/types";
import {
  AssigneesMenu,
  ObjectiveHealthIcon,
  PrioritiesMenu,
  PriorityIcon,
} from "@/components/ui";
import { ObjectiveStatusIcon } from "@/components/ui/objective-status-icon";
import { ObjectiveStatusesMenu } from "@/components/ui/objective-statuses-menu";
import { useMembers } from "@/lib/hooks/members";
import { useObjectiveStatuses } from "@/lib/hooks/objective-statuses";
import { ObjectiveHealthEditor } from "@/modules/objectives/components/objective-health-editor";
import { ObjectiveForecastRiskBadge } from "@/modules/objectives/components/objective-forecast-risk";
import {
  useCanUpdateObjective,
  useUpdateObjectiveMutation,
} from "@/modules/objectives/hooks";
import type { RoadmapLayoutType } from "@/modules/roadmap/types";
import { useTeams } from "@/modules/teams/hooks/teams";
import { hexToRgba } from "@/utils";
import {
  getObjectiveGroupUpdate,
  groupObjectives,
  type ObjectiveViewOptions,
} from "../objective-board-utils";
import { ObjectiveBoardCard } from "./objective-board-card";
import {
  OBJECTIVE_GROUP_DND_ID_PREFIX,
  ObjectivesKanban,
} from "./objectives-kanban";
import { ObjectivesGroupedList } from "./objectives-grouped-list";
import { RoadmapKeyResultSummary } from "./roadmap-key-results";

type ObjectivesBoardProps = {
  objectives: Objective[];
  layout: Extract<RoadmapLayoutType, "kanban" | "list">;
  viewOptions: ObjectiveViewOptions;
  setViewOptions: (viewOptions: ObjectiveViewOptions) => void;
  onObjectiveSelect: (objective: Objective) => void;
  onKeyResultSelect: (objective: Objective, keyResult: KeyResult) => void;
  onCreateObjective: () => void;
  selectedObjectiveId?: string;
};

export const ObjectivesBoard = ({
  objectives,
  layout,
  viewOptions,
  setViewOptions,
  onObjectiveSelect,
  onKeyResultSelect,
  onCreateObjective,
  selectedObjectiveId,
}: ObjectivesBoardProps) => {
  const { data: statuses = [] } = useObjectiveStatuses();
  const { data: members = [] } = useMembers();
  const { data: teams = [] } = useTeams();
  const updateMutation = useUpdateObjectiveMutation();
  const canUpdate = useCanUpdateObjective();
  const [activeObjective, setActiveObjective] = useState<Objective | null>(
    null,
  );
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 8 } }),
  );
  const groups = useMemo(
    () =>
      groupObjectives({
        objectives,
        statuses,
        members,
        viewOptions,
      }),
    [members, objectives, statuses, viewOptions],
  );
  const memberById = useMemo(
    () => new Map(members.map((member) => [member.id, member])),
    [members],
  );
  const objectivesById = useMemo(
    () => new Map(objectives.map((objective) => [objective.id, objective])),
    [objectives],
  );
  const statusById = useMemo(
    () => new Map(statuses.map((status) => [status.id, status])),
    [statuses],
  );
  const teamCodeById = useMemo(
    () => new Map(teams.map((team) => [team.id, team.code])),
    [teams],
  );

  const handleObjectiveUpdate = (
    objectiveId: string,
    data: ObjectiveUpdate,
  ) => {
    updateMutation.mutate({ objectiveId, data });
  };

  const renderCardControls = (objective: Objective) => {
    const lead = memberById.get(objective.leadUser);
    const status = statusById.get(objective.statusId);

    return (
      <Flex align="center" className="mt-1 gap-1.5" wrap>
        <ObjectiveForecastRiskBadge objective={objective} size="control" />
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
              handleObjectiveUpdate(objective.id, { leadUser });
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
              handleObjectiveUpdate(objective.id, { statusId });
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
              handleObjectiveUpdate(objective.id, { priority });
            }}
          />
        </PrioritiesMenu>
        <ObjectiveHealthEditor
          health={objective.health}
          objectiveId={objective.id}
        >
          <Button
            className="gap-1 pr-2"
            color="tertiary"
            disabled={!canUpdate}
            rounded="md"
            size="xs"
            type="button"
            variant="outline"
          >
            <ObjectiveHealthIcon health={objective.health} />
            {objective.health ?? "No Health"}
          </Button>
        </ObjectiveHealthEditor>
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
              handleObjectiveUpdate(objective.id, {
                endDate: formatISO(day, { representation: "date" }),
              });
            }}
            selected={
              objective.endDate ? new Date(objective.endDate) : undefined
            }
          />
        </DatePicker>
        <RoadmapKeyResultSummary
          objective={objective}
          onSelect={(keyResult) => {
            onKeyResultSelect(objective, keyResult);
          }}
        />
      </Flex>
    );
  };

  const handleDragStart = ({ active }: DragStartEvent) => {
    setActiveObjective(objectivesById.get(String(active.id)) ?? null);
  };

  const handleDragEnd = ({ active, over }: DragEndEvent) => {
    setActiveObjective(null);
    if (!over) return;

    const overId = String(over.id);
    if (!overId.startsWith(OBJECTIVE_GROUP_DND_ID_PREFIX)) return;

    const objective = objectivesById.get(String(active.id));
    if (!objective) return;

    const groupKey = overId.slice(OBJECTIVE_GROUP_DND_ID_PREFIX.length);
    handleObjectiveUpdate(
      objective.id,
      getObjectiveGroupUpdate(viewOptions.groupBy, groupKey),
    );
  };

  if (layout === "list") {
    return (
      <ObjectivesGroupedList
        groups={groups}
        onCreateObjective={onCreateObjective}
        onKeyResultSelect={onKeyResultSelect}
        onObjectiveSelect={onObjectiveSelect}
        selectedObjectiveId={selectedObjectiveId}
        teamCodeById={teamCodeById}
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
        activeObjectiveId={activeObjective?.id}
        canDrag={canUpdate}
        groups={groups}
        onCreateObjective={onCreateObjective}
        onObjectiveSelect={onObjectiveSelect}
        renderCardControls={renderCardControls}
        selectedObjectiveId={selectedObjectiveId}
        setViewOptions={setViewOptions}
        teamCodeById={teamCodeById}
        viewOptions={viewOptions}
      />
      <DragOverlay>
        {activeObjective ? (
          <ObjectiveBoardCard
            canDrag={false}
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
