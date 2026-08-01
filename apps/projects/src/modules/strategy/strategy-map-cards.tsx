"use client";

import type { CSSProperties } from "react";
import { memo, useState } from "react";
import { format, formatISO } from "date-fns";
import Link from "next/link";
import {
  CalendarIcon,
  ChevronRightIcon,
  DeleteIcon,
  EditIcon,
  ExternalLinkIcon,
  ObjectiveIcon,
  UnlinkIcon,
} from "icons";
import { cn } from "lib";
import {
  Avatar,
  Box,
  Button,
  CircleProgressBar,
  ContextMenu,
  DatePicker,
  Flex,
  Text,
  Tooltip,
} from "ui";
import {
  AssigneesMenu,
  ObjectiveHealthIcon,
  PrioritiesMenu,
  PriorityIcon,
} from "@/components/ui";
import { ObjectiveStatusIcon } from "@/components/ui/objective-status-icon";
import { ObjectiveStatusesMenu } from "@/components/ui/objective-statuses-menu";
import { useWorkspacePath } from "@/hooks";
import { useMembers } from "@/lib/hooks/members";
import { ObjectiveHealthEditor } from "@/modules/objectives/components/objective-health-editor";
import { useKeyResults } from "@/modules/objectives/hooks";
import type {
  KeyResult,
  Objective,
  ObjectiveStatus,
  ObjectiveUpdate,
} from "@/modules/objectives/types";
import { hexToRgba } from "@/utils";
import type { StrategicPillar } from "./types";
import {
  GOAL_NODE_WIDTH,
  OBJECTIVE_NODE_WIDTH,
  PILLAR_NODE_WIDTH,
} from "./strategy-map-layout";

const NUMBER_FORMAT = new Intl.NumberFormat();

const getStrategyDescriptionPreview = (description: string | null) => {
  if (!description) return "";
  if (!/<\/?[a-z][\s\S]*>/i.test(description)) return description.trim();

  return description
    .replace(/<[^>]*>/g, " ")
    .replace(/&nbsp;/g, " ")
    .replace(/&amp;/g, "&")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/&#39;/g, "'")
    .replace(/&quot;/g, '"')
    .replace(/\s+/g, " ")
    .trim();
};

export const getObjectiveProgress = (objective: Objective) => {
  const total = objective.stats?.total ?? 0;
  const completed = objective.stats?.completed ?? 0;
  return total > 0 ? Math.round((completed / total) * 100) : 0;
};

const getKeyResultProgress = (keyResult: KeyResult) => {
  const range = keyResult.targetValue - keyResult.startValue;
  if (range === 0) {
    return keyResult.currentValue >= keyResult.targetValue ? 100 : 0;
  }

  return Math.max(
    0,
    Math.min(
      100,
      Math.round(
        ((keyResult.currentValue - keyResult.startValue) / range) * 100,
      ),
    ),
  );
};

const formatValue = (value: number, measurementType: string) => {
  if (measurementType === "percentage") return `${value}%`;
  if (measurementType === "boolean") {
    return value >= 1 ? "Complete" : "Incomplete";
  }
  return NUMBER_FORMAT.format(value);
};

const getProgressBadgeClassName = (progress: number) => {
  if (progress < 25) return "border-danger/20 bg-danger/10";
  if (progress < 50) return "border-warning/20 bg-warning/10";
  if (progress < 75) return "border-info/20 bg-info/10";
  return "border-success/20 bg-success/10";
};

const getProgressTrackClassName = (progress: number) => {
  if (progress === 0) return "[&_circle:first-child]:stroke-danger";
  if (progress < 25) return "[&_circle:first-child]:stroke-danger/25";
  if (progress < 50) return "[&_circle:first-child]:stroke-warning/25";
  if (progress < 75) return "[&_circle:first-child]:stroke-info/25";
  return "[&_circle:first-child]:stroke-success/25";
};

const cardClasses = cn(
  "border-border-strong/65 bg-white shadow-shadow dark:border-border-strong/80 dark:bg-surface-elevated/55",
  "rounded-[14px] border-2 shadow-lg backdrop-blur",
  "transition-[border-color,box-shadow,background-color] duration-150",
  "hover:border-foreground/35 hover:bg-surface-elevated hover:shadow-xl",
  "group-data-[dragging=true]/node:border-foreground/65 group-data-[dragging=true]/node:shadow-2xl",
);

const NodeEyebrow = ({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) => (
  <Text
    className={cn(
      "text-[0.68rem] font-medium tracking-[0.12em] uppercase",
      className,
    )}
    color="muted"
  >
    {children}
  </Text>
);

const ContextMenuLabel = ({
  children,
  icon,
}: {
  children: React.ReactNode;
  icon: React.ReactNode;
}) => (
  <Flex align="center" className="w-full gap-2">
    <span className="grid h-4 w-4 place-items-center text-current">{icon}</span>
    <Text>{children}</Text>
  </Flex>
);

const Metric = ({
  label,
  value,
}: {
  label: string;
  value: string | number;
}) => (
  <Box className="min-w-0 flex-1 text-center">
    <Text className="text-xl tabular-nums" fontWeight="semibold">
      {value}
    </Text>
    <Text
      className="mt-0.5 truncate text-[0.75rem] tracking-[0.1em] uppercase"
      color="muted"
    >
      {label}
    </Text>
  </Box>
);

export const UltimateGoalNodeCard = memo(
  ({
    averageProgress,
    canEdit,
    description,
    objectiveCount,
    onEdit,
    onOpenDetails,
    pillarCount,
    title,
  }: {
    averageProgress: number;
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
          <NodeEyebrow className="text-text-primary text-[0.74rem] font-bold">
            ◆ Ultimate goal
          </NodeEyebrow>
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
            <Metric label="Avg progress" value={`${averageProgress}%`} />
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
  ),
);
UltimateGoalNodeCard.displayName = "UltimateGoalNodeCard";

export const PillarNodeCard = memo(
  ({
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
              "border-foreground bg-surface-elevated ring-foreground/10 ring-4",
          )}
          style={{ width: PILLAR_NODE_WIDTH }}
        >
          <NodeEyebrow className="text-[0.74rem] font-semibold">
            Strategic pillar
          </NodeEyebrow>
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
  ),
);
PillarNodeCard.displayName = "PillarNodeCard";

const ObjectiveKeyResults = ({ objectiveId }: { objectiveId: string }) => {
  const { data: keyResults = [], isPending } = useKeyResults(objectiveId);
  const [isExpanded, setIsExpanded] = useState(false);

  if (isPending || keyResults.length === 0) return null;

  return (
    <Box className="border-border mt-4 border-t pt-3" data-no-drag>
      <button
        aria-expanded={isExpanded}
        className="text-text-muted hover:text-text-primary flex w-full items-center justify-between gap-1 rounded-md text-left text-[0.95rem] transition-colors"
        onClick={() => {
          setIsExpanded((current) => !current);
        }}
        type="button"
      >
        <span>
          {keyResults.length} key result{keyResults.length === 1 ? "" : "s"}
        </span>
        <ChevronRightIcon
          className={cn(
            "h-4 w-4 shrink-0 transition-transform duration-150",
            isExpanded && "rotate-90",
          )}
          strokeWidth={2}
        />
      </button>
      {isExpanded ? (
        <ul className="mt-3 space-y-4">
          {keyResults.map((keyResult) => {
            const progress = getKeyResultProgress(keyResult);
            return (
              <Tooltip
                className="min-w-44"
                delayDuration={300}
                key={keyResult.id}
                title={
                  <Box>
                    <Flex align="center" className="gap-4" justify="between">
                      <Text>Progress</Text>
                      <Text className="tabular-nums" fontWeight="semibold">
                        {progress}%
                      </Text>
                    </Flex>
                    <Text className="mt-1 text-[0.9rem]" color="muted">
                      {formatValue(
                        keyResult.currentValue,
                        keyResult.measurementType,
                      )}{" "}
                      of{" "}
                      {formatValue(
                        keyResult.targetValue,
                        keyResult.measurementType,
                      )}
                    </Text>
                  </Box>
                }
              >
                <li className="flex items-start gap-3">
                  <span
                    aria-label={`Progress ${progress}%`}
                    className="mt-0.5 shrink-0"
                  >
                    <CircleProgressBar
                      className={getProgressTrackClassName(progress)}
                      progress={progress}
                      size={16}
                      strokeWidth={2}
                    />
                  </span>
                  <Box className="min-w-0 flex-1">
                    <Text className="text-foreground line-clamp-2 text-base leading-5">
                      {keyResult.name}
                    </Text>
                  </Box>
                </li>
              </Tooltip>
            );
          })}
        </ul>
      ) : null}
    </Box>
  );
};

export const ObjectiveNodeCard = memo(
  ({
    canEdit,
    currentPillarId,
    objective,
    onAlign,
    onOpenDetails,
    onUpdate,
    pillars,
    status,
    statuses,
    teamCode,
  }: {
    canEdit: boolean;
    currentPillarId: string | null;
    objective: Objective;
    onAlign: (objectiveId: string, pillarId: string | null) => void;
    onOpenDetails: () => void;
    onUpdate: (objectiveId: string, data: ObjectiveUpdate) => void;
    pillars: StrategicPillar[];
    status?: ObjectiveStatus;
    statuses: ObjectiveStatus[];
    teamCode?: string;
  }) => {
    const { withWorkspace } = useWorkspacePath();
    const { data: members = [] } = useMembers();
    const progress = getObjectiveProgress(objective);
    const lead = members.find(({ id }) => id === objective.leadUser);
    const objectivePath = withWorkspace(
      `/teams/${objective.teamId}/objectives/${objective.id}`,
    );
    const objectiveReference = teamCode
      ? `${teamCode}-${objective.sequenceId}`
      : String(objective.sequenceId);
    const statusDotStyle: CSSProperties = {
      backgroundColor: status?.color ?? "var(--color-text-muted)",
    };

    return (
      <ContextMenu>
        <ContextMenu.Trigger>
          <Box
            className={cn(cardClasses, "px-4 py-4")}
            style={{ width: OBJECTIVE_NODE_WIDTH }}
          >
            <Flex align="start" className="gap-3" justify="between">
              <button
                className="focus-visible:ring-foreground/30 min-w-0 flex-1 rounded-sm text-left outline-none focus-visible:ring-1"
                data-card-select
                onClick={(event) => {
                  if (event.detail === 0) onOpenDetails();
                }}
                type="button"
              >
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
            <Flex align="center" className="mt-3 gap-1.5" data-no-drag wrap>
              <AssigneesMenu>
                <AssigneesMenu.Trigger>
                  <Button
                    aria-label={lead ? `Lead: ${lead.username}` : "Add lead"}
                    asIcon
                    className="gap-1 px-1"
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
                    className="gap-1 pr-2"
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
                    className="gap-1 pr-2"
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
              <Flex
                align="center"
                className={cn(
                  "h-[1.85rem] gap-1.5 rounded-xl border px-1.5",
                  getProgressBadgeClassName(progress),
                )}
              >
                <CircleProgressBar
                  className={getProgressTrackClassName(progress)}
                  progress={progress}
                  size={16}
                  strokeWidth={2}
                />
                <Text className="text-[0.95rem] tabular-nums">{progress}%</Text>
              </Flex>
              <ObjectiveHealthEditor
                health={objective.health}
                objectiveId={objective.id}
              >
                <Button
                  className="gap-1 px-1"
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
                    className="gap-1 pr-2"
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
            <ObjectiveKeyResults objectiveId={objective.id} />
          </Box>
        </ContextMenu.Trigger>
        <ContextMenu.Items className="w-64">
          <ContextMenu.Group>
            <ContextMenu.Item asChild>
              <Link href={objectivePath}>
                <ContextMenuLabel
                  icon={<ExternalLinkIcon className="h-4 w-4" />}
                >
                  Open objective
                </ContextMenuLabel>
              </Link>
            </ContextMenu.Item>
          </ContextMenu.Group>
          {canEdit ? (
            <>
              <ContextMenu.Separator />
              {statuses.length > 0 ? (
                <ContextMenu.Group>
                  <ContextMenu.SubMenu>
                    <ContextMenu.SubTrigger className="justify-between">
                      <Flex align="center" className="gap-2">
                        <span
                          aria-hidden
                          className="h-2 w-2 rounded-full"
                          style={statusDotStyle}
                        />
                        <Text>Set status</Text>
                      </Flex>
                      <ChevronRightIcon className="text-text-muted h-4 w-4" />
                    </ContextMenu.SubTrigger>
                    <ContextMenu.SubItems className="min-w-44">
                      <ContextMenu.Group>
                        {statuses.map((statusOption) => (
                          <ContextMenu.Item
                            active={statusOption.id === objective.statusId}
                            key={statusOption.id}
                            onSelect={() => {
                              onUpdate(objective.id, {
                                statusId: statusOption.id,
                              });
                            }}
                          >
                            <span
                              aria-hidden
                              className="h-2 w-2 rounded-full"
                              style={{ backgroundColor: statusOption.color }}
                            />
                            <Text>{statusOption.name}</Text>
                          </ContextMenu.Item>
                        ))}
                      </ContextMenu.Group>
                    </ContextMenu.SubItems>
                  </ContextMenu.SubMenu>
                </ContextMenu.Group>
              ) : null}
              {pillars.length > 0 ? (
                <ContextMenu.Group>
                  <ContextMenu.SubMenu>
                    <ContextMenu.SubTrigger className="justify-between">
                      <Flex align="center" className="gap-2">
                        <ObjectiveIcon className="text-text-secondary h-4 w-4" />
                        <Text>Align to pillar</Text>
                      </Flex>
                      <ChevronRightIcon className="text-text-muted h-4 w-4" />
                    </ContextMenu.SubTrigger>
                    <ContextMenu.SubItems className="max-w-64 min-w-48">
                      <ContextMenu.Group>
                        {pillars.map((pillar) => (
                          <ContextMenu.Item
                            active={pillar.id === currentPillarId}
                            key={pillar.id}
                            onSelect={() => {
                              onAlign(objective.id, pillar.id);
                            }}
                          >
                            <Text className="max-w-52 truncate">
                              {pillar.name}
                            </Text>
                          </ContextMenu.Item>
                        ))}
                      </ContextMenu.Group>
                    </ContextMenu.SubItems>
                  </ContextMenu.SubMenu>
                </ContextMenu.Group>
              ) : null}
              {currentPillarId ? (
                <>
                  <ContextMenu.Separator />
                  <ContextMenu.Group>
                    <ContextMenu.Item
                      className="whitespace-nowrap"
                      onSelect={() => {
                        onAlign(objective.id, null);
                      }}
                    >
                      <ContextMenuLabel
                        icon={<UnlinkIcon className="h-4 w-4" />}
                      >
                        Remove pillar alignment
                      </ContextMenuLabel>
                    </ContextMenu.Item>
                  </ContextMenu.Group>
                </>
              ) : null}
            </>
          ) : null}
        </ContextMenu.Items>
      </ContextMenu>
    );
  },
);
ObjectiveNodeCard.displayName = "ObjectiveNodeCard";
