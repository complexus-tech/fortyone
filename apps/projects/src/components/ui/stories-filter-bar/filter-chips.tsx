import type { ComponentProps } from "react";
import { format } from "date-fns";
import {
  CalendarIcon,
  EstimateIcon,
  ListIcon,
  ObjectiveIcon,
  OKRIcon,
  TagsIcon,
} from "icons";
import type { EstimateScheme } from "@/lib/estimate";
import { PriorityIcon } from "../priority-icon";
import { StoryStatusIcon } from "../story-status-icon";
import { TeamColor } from "../team-color";
import type { StoriesFilter } from "../stories-filter-types";
import {
  EstimateChipValue,
  LabelChipValue,
  PeopleChipValue,
  PriorityChipValue,
  StatusChipValue,
} from "./chip-values";
import { getNames, getOperatorConfig } from "./filter-model";
import type {
  FilterChip,
  LabelChipSummary,
  StoriesFilterField,
  UserChipSummary,
} from "./types";

type StoryPriority = NonNullable<
  ComponentProps<typeof PriorityIcon>["priority"]
>;

type GetFilterTermDisplay = (
  term: "keyResultTerm" | "objectiveTerm",
  options: { capitalize: true },
) => string;

type BuildFilterChipsOptions = {
  estimateScheme: EstimateScheme;
  filters: StoriesFilter;
  getTermDisplay: GetFilterTermDisplay;
  hiddenFields: ReadonlySet<StoriesFilterField>;
  keyResults: ReadonlyMap<string, string>;
  labels: ReadonlyMap<string, LabelChipSummary>;
  objectives: ReadonlyMap<string, string>;
  sprints: ReadonlyMap<string, string>;
  statuses: ReadonlyMap<string, string>;
  teamColors: ReadonlyMap<string, string>;
  teams: ReadonlyMap<string, string>;
  users: ReadonlyMap<string, UserChipSummary>;
};

export const buildFilterChips = ({
  estimateScheme,
  filters,
  getTermDisplay,
  hiddenFields,
  keyResults,
  labels,
  objectives,
  sprints,
  statuses,
  teamColors,
  teams,
  users,
}: BuildFilterChipsOptions) => {
  const items: FilterChip[] = [];

  if (filters.contentContains?.trim()) {
    items.push({
      field: "contentContains",
      icon: <ListIcon className="h-4 w-auto" />,
      label: "Content",
      ...getOperatorConfig(filters, "contentContains"),
      value: filters.contentContains.trim(),
    });
  }

  if (filters.startDate) {
    items.push({
      field: "startDate",
      icon: <CalendarIcon className="h-4 w-auto" />,
      label: "Start date",
      ...getOperatorConfig(filters, "startDate"),
      value: format(new Date(filters.startDate), "MMM d, yyyy"),
    });
  }

  if (filters.endDate) {
    items.push({
      field: "endDate",
      icon: <CalendarIcon className="h-4 w-auto" />,
      label: "End date",
      ...getOperatorConfig(filters, "endDate"),
      value: format(new Date(filters.endDate), "MMM d, yyyy"),
    });
  }

  if (filters.statusIds?.length) {
    const selectedStatuses = filters.statusIds.map((id) => ({
      id,
      name: statuses.get(id) ?? id,
    }));
    items.push({
      field: "statusIds",
      icon: <StoryStatusIcon statusId={filters.statusIds[0]} />,
      label: "Status",
      ...getOperatorConfig(filters, "statusIds"),
      value: <StatusChipValue statuses={selectedStatuses} />,
    });
  }

  if (filters.assigneeIds?.length) {
    const selectedUsers = filters.assigneeIds
      .map((id) => users.get(id))
      .filter((user): user is UserChipSummary => Boolean(user));
    items.push({
      field: "assigneeIds",
      label: "Assignee",
      ...getOperatorConfig(filters, "assigneeIds"),
      value: (
        <PeopleChipValue
          label="assignee"
          pluralLabel="assignees"
          users={selectedUsers}
        />
      ),
    });
  }

  if (filters.reporterIds?.length) {
    const selectedUsers = filters.reporterIds
      .map((id) => users.get(id))
      .filter((user): user is UserChipSummary => Boolean(user));
    items.push({
      field: "reporterIds",
      label: "Creator",
      ...getOperatorConfig(filters, "reporterIds"),
      value: (
        <PeopleChipValue
          label="creator"
          pluralLabel="creators"
          users={selectedUsers}
        />
      ),
    });
  }

  if (filters.priorities?.length) {
    const selectedPriorities = filters.priorities as StoryPriority[];
    items.push({
      field: "priorities",
      icon: <PriorityIcon priority={selectedPriorities[0]} />,
      label: "Priority",
      ...getOperatorConfig(filters, "priorities"),
      value: <PriorityChipValue priorities={selectedPriorities} />,
    });
  }

  if (filters.teamIds?.length) {
    items.push({
      field: "teamIds",
      icon: <TeamColor color={teamColors.get(filters.teamIds[0])} />,
      label: "Team",
      ...getOperatorConfig(filters, "teamIds"),
      value: getNames(filters.teamIds, teams),
    });
  }

  if (filters.sprintIds?.length) {
    items.push({
      field: "sprintIds",
      label: "Sprint",
      ...getOperatorConfig(filters, "sprintIds"),
      value: getNames(filters.sprintIds, sprints),
    });
  }

  if (filters.labelIds?.length) {
    const selectedLabels = filters.labelIds
      .map((id) => labels.get(id))
      .filter((label): label is LabelChipSummary => Boolean(label));
    items.push({
      field: "labelIds",
      icon: (
        <TagsIcon
          className="h-4 w-auto"
          style={{ color: selectedLabels[0]?.color }}
        />
      ),
      label: "Label",
      ...getOperatorConfig(filters, "labelIds"),
      value: <LabelChipValue labels={selectedLabels} />,
    });
  }

  if (filters.estimateValues?.length) {
    items.push({
      field: "estimateValues",
      icon: <EstimateIcon className="h-4 w-auto" />,
      label: "Complexity",
      ...getOperatorConfig(filters, "estimateValues"),
      value: (
        <EstimateChipValue
          estimateScheme={estimateScheme}
          estimateValues={filters.estimateValues}
        />
      ),
    });
  }

  if (filters.objectiveId) {
    items.push({
      field: "objectiveId",
      icon: <ObjectiveIcon className="h-4 w-auto" />,
      label: getTermDisplay("objectiveTerm", { capitalize: true }),
      ...getOperatorConfig(filters, "objectiveId"),
      value: objectives.get(filters.objectiveId) ?? filters.objectiveId,
    });
  }

  if (filters.keyResultId) {
    items.push({
      field: "keyResultId",
      icon: <OKRIcon className="h-4 w-auto" />,
      label: getTermDisplay("keyResultTerm", { capitalize: true }),
      operator: "is",
      value: keyResults.get(filters.keyResultId) ?? filters.keyResultId,
    });
  }

  if (filters.hasNoAssignee) {
    items.push({
      field: "hasNoAssignee",
      label: "Assignee",
      ...getOperatorConfig(filters, "hasNoAssignee"),
      value: "empty",
    });
  }

  return items.filter((item) => !hiddenFields.has(item.field));
};
