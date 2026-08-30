import type { ReactNode } from "react";
import { format } from "date-fns";
import Link from "next/link";
import { Avatar, Badge, Flex, Tooltip } from "ui";
import { CalendarIcon, TagsIcon, TimeScheduleIcon } from "icons";
import { MayaAvatar } from "./maya-avatar";
import { StoryStatusIcon } from "./story-status-icon";

export type ActivityFieldMeta = {
  label: string;
  render: (value: string) => ReactNode;
};

export type ActivityLabel = {
  color: string;
  id: string;
  name: string;
};

type ActivityAssignee = {
  avatarUrl?: string | null;
  fullName?: string | null;
  id: string;
  isSystem?: boolean;
  username?: string | null;
};

type ActivityStatus = {
  id: string;
  name: string;
};

type ActivityFieldRendererOptions = {
  activityAssignees: ActivityAssignee[];
  renderEstimate: (value: string) => ReactNode;
  renderObjective: (value: string) => ReactNode;
  renderPriority: (value: string) => ReactNode;
  renderSprint: (value: string) => ReactNode;
  renderTimeNeeded: (value: string) => ReactNode;
  statuses: ActivityStatus[];
  withWorkspace: (path: string) => string;
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

export const isAssociationActivityField = (field: string) =>
  ASSOCIATION_ACTIVITY_FIELDS.has(field);

export const ActivityAssociationBadge = ({
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

export const getLabelActivityDisplayValue = (labels: ActivityLabel[]) => {
  if (labels.length === 1) return labels[0].name;
  return `${labels.length} labels`;
};

export const ActivityLabelValue = ({ labels }: { labels: ActivityLabel[] }) => {
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

const createActivityFieldRenderers = ({
  activityAssignees,
  renderEstimate,
  renderObjective,
  renderPriority,
  renderSprint,
  renderTimeNeeded,
  statuses,
  withWorkspace,
}: ActivityFieldRendererOptions): Partial<
  Record<string, ActivityFieldMeta>
> => {
  const findActivityAssignee = (value: string) =>
    activityAssignees.find(({ id }) => id === value);

  return {
    title: {
      label: "Title",
      render: (value) => <span>{value}</span>,
    },
    description: {
      label: "Description",
      render: (value) => (
        <span>{value.length > 50 ? `${value.slice(0, 50)}...` : value}</span>
      ),
    },
    status_id: {
      label: "Status",
      render: (value) => (
        <span className="flex items-center gap-1">
          <StoryStatusIcon className="size-3" statusId={value} />
          {statuses.find((status) => status.id === value)?.name}
        </span>
      ),
    },
    priority: {
      label: "Priority",
      render: renderPriority,
    },
    estimate_unit: {
      label: "Complexity",
      render: renderEstimate,
    },
    estimated_duration_minutes: {
      label: "Time needed",
      render: renderTimeNeeded,
    },
    minimum_focus_block_minutes: {
      label: "Minimum focus block",
      render: renderTimeNeeded,
    },
    auto_scheduling_status: {
      label: "Auto-scheduling",
      render: (value) => (
        <span className="flex items-center gap-1">
          <TimeScheduleIcon className="h-4" />
          {value}
        </span>
      ),
    },
    auto_scheduling_time: {
      label: "Scheduled time",
      render: (value) => (
        <span className="flex items-center gap-1">
          <CalendarIcon className="h-4" />
          {value}
        </span>
      ),
    },
    auto_scheduling_locked: {
      label: "Schedule lock",
      render: (value) => <span>{value}</span>,
    },
    auto_scheduling_enabled: {
      label: "Auto-scheduling",
      render: (value) => <span>{value}</span>,
    },
    assignee_id: {
      label: "Assignee",
      render: (value) => {
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
    collaborator_ids: {
      label: "Collaborators",
      render: (value) => <span>{value}</span>,
    },
    start_date: {
      label: "Start date",
      render: (value) => (
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
      render: (value) => (
        <span className="flex items-center gap-1">
          <CalendarIcon className="h-[1.15rem]" />
          {value ? format(new Date(value.split(" ")[0]), "PP") : "No deadline"}
        </span>
      ),
    },
    sprint_id: {
      label: "Sprint",
      render: renderSprint,
    },
    epic_id: {
      label: "Epic",
      render: (value) => <span>{value}</span>,
    },
    objective_id: {
      label: "Objective",
      render: renderObjective,
    },
    key_result_id: {
      label: "Key result",
      render: (value) => (
        <span>{!value || value.includes("nil") ? "No key result" : value}</span>
      ),
    },
    blocked_by_id: {
      label: "Blocked by",
      render: (value) => <span>{value}</span>,
    },
    blocking_id: {
      label: "Blocking",
      render: (value) => <span>{value}</span>,
    },
    related_id: {
      label: "Related to",
      render: (value) => <span>{value}</span>,
    },
    duplicate_id: {
      label: "Duplicate of",
      render: (value) => <span>{value}</span>,
    },
    duplicated_by_id: {
      label: "Duplicated by",
      render: (value) => <span>{value}</span>,
    },
    labels: {
      label: "Labels",
      render: (value) => <span>{value}</span>,
    },
  };
};

const DEFAULT_ACTIVITY_FIELD_META: ActivityFieldMeta = {
  label: "",
  render: (value) => <span>{value}</span>,
};

export const getActivityFieldMeta = (
  field: string,
  options: ActivityFieldRendererOptions,
): ActivityFieldMeta => {
  const fieldMeta = createActivityFieldRenderers(options)[field];
  return fieldMeta ?? { ...DEFAULT_ACTIVITY_FIELD_META, label: field };
};
