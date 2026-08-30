import type { ComponentProps } from "react";
import { Avatar, Flex } from "ui";
import { EstimateIcon, TagsIcon } from "icons";
import { formatEstimate, type EstimateScheme } from "@/lib/estimate";
import { PriorityIcon } from "../priority-icon";
import { StoryStatusIcon } from "../story-status-icon";
import { getPluralLabel } from "./filter-model";
import type { LabelChipSummary, UserChipSummary } from "./types";

type StoryPriority = NonNullable<
  ComponentProps<typeof PriorityIcon>["priority"]
>;

export const PeopleChipValue = ({
  label,
  pluralLabel,
  users,
}: {
  label: string;
  pluralLabel: string;
  users: UserChipSummary[];
}) => {
  const visibleUsers = users.slice(0, 2);

  if (users.length > 2) {
    return (
      <Flex align="center" gap={1}>
        <Flex align="center" className="-space-x-1">
          {visibleUsers.map((user) => (
            <Avatar
              className="ring-background ring-1"
              color="primary"
              key={user.id}
              name={user.name}
              size="xs"
              src={user.avatarUrl}
            />
          ))}
        </Flex>
        <span>{getPluralLabel(users.length, label, pluralLabel)}</span>
      </Flex>
    );
  }

  return (
    <Flex align="center" gap={2}>
      {visibleUsers.map((user) => (
        <Flex align="center" gap={1} key={user.id}>
          <Avatar
            color="primary"
            name={user.name}
            size="xs"
            src={user.avatarUrl}
          />
          <span>{user.username}</span>
        </Flex>
      ))}
    </Flex>
  );
};

type StatusChipSummary = {
  id: string;
  name: string;
};

export const StatusChipValue = ({
  statuses,
}: {
  statuses: StatusChipSummary[];
}) => {
  const visibleStatuses = statuses.slice(0, 2);

  if (statuses.length > 2) {
    return (
      <Flex align="center" gap={1}>
        <Flex align="center" className="-space-x-0.5">
          {visibleStatuses.map((status) => (
            <StoryStatusIcon
              className="ring-background size-3 ring-1"
              key={status.id}
              statusId={status.id}
            />
          ))}
        </Flex>
        <span>{getPluralLabel(statuses.length, "status", "statuses")}</span>
      </Flex>
    );
  }

  return (
    <Flex align="center" gap={2}>
      {visibleStatuses.map((status) => (
        <Flex align="center" gap={1} key={status.id}>
          <StoryStatusIcon statusId={status.id} />
          <span>{status.name}</span>
        </Flex>
      ))}
    </Flex>
  );
};

export const PriorityChipValue = ({
  priorities,
}: {
  priorities: StoryPriority[];
}) => {
  const visiblePriorities = priorities.slice(0, 2);

  if (priorities.length > 2) {
    return (
      <Flex align="center" gap={1}>
        <PriorityIcon priority="High" />
        <span>
          {getPluralLabel(priorities.length, "priority", "priorities")}
        </span>
      </Flex>
    );
  }

  return (
    <Flex align="center" gap={2}>
      {visiblePriorities.map((priority) => (
        <Flex align="center" gap={1} key={priority}>
          <PriorityIcon priority={priority} />
          <span>{priority}</span>
        </Flex>
      ))}
    </Flex>
  );
};

export const LabelChipValue = ({ labels }: { labels: LabelChipSummary[] }) => {
  const visibleLabels = labels.slice(0, 2);

  if (labels.length > 2) {
    return (
      <Flex align="center" gap={1}>
        <TagsIcon
          className="h-4 w-auto"
          style={{ color: visibleLabels[0]?.color }}
        />
        <span>{getPluralLabel(labels.length, "label", "labels")}</span>
      </Flex>
    );
  }

  return (
    <Flex align="center" gap={2}>
      {visibleLabels.map((label) => (
        <Flex align="center" gap={1} key={label.id}>
          <TagsIcon className="h-4 w-auto" style={{ color: label.color }} />
          <span>{label.name}</span>
        </Flex>
      ))}
    </Flex>
  );
};

export const EstimateChipValue = ({
  estimateScheme,
  estimateValues,
}: {
  estimateScheme: EstimateScheme;
  estimateValues: number[];
}) => {
  const visibleValues = estimateValues.slice(0, 2);

  if (estimateValues.length > 2) {
    return (
      <Flex align="center" gap={1}>
        <EstimateIcon className="h-4 w-auto" />
        <span>
          {getPluralLabel(
            estimateValues.length,
            "complexity value",
            "complexity values",
          )}
        </span>
      </Flex>
    );
  }

  return (
    <Flex align="center" gap={2}>
      {visibleValues.map((estimateValue) => (
        <Flex align="center" gap={1} key={estimateValue}>
          <EstimateIcon className="h-4 w-auto" />
          <span>{formatEstimate(estimateScheme, estimateValue, "full")}</span>
        </Flex>
      ))}
    </Flex>
  );
};
