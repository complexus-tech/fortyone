import type { ReactNode } from "react";
import { format } from "date-fns";
import { Box, Flex, Text, Avatar, TimeAgo, Tooltip, Button, Badge } from "ui";
import Link from "next/link";
import { cn } from "lib";
import {
  CalendarIcon,
  EstimateIcon,
  InfoIcon,
  SprintsIcon,
  TagsIcon,
} from "icons";
import { formatActivityReasonDates } from "@/lib/activity-format";
import { DEFAULT_ESTIMATE_SCHEME, formatEstimate } from "@/lib/estimate";
import { useTerminology, useWorkspacePath } from "@/hooks";
import { useLabels } from "@/lib/hooks/labels";
import { useMayaAssignee, useMembers } from "@/lib/hooks/members";
import { useStatuses } from "@/lib/hooks/statuses";
import { useObjective } from "@/modules/objectives/hooks/use-objective";
import { useSprint } from "@/modules/sprints/hooks/sprint-details";
import type { StoryActivity, StoryPriority } from "@/modules/stories/types";
import { useTeamSettings } from "@/modules/teams/hooks/use-team-settings";
import type { Label } from "@/types";
import { getActivityCopy, getDisplayActivityReason } from "./activity-copy";
import { MayaAvatar } from "./maya-avatar";
import { PriorityIcon } from "./priority-icon";
import { StoryStatusIcon } from "./story-status-icon";

const DisplayEstimate = ({
  value,
  teamId,
}: {
  value: string;
  teamId?: string;
}) => {
  const { data: teamSettings } = useTeamSettings(teamId);
  const estimateValue = Number.parseInt(value, 10);
  const estimateScheme =
    teamSettings?.estimationSettings.scheme ?? DEFAULT_ESTIMATE_SCHEME;

  return (
    <span className="flex items-center gap-1">
      <EstimateIcon className="h-5" />
      {Number.isNaN(estimateValue)
        ? "No estimate"
        : formatEstimate(estimateScheme, estimateValue, "full")}
    </span>
  );
};

const DisplaySprint = ({
  sprintId,
  teamId,
}: {
  sprintId: string;
  teamId?: string;
}) => {
  const { data: sprint } = useSprint(sprintId, teamId);
  const { withWorkspace } = useWorkspacePath();
  return (
    <>
      {!sprintId || sprintId.includes("nil") ? (
        <span>No sprint</span>
      ) : (
        <Link
          className="flex items-center gap-1"
          href={withWorkspace(
            `/teams/${sprint?.teamId}/sprints/${sprintId}/stories`,
          )}
        >
          <SprintsIcon className="h-5" />
          {sprint?.name}
        </Link>
      )}
    </>
  );
};

const DisplayObjective = ({
  objectiveId,
  teamId,
}: {
  objectiveId: string;
  teamId?: string;
}) => {
  const { data: objective } = useObjective(objectiveId, teamId);
  const { withWorkspace } = useWorkspacePath();
  return (
    <>
      {!objectiveId || objectiveId.includes("nil") ? (
        <span>No objective</span>
      ) : (
        <Link
          href={withWorkspace(
            `/teams/${objective?.teamId}/objectives/${objectiveId}`,
          )}
        >
          {objective?.name}
        </Link>
      )}
    </>
  );
};

const ASSOCIATION_ACTIVITY_FIELDS = new Set([
  "blocked_by_id",
  "blocking_id",
  "related_id",
  "duplicate_id",
  "duplicated_by_id",
]);

const getAssociationBadgeColor = (
  field: string,
): "danger" | "tertiary" | "warning" => {
  if (field === "blocking_id") return "warning";
  if (field === "blocked_by_id") return "danger";
  return "tertiary";
};

const AssociationActivityBadge = ({
  field,
  label,
}: {
  field: string;
  label: string;
}) => (
  <Badge
    className="shrink-0 px-2 text-[0.75rem] font-semibold uppercase"
    color={getAssociationBadgeColor(field)}
    rounded="sm"
  >
    {label}
  </Badge>
);

const getActivityLabelIds = (value: unknown): string[] => {
  if (!Array.isArray(value)) return [];
  return value.filter(
    (labelId): labelId is string => typeof labelId === "string",
  );
};

const getLabelActivityDisplayValue = (labels: Label[]) => {
  if (labels.length === 1) return labels[0].name;
  return `${labels.length} labels`;
};

const ActivityLabelValue = ({ labels }: { labels: Label[] }) => {
  if (labels.length === 0) {
    return <span>No labels</span>;
  }

  const firstLabel = labels[0];
  const tooltip = (
    <Flex className="min-w-28" direction="column" gap={2}>
      {labels.map((label) => (
        <Flex align="center" gap={1} key={label.id}>
          <TagsIcon className="h-4" style={{ color: label.color }} />
          {label.name}
        </Flex>
      ))}
    </Flex>
  );

  return (
    <Tooltip title={labels.length > 1 ? tooltip : null}>
      <Badge
        className="h-6 shrink-0 gap-1.5 px-2 text-[0.85rem]"
        color="tertiary"
        rounded="xl"
        variant="outline"
      >
        <TagsIcon className="h-4" style={{ color: firstLabel.color }} />
        <span className="inline-block max-w-[12ch] truncate">
          {labels.length === 1 ? firstLabel.name : `${labels.length} labels`}
        </span>
      </Badge>
    </Tooltip>
  );
};

const getActivityVerb = (type: StoryActivity["type"], storyTerm: string) => {
  if (type === "create") return `created the ${storyTerm}`;
  if (type === "link") return "linked";
  return "changed";
};

export const Activity = ({
  avatarSurfaceClassName,
  teamId,
  field,
  currentValue,
  type,
  createdAt,
  user,
  newValue,
  oldValue,
  reason,
}: StoryActivity & { avatarSurfaceClassName?: string; teamId?: string }) => {
  const { data: members = [] } = useMembers();
  const { data: mayaAssignee } = useMayaAssignee();
  const { data: statuses = [] } = useStatuses();
  const { data: allLabels = [] } = useLabels();
  const { withWorkspace } = useWorkspacePath();
  const { getTermDisplay } = useTerminology();
  const storyTerm = getTermDisplay("storyTerm");
  const member = user;
  const activityAssignees = mayaAssignee
    ? [...members.filter(({ id }) => id !== mayaAssignee.id), mayaAssignee]
    : members;
  const findActivityAssignee = (value: string) =>
    activityAssignees.find(({ id }) => id === value);
  const activityVerb = getActivityVerb(type, storyTerm);
  const activityReason = formatActivityReasonDates(
    getDisplayActivityReason(reason),
  );
  const isLinkedUrl =
    type === "link" &&
    currentValue &&
    typeof newValue === "string" &&
    newValue.startsWith("http");
  let linkedValue: ReactNode = null;

  if (type === "link" && currentValue) {
    if (isLinkedUrl && typeof newValue === "string") {
      linkedValue = (
        <a
          className="inline-block shrink-0 text-sm text-black underline md:text-[0.95rem] dark:text-white"
          href={newValue}
          rel="noopener noreferrer"
          target="_blank"
        >
          {currentValue}
        </a>
      );
    } else {
      linkedValue = (
        <Text
          as="span"
          className="inline-block shrink-0 text-sm text-black md:text-[0.95rem] dark:text-white"
          fontWeight="medium"
        >
          {currentValue}
        </Text>
      );
    }
  }

  if (field === "completed_at") {
    return null;
  }

  const fieldMap = {
    title: {
      label: "Title",
      render: (value: string) => <span>{value}</span>,
    },
    description: {
      label: "Description",
      render: (value: string) => (
        <span>{value.length > 50 ? `${value.slice(0, 50)}...` : value}</span>
      ),
    },
    status_id: {
      label: "Status",
      render: (value: string) => (
        <span className="flex items-center gap-1">
          <StoryStatusIcon className="size-3" statusId={value} />
          {statuses.find((status) => status.id === value)?.name}
        </span>
      ),
    },
    priority: {
      label: "Priority",
      render: (value: string) => (
        <span className="flex items-center gap-1">
          <PriorityIcon className="h-5" priority={value as StoryPriority} />
          {value}
        </span>
      ),
    },
    estimate_unit: {
      label: "Estimate",
      render: (value: string) => (
        <DisplayEstimate teamId={teamId} value={value} />
      ),
    },
    assignee_id: {
      label: "Assignee",
      render: (value: string) => {
        const assignee = findActivityAssignee(value);
        const assigneeLabel =
          assignee?.username || assignee?.fullName || "Unknown user";

        if (!value || value.includes("nil")) {
          return <span>Unassigned</span>;
        }

        const content = (
          <>
            {assignee?.isSystem ? (
              <MayaAvatar
                className="relative top-px"
                name={assignee.fullName || assigneeLabel}
                size="xs"
                src={assignee.avatarUrl}
              />
            ) : (
              <Avatar
                className="relative top-px"
                name={assignee?.fullName || assigneeLabel}
                size="xs"
                src={assignee?.avatarUrl}
              />
            )}
            {assigneeLabel}
          </>
        );

        if (!assignee || assignee.isSystem) {
          return (
            <span className="flex items-center gap-1.5 pb-0.5">{content}</span>
          );
        }

        return (
          <Link
            className="flex items-center gap-1.5 pb-0.5"
            href={withWorkspace(`/profile/${assignee.id}`)}
          >
            {content}
          </Link>
        );
      },
    },
    start_date: {
      label: "Start date",
      render: (value: string) => (
        <span className="flex items-center gap-1">
          <CalendarIcon className="h-[1.15rem]" />
          {value
            ? format(new Date(value.split(" ")[0]), "PP")
            : "No start date"}
        </span>
      ),
    },
    end_date: {
      label: "Deadline",
      render: (value: string) => (
        <span className="flex items-center gap-1">
          <CalendarIcon className="h-[1.15rem]" />
          {value ? format(new Date(value.split(" ")[0]), "PP") : "No deadline"}
        </span>
      ),
    },
    sprint_id: {
      label: "Sprint",
      render: (value: string) => (
        <DisplaySprint sprintId={value} teamId={teamId} />
      ),
    },
    epic_id: {
      label: "Epic",
      render: (value: string) => <span>{value}</span>,
    },
    objective_id: {
      label: "Objective",
      render: (value: string) => (
        <DisplayObjective objectiveId={value} teamId={teamId} />
      ),
    },
    blocked_by_id: {
      label: "Blocked by",
      render: (value: string) => <span>{value}</span>,
    },
    blocking_id: {
      label: "Blocking",
      render: (value: string) => <span>{value}</span>,
    },
    related_id: {
      label: "Related to",
      render: (value: string) => <span>{value}</span>,
    },
    duplicate_id: {
      label: "Duplicate of",
      render: (value: string) => <span>{value}</span>,
    },
    duplicated_by_id: {
      label: "Duplicated by",
      render: (value: string) => <span>{value}</span>,
    },
    labels: {
      label: "Labels",
      render: (value: string) => <span>{value}</span>,
    },
  } as Record<
    string,
    {
      label: string;
      icon?: ReactNode;
      render: (value: string) => ReactNode;
    }
  >;

  const fieldMeta = fieldMap[field] ?? {
    label: field,
    render: (value: string) => <span>{value}</span>,
  };
  const activityLabels =
    field === "labels"
      ? getActivityLabelIds(newValue)
          .map((labelId) => allLabels.find((label) => label.id === labelId))
          .filter((label): label is Label => Boolean(label))
      : [];
  const displayCurrentValue =
    field === "labels" && activityLabels.length > 0
      ? getLabelActivityDisplayValue(activityLabels)
      : currentValue;
  const activityCopy = getActivityCopy({
    currentValue: displayCurrentValue,
    field,
    fieldLabel: fieldMeta.label,
    oldValue,
    reason,
    storyTerm,
    type,
  });
  const renderUpdateSegment = (
    segment: (typeof activityCopy.segments)[number],
  ) => {
    if (segment.type === "text") {
      return (
        <Text
          as="span"
          className="text-sm md:text-[0.95rem]"
          color="muted"
          key={`text-${segment.text}`}
        >
          {segment.text}
        </Text>
      );
    }

    return (
      <Text
        as="span"
        className="inline-block shrink-0 text-sm text-black md:text-[0.95rem] dark:text-white"
        fontWeight="medium"
        key={
          segment.type === "oldValue"
            ? `oldValue-${segment.value}`
            : segment.type
        }
      >
        {(() => {
          if (
            segment.type === "fieldLabel" &&
            ASSOCIATION_ACTIVITY_FIELDS.has(field)
          ) {
            return (
              <AssociationActivityBadge field={field} label={fieldMeta.label} />
            );
          }

          if (
            segment.type === "oldValue" &&
            ASSOCIATION_ACTIVITY_FIELDS.has(field)
          ) {
            return (
              <AssociationActivityBadge field={field} label={segment.value} />
            );
          }

          if (segment.type === "fieldLabel") {
            return fieldMeta.label;
          }

          if (segment.type === "currentValue" && field === "labels") {
            return <ActivityLabelValue labels={activityLabels} />;
          }

          return fieldMeta.render(
            segment.type === "oldValue" ? segment.value : currentValue,
          );
        })()}
      </Text>
    );
  };

  return (
    <Box className="relative pb-2 last-of-type:pb-0 md:pb-4">
      <Box
        className={cn(
          "border-border pointer-events-none absolute top-0 left-4 z-0 h-full border-l border-dashed",
        )}
      />
      <Flex align="center" className="z-1" gap={1}>
        <Tooltip
          className="py-2.5"
          title={
            <Box>
              <Flex gap={2}>
                {member.isSystem ? (
                  <MayaAvatar
                    className="mt-0.5"
                    name={member.fullName}
                    size="md"
                    src={member.avatarUrl}
                  />
                ) : (
                  <Avatar
                    className="mt-0.5"
                    name={member.fullName}
                    src={member.avatarUrl}
                  />
                )}
                <Box>
                  <Link
                    className={cn("mb-2 flex gap-1", {
                      "mb-0": member.isSystem,
                    })}
                    href={member.isSystem ? "" : `/profile/${member.id}`}
                  >
                    <Text fontSize="md" fontWeight="medium">
                      {member.fullName}
                    </Text>
                    <Text color="muted" fontSize="md">
                      ({member.username})
                    </Text>
                  </Link>
                  {!member.isSystem ? (
                    <Button
                      className="mb-0.5 ml-px px-2"
                      color="tertiary"
                      href={withWorkspace(`/profile/${member.id}`)}
                      size="xs"
                    >
                      Go to profile
                    </Button>
                  ) : (
                    <Text color="muted" fontSize="md">
                      ({member.username === "maya" ? "AI Assistant" : "Bot"})
                    </Text>
                  )}
                </Box>
              </Flex>
            </Box>
          }
        >
          <Flex align="center" className="cursor-pointer" gap={1}>
            <Box
              className={cn(
                "bg-surface relative left-px flex aspect-square items-center rounded-full p-[0.3rem]",
                avatarSurfaceClassName,
              )}
            >
              {member.isSystem ? (
                <MayaAvatar
                  name={member.fullName}
                  size="xs"
                  src={member.avatarUrl}
                />
              ) : (
                <Avatar
                  name={member.fullName}
                  size="xs"
                  src={member.avatarUrl}
                />
              )}
            </Box>
            <Text
              className="relative ml-1 text-sm text-black md:text-[0.95rem] dark:text-white"
              fontWeight="medium"
            >
              {member.username}
            </Text>
          </Flex>
        </Tooltip>
        <Box className="line-clamp-1 flex items-center gap-1 text-sm md:text-[0.95rem]">
          {type === "update" ? (
            activityCopy.segments.map(renderUpdateSegment)
          ) : (
            <Text as="span" className="text-sm md:text-[0.95rem]" color="muted">
              {activityVerb}
            </Text>
          )}
          {linkedValue}
          {activityReason ? (
            <Tooltip title={activityReason}>
              <span className="inline-flex shrink-0 cursor-help items-center">
                <InfoIcon className="text-icon-muted h-4" />
              </span>
            </Tooltip>
          ) : null}
          <Text
            as="span"
            className="mx-0.5 text-sm md:text-[0.95rem]"
            color="muted"
          >
            ·
          </Text>
          <Text
            as="span"
            className="shrink-0 text-sm md:text-[0.95rem]"
            color="muted"
          >
            <TimeAgo timestamp={createdAt} />
          </Text>
        </Box>
      </Flex>
    </Box>
  );
};
