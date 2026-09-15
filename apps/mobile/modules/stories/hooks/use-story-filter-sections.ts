import type {
  StoryFilterOption,
  StoryFilterSection,
  StoryFiltersSheetProps,
} from "../components/story-filters.types";
import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  memberKeys,
  objectiveKeys,
  sprintKeys,
  statusKeys,
} from "@/constants/keys";
import { useTerminology } from "@/hooks/use-terminology";
import { useFeatures } from "@/hooks";
import { useTeams } from "@/modules/teams/hooks/use-teams";
import { getMembers } from "@/modules/members/queries/get-members";
import { getMayaAssignee } from "@/modules/members/queries/get-maya-assignee";
import {
  getStatuses,
  getTeamStatuses,
} from "@/modules/statuses/queries/get-statuses";
import {
  getSprints,
  getTeamSprints,
} from "@/modules/sprints/queries/get-sprints";
import {
  getObjectives,
  getTeamObjectives,
} from "@/modules/objectives/queries/get-objectives";
import { STORY_FILTER_PRIORITIES } from "../components/story-filter-selection";

function selectionLabel(
  options: StoryFilterOption[],
  ids: string[],
  empty: string,
) {
  if (!ids.length) return empty;
  if (ids.length > 1) return `${ids.length} selected`;
  const option = options.find((item) => item.id === ids[0]);
  if (!option) return "1 selected";
  return option.description
    ? `${option.label} · ${option.description}`
    : option.label;
}

export function useStoryFilterSections({
  isOpen,
  teamId,
  filters,
  allowAssignee = true,
}: StoryFiltersSheetProps): StoryFilterSection[] {
  const { getTermDisplay } = useTerminology();
  const { objectiveEnabled } = useFeatures();
  const { data: teams } = useTeams();
  const teamsById = useMemo(
    () => new Map((teams ?? []).map((team) => [team.id, team])),
    [teams],
  );
  const teamDescription = (optionTeamId: string) => {
    if (teamId) return undefined;
    const team = teamsById.get(optionTeamId);
    return team ? `${team.code} · ${team.name}` : undefined;
  };
  const statuses = useQuery({
    queryKey: teamId ? statusKeys.team(teamId) : statusKeys.lists(),
    queryFn: ({ signal }) =>
      teamId ? getTeamStatuses(teamId, signal) : getStatuses(signal),
    enabled: isOpen,
  });
  const members = useQuery({
    queryKey: memberKeys.lists(),
    queryFn: ({ signal }) => getMembers(signal),
    enabled: isOpen && allowAssignee,
  });
  const maya = useQuery({
    queryKey: memberKeys.maya(),
    queryFn: ({ signal }) => getMayaAssignee(signal),
    enabled: isOpen && allowAssignee,
    staleTime: 10 * 60 * 1000,
  });
  const sprints = useQuery({
    queryKey: teamId ? sprintKeys.team(teamId) : sprintKeys.lists(),
    queryFn: ({ signal }) =>
      teamId ? getTeamSprints(teamId, signal) : getSprints(signal),
    enabled: isOpen,
  });
  const objectives = useQuery({
    queryKey: teamId ? objectiveKeys.team(teamId) : objectiveKeys.lists(),
    queryFn: ({ signal }) =>
      teamId ? getTeamObjectives(teamId, signal) : getObjectives(signal),
    enabled: isOpen && Boolean(objectiveEnabled || filters.objectiveId),
  });
  const statusOptions = (statuses.data ?? []).map(
    ({ id, name, color, category, teamId: optionTeamId }) => ({
      id,
      label: name,
      description: teamDescription(optionTeamId),
      color,
      statusCategory: category,
    }),
  );
  const people = [...(members.data ?? [])];
  const memberOptions: StoryFilterOption[] = people.map(
    ({ id, fullName, username }) => ({ id, label: fullName || username }),
  );
  if (
    maya.data?.isSystem &&
    !memberOptions.some(({ id }) => id === maya.data?.id)
  )
    memberOptions.push({
      id: maya.data.id,
      label: maya.data.fullName || maya.data.username,
    });
  const sprintOptions = (sprints.data ?? []).map(
    ({ id, name, teamId: optionTeamId }) => ({
      id,
      label: name,
      description: teamDescription(optionTeamId),
    }),
  );
  const objectiveOptions = (objectives.data ?? []).map(
    ({ id, name, teamId: optionTeamId }) => ({
      id,
      label: name,
      description: teamDescription(optionTeamId),
    }),
  );
  const priorityOptions = STORY_FILTER_PRIORITIES.map((id) => ({
    id,
    label: id,
    priority: id,
  }));
  const assigneeIds =
    filters.assignee.kind === "members"
      ? filters.assignee.ids
      : filters.assignee.kind === "unassigned"
        ? ["unassigned"]
        : [];
  return [
    {
      id: "status",
      label: "Status",
      options: statusOptions,
      selectedIds: filters.statusIds,
      value: selectionLabel(statusOptions, filters.statusIds, "Any status"),
      loading: statuses.isPending,
      error: statuses.error,
      retry: () => {
        void statuses.refetch();
      },
    },
    {
      id: "priority",
      label: "Priority",
      options: priorityOptions,
      selectedIds: filters.priorities,
      value: selectionLabel(
        priorityOptions,
        filters.priorities,
        "Any priority",
      ),
      loading: false,
      error: null,
      retry: () => {},
    },
    ...(allowAssignee
      ? [
          {
            id: "assignee" as const,
            label: "Assignee",
            options: [
              { id: "unassigned", label: "Unassigned" },
              ...memberOptions,
            ],
            selectedIds: assigneeIds,
            value:
              filters.assignee.kind === "unassigned"
                ? "Unassigned"
                : selectionLabel(memberOptions, assigneeIds, "Anyone"),
            loading: members.isPending,
            error: members.error ?? maya.error,
            retry: () => {
              void members.refetch();
              void maya.refetch();
            },
          },
        ]
      : []),
    {
      id: "sprint",
      label: getTermDisplay("sprintTerm", { capitalize: true }),
      options: sprintOptions,
      selectedIds: filters.sprintIds,
      value: selectionLabel(
        sprintOptions,
        filters.sprintIds,
        `Any ${getTermDisplay("sprintTerm")}`,
      ),
      loading: sprints.isPending,
      error: sprints.error,
      retry: () => {
        void sprints.refetch();
      },
    },
    ...(objectiveEnabled || filters.objectiveId
      ? [
          {
            id: "objective" as const,
            label: getTermDisplay("objectiveTerm", { capitalize: true }),
            options: objectiveOptions,
            selectedIds: filters.objectiveId ? [filters.objectiveId] : [],
            value: selectionLabel(
              objectiveOptions,
              filters.objectiveId ? [filters.objectiveId] : [],
              `Any ${getTermDisplay("objectiveTerm")}`,
            ),
            loading: objectives.isPending,
            error: objectives.error,
            retry: () => {
              void objectives.refetch();
            },
          },
        ]
      : []),
  ];
}
