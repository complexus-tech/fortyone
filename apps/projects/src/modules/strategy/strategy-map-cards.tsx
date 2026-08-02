"use client";

import type { CSSProperties } from "react";
import { memo } from "react";
import { format, formatISO } from "date-fns";
import Link from "next/link";
import {
  CalendarIcon,
  ChevronRightIcon,
  DeleteIcon,
  EditIcon,
  ExternalLinkIcon,
  HelpIcon,
  ObjectiveIcon,
  OKRIcon,
  UnlinkIcon,
} from "icons";
import { cn } from "lib";
import {
  Avatar,
  Box,
  Button,
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
import { getKeyResultProgress } from "@/modules/key-results/utils";
import { ObjectiveHealthEditor } from "@/modules/objectives/components/objective-health-editor";
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
  KEY_RESULT_NODE_WIDTH,
  OBJECTIVE_NODE_WIDTH,
  PILLAR_NODE_WIDTH,
} from "./strategy-map-layout";

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

export const getObjectiveProgress = (
  objective: Objective,
  keyResults: KeyResult[] = [],
) => {
  if (objective.keyResultCount > 0 || keyResults.length > 0) {
    if (keyResults.length === 0) return 0;
    return Math.round(
      keyResults.reduce(
        (total, keyResult) => total + getKeyResultProgress(keyResult),
        0,
      ) / keyResults.length,
    );
  }

  const total = objective.stats?.total ?? 0;
  const completed = objective.stats?.completed ?? 0;
  return total > 0 ? Math.round((completed / total) * 100) : 0;
};

const getProgressFillClassName = (progress: number) => {
  if (progress < 25) return "bg-danger";
  if (progress < 50) return "bg-warning";
  if (progress < 75) return "bg-info";
  return "bg-success";
};

const StrategyProgressBar = ({
  className,
  progress,
}: {
  className?: string;
  progress: number;
}) => (
  <Flex align="center" className={cn("gap-2.5", className)}>
    <div
      aria-label={`Progress ${progress}%`}
      aria-valuemax={100}
      aria-valuemin={0}
      aria-valuenow={progress}
      className="bg-border h-1.5 min-w-0 flex-1 overflow-hidden rounded-full"
      role="progressbar"
    >
      <div
        className={cn(
          "h-full rounded-full",
          getProgressFillClassName(progress),
        )}
        style={{ width: `${progress}%` }}
      />
    </div>
    <Text className="shrink-0 text-sm tabular-nums" color="muted">
      {progress}%
    </Text>
  </Flex>
);

const cardClasses = cn(
  "border-border-strong/65 bg-white shadow-shadow dark:border-foreground/20 dark:bg-accent/70",
  "rounded-[14px] border-2 shadow-lg backdrop-blur",
  "transition-[border-color,box-shadow,background-color] duration-150",
  "hover:border-foreground/35 hover:bg-surface-elevated hover:shadow-xl dark:hover:border-foreground/45 dark:hover:bg-accent/70",
  "group-data-[dragging=true]/node:border-foreground/65 group-data-[dragging=true]/node:shadow-2xl",
);

const objectivePropertyControlClasses =
  "dark:border-foreground/15 dark:bg-state-hover! dark:hover:bg-state-active!";

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

const StrategyConceptInfo = ({
  description,
  label,
}: {
  description: string;
  label: string;
}) => (
  <Tooltip
    className="border-border-strong dark:border-border-strong dark:bg-surface-elevated max-w-72 py-2.5"
    delayDuration={250}
    title={description}
  >
    <button
      aria-label={`About ${label}`}
      className="text-text-muted hover:text-foreground inline-grid shrink-0 place-items-center rounded-sm transition-colors"
      data-no-drag
      onClick={(event) => {
        event.stopPropagation();
      }}
      type="button"
    >
      <HelpIcon className="h-4 w-4 text-current" />
    </button>
  </Tooltip>
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
  ),
);
PillarNodeCard.displayName = "PillarNodeCard";

export const KeyResultNodeCard = memo(
  ({
    code,
    keyResult,
    onOpenDetails,
  }: {
    code: string;
    keyResult: KeyResult;
    onOpenDetails: () => void;
  }) => {
    const { data: members = [] } = useMembers();
    const progress = getKeyResultProgress(keyResult);
    const lead = keyResult.lead
      ? members.find(({ id }) => id === keyResult.lead)
      : undefined;

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
  },
);
KeyResultNodeCard.displayName = "KeyResultNodeCard";

export const ObjectiveNodeCard = memo(
  ({
    canEdit,
    currentPillarId,
    isKeyResultsExpanded,
    keyResultCount,
    keyResults,
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
    const { data: members = [] } = useMembers();
    const progress = getObjectiveProgress(objective, keyResults);
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
                    className={cn(
                      objectivePropertyControlClasses,
                      "gap-1 px-1",
                    )}
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
                    className={cn(
                      objectivePropertyControlClasses,
                      "gap-1 pr-2",
                    )}
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
                    className={cn(
                      objectivePropertyControlClasses,
                      "gap-1 pr-2",
                    )}
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
