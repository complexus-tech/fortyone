import {
  AssigneeIcon,
  CalendarIcon,
  EstimateIcon,
  ListIcon,
  ObjectiveIcon,
  OKRIcon,
  SprintsIcon,
  TagsIcon,
  TeamIcon,
  UserIcon,
} from "icons";
import type { StoriesFilter } from "../stories-filter-types";
import { PriorityIcon } from "../priority-icon";
import { StoryStatusIcon } from "../story-status-icon";
import type { FilterOption, StoriesFilterField } from "./types";

type GetFilterTermDisplay = (
  term: "keyResultTerm" | "objectiveTerm",
  options: { capitalize: true },
) => string;

export const buildFilterOptions = ({
  filters,
  getTermDisplay,
  hasRouteTeam,
  hiddenFields,
}: {
  filters: StoriesFilter;
  getTermDisplay: GetFilterTermDisplay;
  hasRouteTeam: boolean;
  hiddenFields: ReadonlySet<StoriesFilterField>;
}) => {
  const options: FilterOption[] = [
    {
      field: "statusIds",
      icon: <StoryStatusIcon statusId={filters.statusIds?.[0] ?? ""} />,
      label: "Status",
    },
    {
      field: "assigneeIds",
      icon: <AssigneeIcon className="h-5 w-auto" />,
      label: "Assignee",
    },
    {
      field: "reporterIds",
      icon: <UserIcon className="h-5 w-auto" />,
      label: "Creator",
    },
    {
      field: "contentContains",
      icon: <ListIcon className="h-5 w-auto" />,
      label: "Content",
    },
    {
      field: "priorities",
      icon: <PriorityIcon priority="No Priority" />,
      label: "Priority",
    },
    ...(!hasRouteTeam
      ? ([
          {
            field: "teamIds",
            icon: <TeamIcon className="h-5 w-auto" />,
            label: "Team",
          },
        ] satisfies FilterOption[])
      : []),
    {
      field: "sprintIds",
      icon: <SprintsIcon className="h-5 w-auto" />,
      label: "Sprint",
    },
    {
      field: "labelIds",
      icon: <TagsIcon className="h-5 w-auto" />,
      label: "Label",
    },
    {
      field: "estimateValues",
      icon: <EstimateIcon className="h-5 w-auto" />,
      label: "Complexity",
    },
    {
      field: "objectiveId",
      icon: <ObjectiveIcon className="h-5 w-auto" />,
      label: getTermDisplay("objectiveTerm", { capitalize: true }),
    },
    {
      field: "keyResultId",
      icon: <OKRIcon className="h-5 w-auto" />,
      label: getTermDisplay("keyResultTerm", { capitalize: true }),
    },
    {
      field: "startDate",
      icon: <CalendarIcon className="h-5 w-auto" />,
      label: "Start date",
    },
    {
      field: "endDate",
      icon: <CalendarIcon className="h-5 w-auto" />,
      label: "End date",
    },
  ];

  return options.filter((option) => !hiddenFields.has(option.field));
};
