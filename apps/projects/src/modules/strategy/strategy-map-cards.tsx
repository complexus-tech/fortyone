"use client";

import { format, formatISO } from "date-fns";
import {
  CalendarIcon,
  ChevronRightIcon,
  DeleteIcon,
  EditIcon,
  ObjectiveIcon,
  OKRIcon,
} from "icons";
import { cn } from "lib";
import { Avatar, Box, Button, ContextMenu, DatePicker, Flex, Text } from "ui";
import {
  AssigneesMenu,
  ObjectiveHealthIcon,
  PrioritiesMenu,
  PriorityIcon,
} from "@/components/ui";
import { ObjectiveStatusIcon } from "@/components/ui/objective-status-icon";
import { ObjectiveStatusesMenu } from "@/components/ui/objective-statuses-menu";
import { useWorkspacePath } from "@/hooks";
import { getKeyResultProgress } from "@/shared/key-results/progress";
import { ObjectiveHealthEditor } from "@/modules/objectives/components/objective-health-editor";
import type {
  KeyResult,
  Objective,
  ObjectiveStatus,
  ObjectiveUpdate,
} from "@/modules/objectives/types";
import { hexToRgba } from "@/utils";
import type { StrategicPillar, StrategyMember } from "./types";
import {
  GOAL_NODE_WIDTH,
  KEY_RESULT_NODE_WIDTH,
  OBJECTIVE_NODE_WIDTH,
  PILLAR_NODE_WIDTH,
} from "./strategy-map-layout";
import {
  cardClasses,
  ContextMenuLabel,
  getStrategyDescriptionPreview,
  Metric,
  NodeEyebrow,
  objectivePropertyControlClasses,
  StrategyConceptInfo,
  StrategyProgressBar,
} from "./strategy-map-card-primitives";
import { StrategyMapObjectiveContextMenu } from "./strategy-map-objective-card-menu";
import { getObjectiveProgress } from "./strategy-map-progress";

export const UltimateGoalNodeCard = ({
  averageProgress,
  canEdit,
  description,
  objectiveCount,
  onEdit,
  onOpenDetails,
  pillarCount,
  title,
}: {
  averageProgress: number | null;
  canEdit: boolean;
  description: string | null;
  objectiveCount: number;
  onEdit: () => void;
  onOpenDetails: () => void;
  pillarCount: number;
  title: string;
}) => (
  <ContextMenu>
    <ContextMenu.Trigger>
      <Box
        className={cn(cardClasses, "relative px-7 py-6 text-center")}
        style={{ width: GOAL_NODE_WIDTH }}
      >
        <Flex align="center" className="gap-1.5" justify="center">
          <NodeEyebrow className="text-text-primary text-[0.74rem] font-bold">
            ◆ Ultimate goal
          </NodeEyebrow>
          <StrategyConceptInfo
            description="Defines what winning looks like for the organization and the long-term outcome every part of the strategy should support."
            label="ultimate goal"
          />
        </Flex>
        <button
          className="mt-3 w-full text-center"
          data-card-select
          onClick={(event) => {
            if (event.detail === 0) onOpenDetails();
          }}
          type="button"
        >
          <Text className="text-xl leading-7" fontWeight="semibold">
            {title || "Define your ultimate goal"}
          </Text>
        </button>
        <Text className="mx-auto mt-2 max-w-[36rem] leading-5" color="muted">
          {getStrategyDescriptionPreview(description) ||
            "Describe the long-term outcome that every pillar should support."}
        </Text>
        <Flex
          align="center"
          className="border-border mt-5 gap-0 border-t pt-4"
          justify="center"
        >
          <Metric label="Pillars" value={pillarCount} />
          <span aria-hidden className="bg-border h-8 w-px" />
          <Metric label="Objectives" value={objectiveCount} />
          <span aria-hidden className="bg-border h-8 w-px" />
          <Metric
            label="Avg progress"
            value={averageProgress === null ? "—" : `${averageProgress}%`}
          />
        </Flex>
      </Box>
    </ContextMenu.Trigger>
    {canEdit ? (
      <ContextMenu.Items className="w-52">
        <ContextMenu.Group>
          <ContextMenu.Item onSelect={onEdit}>
            <ContextMenuLabel icon={<EditIcon className="h-4 w-4" />}>
              Edit ultimate goal
            </ContextMenuLabel>
          </ContextMenu.Item>
        </ContextMenu.Group>
      </ContextMenu.Items>
    ) : null}
  </ContextMenu>
);

export const PillarNodeCard = ({
  canEdit,
  description,
  isDropTarget,
  name,
  objectiveCount,
  onDelete,
  onEdit,
  onOpenDetails,
}: {
  canEdit: boolean;
  description: string | null;
  isDropTarget: boolean;
  name: string;
  objectiveCount: number;
  onDelete: () => void;
  onEdit: () => void;
  onOpenDetails: () => void;
}) => (
  <ContextMenu>
    <ContextMenu.Trigger>
      <Box
        className={cn(
          cardClasses,
          "px-5 py-5",
          isDropTarget &&
            "border-foreground bg-surface-elevated ring-foreground/10 dark:bg-surface-prominent/55 ring-4",
        )}
        style={{ width: PILLAR_NODE_WIDTH }}
      >
        <Flex align="center" className="gap-1.5">
          <NodeEyebrow className="text-[0.74rem] font-semibold">
            Strategic pillar
          </NodeEyebrow>
          <StrategyConceptInfo
            description="A how-to-win choice that defines where the organization will focus and differentiate to achieve its ultimate goal."
            label="strategic pillar"
          />
        </Flex>
        <button
          className="mt-1 w-full text-left"
          data-card-select
          onClick={(event) => {
            if (event.detail === 0) onOpenDetails();
          }}
          type="button"
        >
          <Text className="line-clamp-2 text-[1.08rem]" fontWeight="semibold">
            {name}
          </Text>
        </button>
        {getStrategyDescriptionPreview(description) ? (
          <Text className="mt-1 line-clamp-3 leading-5" color="muted">
            {getStrategyDescriptionPreview(description)}
          </Text>
        ) : null}
        <Flex align="center" className="mt-2 gap-2">
          <ObjectiveIcon className="text-text-muted h-4 w-4" />
          <Text color="muted">
            {objectiveCount} objective{objectiveCount === 1 ? "" : "s"}
          </Text>
        </Flex>
      </Box>
    </ContextMenu.Trigger>
    {canEdit ? (
      <ContextMenu.Items className="w-52">
        <ContextMenu.Group>
          <ContextMenu.Item onSelect={onEdit}>
            <ContextMenuLabel icon={<EditIcon className="h-4 w-4" />}>
              Edit pillar
            </ContextMenuLabel>
          </ContextMenu.Item>
        </ContextMenu.Group>
        <ContextMenu.Separator />
        <ContextMenu.Group>
          <ContextMenu.Item
            className="text-danger dark:text-danger"
            onSelect={onDelete}
          >
            <ContextMenuLabel
              icon={
                <DeleteIcon className="text-danger dark:text-danger h-4 w-4" />
              }
            >
              Delete pillar
            </ContextMenuLabel>
          </ContextMenu.Item>
        </ContextMenu.Group>
      </ContextMenu.Items>
    ) : null}
  </ContextMenu>
);

export const KeyResultNodeCard = ({
  code,
  keyResult,
  memberById,
  onOpenDetails,
}: {
  code: string;
  keyResult: KeyResult;
  memberById: ReadonlyMap<string, StrategyMember>;
  onOpenDetails: () => void;
}) => {
  const progress = getKeyResultProgress(keyResult);
  const lead = keyResult.lead ? memberById.get(keyResult.lead) : undefined;

  return (
    <Box
      className={cn(cardClasses, "cursor-pointer px-4 py-4")}
      data-card-select
      onClick={(event) => {
        if (event.detail === 0) onOpenDetails();
      }}
      onKeyDown={(event) => {
        if (event.key !== "Enter" && event.key !== " ") return;
        event.preventDefault();
        onOpenDetails();
      }}
      role="button"
      style={{ width: KEY_RESULT_NODE_WIDTH }}
      tabIndex={0}
    >
      <Flex align="start" className="gap-3" justify="between">
        <Flex align="start" className="min-w-0 flex-1 gap-2">
          <OKRIcon
            className="text-text-muted mt-0.5 h-4.5 w-4.5 shrink-0"
            strokeWidth={2}
          />
          <Text
            className="line-clamp-2 text-[1.08rem] leading-[1.4rem]"
            fontWeight="semibold"
          >
            {keyResult.name}
          </Text>
        </Flex>
        {keyResult.sequenceId > 0 ? (
          <Text
            className="shrink-0 text-[0.9rem] leading-[1.4rem] uppercase"
            color="muted"
          >
            {code}
          </Text>
        ) : null}
      </Flex>
      <StrategyProgressBar className="mt-3" progress={progress} />
      {lead ? (
        <Flex align="center" className="mt-3 gap-2">
          <Avatar
            name={lead.fullName || lead.username}
            rounded="md"
            size="xs"
            src={lead.avatarUrl}
          />
          <Text className="min-w-0 truncate text-sm" color="muted">
            {lead.fullName || lead.username}
          </Text>
        </Flex>
      ) : null}
    </Box>
  );
};

export const ObjectiveNodeCard = ({
  canEdit,
  currentPillarId,
  isKeyResultsExpanded,
  keyResultCount,
  keyResults,
  memberById,
  objective,
  onAlign,
  onOpenDetails,
  onToggleKeyResults,
  onUpdate,
  pillars,
  status,
  statuses,
  teamCode,
}: {
  canEdit: boolean;
  currentPillarId: string | null;
  isKeyResultsExpanded: boolean;
  keyResultCount: number;
  keyResults: KeyResult[];
  memberById: ReadonlyMap<string, StrategyMember>;
  objective: Objective;
  onAlign: (objectiveId: string, pillarId: string | null) => void;
  onOpenDetails: () => void;
  onToggleKeyResults: () => void;
  onUpdate: (objectiveId: string, data: ObjectiveUpdate) => void;
  pillars: StrategicPillar[];
  status?: ObjectiveStatus;
  statuses: ObjectiveStatus[];
  teamCode?: string;
}) => {
  const { withWorkspace } = useWorkspacePath();
  const progress = getObjectiveProgress(objective, keyResults);
  const lead = memberById.get(objective.leadUser);
  const objectivePath = withWorkspace(
    `/teams/${objective.teamId}/objectives/${objective.id}`,
  );
  const objectiveReference = teamCode
    ? `${teamCode}-${objective.sequenceId}`
    : String(objective.sequenceId);
  return (
    <ContextMenu>
      <ContextMenu.Trigger>
        <Box
          className={cn(cardClasses, "px-4 py-4")}
          style={{ width: OBJECTIVE_NODE_WIDTH }}
        >
          <Flex align="start" className="gap-3" justify="between">
            <button
              className="focus-visible:ring-foreground/30 flex min-w-0 flex-1 items-start gap-2 rounded-sm text-left outline-none focus-visible:ring-1"
              data-card-select
              onClick={(event) => {
                if (event.detail === 0) onOpenDetails();
              }}
              type="button"
            >
              <ObjectiveIcon
                className="text-text-muted mt-0.5 h-4.5 w-4.5 shrink-0"
                strokeWidth={2}
              />
              <Text
                className="line-clamp-2 text-[1.15rem] leading-[1.45rem]"
                fontWeight="semibold"
              >
                {objective.name}
              </Text>
            </button>
            {objective.sequenceId > 0 ? (
              <Text
                className="shrink-0 text-[0.95rem] leading-[1.45rem] uppercase"
                color="muted"
              >
                {objectiveReference}
              </Text>
            ) : null}
          </Flex>
          <StrategyProgressBar className="mt-3" progress={progress} />
          <Flex align="center" className="mt-3 gap-1.5" data-no-drag wrap>
            <AssigneesMenu>
              <AssigneesMenu.Trigger>
                <Button
                  aria-label={lead ? `Lead: ${lead.username}` : "Add lead"}
                  asIcon
                  className={cn(objectivePropertyControlClasses, "gap-1 px-1")}
                  color="tertiary"
                  disabled={!canEdit}
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
                  onUpdate(objective.id, { leadUser });
                }}
                teamId={objective.teamId}
              />
            </AssigneesMenu>
            <ObjectiveStatusesMenu>
              <ObjectiveStatusesMenu.Trigger>
                <Button
                  className="dark:border-foreground/15 gap-1 pr-2"
                  disabled={!canEdit}
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
                  onUpdate(objective.id, { statusId });
                }}
                statusId={objective.statusId}
              />
            </ObjectiveStatusesMenu>
            <PrioritiesMenu>
              <PrioritiesMenu.Trigger>
                <Button
                  className={cn(objectivePropertyControlClasses, "gap-1 pr-2")}
                  color="tertiary"
                  disabled={!canEdit}
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
                  onUpdate(objective.id, { priority });
                }}
              />
            </PrioritiesMenu>
            <ObjectiveHealthEditor
              health={objective.health}
              objectiveId={objective.id}
            >
              <Button
                className={cn(objectivePropertyControlClasses, "gap-1 px-1")}
                color="tertiary"
                disabled={!canEdit}
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
                  className={cn(objectivePropertyControlClasses, "gap-1 pr-2")}
                  color="tertiary"
                  disabled={!canEdit}
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
                  onUpdate(objective.id, {
                    endDate: formatISO(day, { representation: "date" }),
                  });
                }}
                selected={
                  objective.endDate ? new Date(objective.endDate) : undefined
                }
              />
            </DatePicker>
          </Flex>
          {keyResultCount > 0 ? (
            <button
              aria-expanded={isKeyResultsExpanded}
              className="border-border text-foreground hover:bg-state-hover mt-3 flex w-full items-center justify-between rounded-md border-t px-1 pt-3 pb-1 text-left text-[0.95rem] transition-colors"
              data-no-drag
              onClick={onToggleKeyResults}
              type="button"
            >
              <span>
                {keyResultCount} key result{keyResultCount === 1 ? "" : "s"}
              </span>
              <ChevronRightIcon
                className={cn(
                  "h-4 w-4 shrink-0 transition-transform duration-150",
                  isKeyResultsExpanded && "rotate-90",
                )}
                strokeWidth={2}
              />
            </button>
          ) : null}
        </Box>
      </ContextMenu.Trigger>
      <StrategyMapObjectiveContextMenu
        canEdit={canEdit}
        currentPillarId={currentPillarId}
        objectiveId={objective.id}
        objectivePath={objectivePath}
        objectiveStatusId={objective.statusId}
        onAlign={onAlign}
        onSetStatus={(statusId) => {
          onUpdate(objective.id, { statusId });
        }}
        pillars={pillars}
        status={status}
        statuses={statuses}
      />
    </ContextMenu>
  );
};
