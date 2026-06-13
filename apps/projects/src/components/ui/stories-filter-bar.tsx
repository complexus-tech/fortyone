"use client";
import {
  useDeferredValue,
  useMemo,
  useState,
  type ReactNode,
  type UIEvent,
} from "react";
import {
  Avatar,
  Box,
  Button,
  Calendar,
  Command,
  Divider,
  Dialog,
  Flex,
  Input,
  Menu,
  Popover,
  Text,
} from "ui";
import {
  AssigneeIcon,
  ArrowRightIcon,
  CalendarIcon,
  CheckIcon,
  CloseIcon,
  EstimateIcon,
  ListIcon,
  ObjectiveIcon,
  PlusIcon,
  SprintsIcon,
  TagsIcon,
  TeamIcon,
  UserIcon,
} from "icons";
import { format, formatISO } from "date-fns";
import { useParams } from "next/navigation";
import { useStatuses } from "@/lib/hooks/statuses";
import {
  MEMBER_MENU_PAGE_SIZE,
  useMembers,
  useMembersInfinite,
} from "@/lib/hooks/members";
import {
  useTeamMembers,
  useTeamMembersInfinite,
} from "@/lib/hooks/team-members";
import {
  TEAM_MENU_PAGE_SIZE,
  useTeams,
  useTeamsInfinite,
} from "@/modules/teams/hooks/teams";
import {
  SPRINT_MENU_PAGE_SIZE,
  useTeamSprints,
  useTeamSprintsInfinite,
} from "@/modules/sprints/hooks/team-sprints";
import {
  OBJECTIVE_MENU_PAGE_SIZE,
  useTeamObjectives,
  useTeamObjectivesInfinite,
} from "@/modules/objectives/hooks/use-objectives";
import type { StoryPriority } from "@/modules/stories/types";
import { useTeamSettings } from "@/modules/teams/hooks/use-team-settings";
import {
  LABEL_MENU_PAGE_SIZE,
  useLabels,
  useLabelsInfinite,
} from "@/lib/hooks/labels";
import {
  formatEstimate,
  getEstimateOptions,
  type EstimateScheme,
} from "@/lib/estimate";
import { getScopedStoriesFilterTeamId } from "./stories-filter-query";
import type { StoriesFilter } from "./stories-filter-types";
import { MenuLoadingSkeleton } from "./menu-loading-skeleton";
import { PriorityIcon } from "./priority-icon";
import { StoryStatusIcon } from "./story-status-icon";
import { TeamColor } from "./team-color";
import { hasActiveStoriesFilters } from "./stories-filter-utils";

export type StoriesFilterField =
  | "contentContains"
  | "statusIds"
  | "assigneeIds"
  | "reporterIds"
  | "priorities"
  | "teamIds"
  | "sprintIds"
  | "labelIds"
  | "estimateValues"
  | "objectiveId"
  | "startDate"
  | "endDate"
  | "assignedToMe"
  | "createdByMe"
  | "hasNoAssignee";

type FilterChip = {
  field: StoriesFilterField;
  label: string;
  operator: string;
  value: ReactNode;
  icon?: ReactNode;
};

type FilterOption = {
  field: StoriesFilterField;
  icon: ReactNode;
  label: string;
};

type StoriesFilterBarProps = {
  filters: StoriesFilter;
  setFilters: (value: StoriesFilter) => void;
  resetFilters: () => void;
  hiddenFields?: StoriesFilterField[];
  showWhenEmpty?: boolean;
};

const EMPTY_FILTER_FIELDS: StoriesFilterField[] = [];

const getNames = (
  ids: string[] | null | undefined,
  labelsById: Map<string, string>,
) => {
  if (!ids?.length) {
    return "";
  }

  return ids.map((id) => labelsById.get(id) ?? id).join(", ");
};

const normalizeArrayFilter = (values: string[]) =>
  values.length > 0 ? values : null;

const normalizeNumberArrayFilter = (values: number[]) =>
  values.length > 0 ? values : null;

const getPluralLabel = (count: number, singular: string, plural: string) =>
  `${count} ${count === 1 ? singular : plural}`;

type UserChipSummary = {
  avatarUrl: string | null;
  id: string;
  name: string;
  username: string;
};

const PeopleChipValue = ({
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

const StatusChipValue = ({ statuses }: { statuses: StatusChipSummary[] }) => {
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

const PriorityChipValue = ({ priorities }: { priorities: StoryPriority[] }) => {
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

type LabelChipSummary = {
  color: string;
  id: string;
  name: string;
};

const LabelChipValue = ({ labels }: { labels: LabelChipSummary[] }) => {
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

const EstimateChipValue = ({
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
          {getPluralLabel(estimateValues.length, "estimate", "estimates")}
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

const getEditorContentClassName = (field: StoriesFilterField) => {
  if (field === "contentContains") {
    return "w-80 overflow-hidden py-2";
  }

  if (field === "objectiveId") {
    return "w-96 overflow-hidden py-2";
  }

  if (field === "assigneeIds" || field === "reporterIds") {
    return "w-80 overflow-hidden py-2";
  }

  if (field === "startDate" || field === "endDate") {
    return "w-auto overflow-hidden py-2";
  }

  return "w-64 overflow-hidden py-2";
};

const TitleFilterDialog = ({
  open,
  onOpenChange,
  filters,
  setFilters,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  filters: StoriesFilter;
  setFilters: (value: StoriesFilter) => void;
}) => {
  const [draft, setDraft] = useState(filters.contentContains ?? "");

  const applyTitleFilter = () => {
    const contentContains = draft.trim();
    setFilters({
      ...filters,
      contentContains: contentContains ? contentContains : null,
    });
    onOpenChange(false);
  };

  return (
    <Dialog onOpenChange={onOpenChange} open={open}>
      <Dialog.Content className="max-w-lg" hideClose>
        <Dialog.Header className="px-6 pt-6">
          <Dialog.Title className="text-lg">Filter by content</Dialog.Title>
        </Dialog.Header>
        <Dialog.Body className="pt-1">
          <Input
            autoFocus
            onChange={(event) => {
              setDraft(event.target.value);
            }}
            onKeyDown={(event) => {
              if (event.key === "Enter") {
                applyTitleFilter();
              }
            }}
            placeholder="Content contains..."
            value={draft}
          />
        </Dialog.Body>
        <Dialog.Footer className="justify-end gap-3 border-0 pt-2">
          <Button
            color="tertiary"
            onClick={() => {
              onOpenChange(false);
            }}
            variant="outline"
          >
            Cancel
          </Button>
          <Button onClick={applyTitleFilter}>Apply</Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};

const StatusEditor = ({
  filters,
  setFilters,
}: {
  filters: StoriesFilter;
  setFilters: (value: StoriesFilter) => void;
}) => {
  const { teamId } = useParams<{ teamId?: string }>();
  const [query, setQuery] = useState("");
  const { data: allStatuses = [] } = useStatuses();
  const statuses = teamId
    ? allStatuses.filter((status) => status.teamId === teamId)
    : allStatuses;
  const filteredStatuses = statuses.filter((status) =>
    status.name.toLowerCase().includes(query.toLowerCase()),
  );

  const toggleStatus = (statusId: string) => {
    const selected = filters.statusIds ?? [];
    const statusIds = selected.includes(statusId)
      ? selected.filter((id) => id !== statusId)
      : [...selected, statusId];
    setFilters({ ...filters, statusIds: normalizeArrayFilter(statusIds) });
  };

  return (
    <Command>
      <Command.Input
        autoFocus
        onValueChange={setQuery}
        placeholder="Search status..."
        value={query}
      />
      <Divider className="my-2" />
      <Command.Empty className="py-2">
        <Text color="muted">No statuses found.</Text>
      </Command.Empty>
      <Command.Group className="max-h-80 overflow-y-auto">
        {filteredStatuses.map((status, idx) => (
          <Command.Item
            active={Boolean(filters.statusIds?.includes(status.id))}
            className="justify-between gap-4"
            key={status.id}
            onSelect={() => {
              toggleStatus(status.id);
            }}
            value={status.name}
          >
            <Box className="grid min-w-0 flex-1 grid-cols-[16px_minmax(0,1fr)] items-center">
              <span className="min-w-0">
                <StoryStatusIcon statusId={status.id} />
              </span>
              <Text className="max-w-[22ch] truncate">{status.name}</Text>
            </Box>
            <Flex align="center" className="shrink-0" gap={2}>
              {filters.statusIds?.includes(status.id) ? (
                <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
              ) : null}
              <Text color="muted">{idx}</Text>
            </Flex>
          </Command.Item>
        ))}
      </Command.Group>
    </Command>
  );
};

const PeopleEditor = ({
  field,
  filters,
  setFilters,
}: {
  field: "assigneeIds" | "reporterIds";
  filters: StoriesFilter;
  setFilters: (value: StoriesFilter) => void;
}) => {
  const { teamId } = useParams<{ teamId?: string }>();
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);
  const workspaceMembersQuery = useMembersInfinite(
    deferredQuery,
    MEMBER_MENU_PAGE_SIZE,
    !teamId,
  );
  const teamMembersQuery = useTeamMembersInfinite(
    teamId,
    deferredQuery,
    MEMBER_MENU_PAGE_SIZE,
    Boolean(teamId),
  );
  const membersQuery = teamId ? teamMembersQuery : workspaceMembersQuery;
  const members =
    membersQuery.data?.pages.flatMap((page) => page.members) ?? [];
  const isLoadingMembers =
    membersQuery.isFetching && !membersQuery.isFetchingNextPage;

  const toggleMember = (memberId: string) => {
    const selected = filters[field] ?? [];
    const memberIds = selected.includes(memberId)
      ? selected.filter((id) => id !== memberId)
      : [...selected, memberId];
    setFilters({ ...filters, [field]: normalizeArrayFilter(memberIds) });
  };

  const handleScroll = (event: UIEvent<HTMLDivElement>) => {
    const target = event.currentTarget;
    const distanceToBottom =
      target.scrollHeight - target.scrollTop - target.clientHeight;

    if (
      distanceToBottom <= 80 &&
      membersQuery.hasNextPage &&
      !membersQuery.isFetchingNextPage
    ) {
      void membersQuery.fetchNextPage();
    }
  };

  return (
    <Command>
      <Command.Input
        autoFocus
        onValueChange={setQuery}
        placeholder="Search people..."
        value={query}
      />
      <Divider className="my-2" />
      {!isLoadingMembers ? (
        <Command.Empty className="py-2">
          <Text color="muted">No user found.</Text>
        </Command.Empty>
      ) : null}
      <Command.Group
        className="max-h-80 overflow-y-auto md:max-h-100"
        onScroll={handleScroll}
      >
        {isLoadingMembers ? (
          <Command.Loading className="p-2">
            <MenuLoadingSkeleton avatar rows={5} />
          </Command.Loading>
        ) : null}
        {field === "assigneeIds" ? (
          <Command.Item
            active={Boolean(filters.hasNoAssignee)}
            className="justify-between gap-4"
            onSelect={() => {
              setFilters({
                ...filters,
                assigneeIds: null,
                hasNoAssignee: filters.hasNoAssignee ? null : true,
              });
            }}
            value="No assignee"
          >
            <Flex align="center" className="min-w-0 flex-1" gap={2}>
              <Avatar
                className="text-foreground/80"
                color="primary"
                size="sm"
              />
              <Text className="max-w-48 truncate">No assignee</Text>
            </Flex>
            <Flex align="center" className="shrink-0" gap={1}>
              {filters.hasNoAssignee ? (
                <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
              ) : null}
              <Text color="muted">0</Text>
            </Flex>
          </Command.Item>
        ) : null}
        {members.map((member, idx) => {
          const name = member.fullName || member.username || member.email;
          return (
            <Command.Item
              active={Boolean(filters[field]?.includes(member.id))}
              className="justify-between gap-4"
              key={member.id}
              onSelect={() => {
                toggleMember(member.id);
              }}
              value={name}
            >
              <Flex align="center" className="min-w-0 flex-1" gap={2}>
                <Avatar
                  color="primary"
                  name={name}
                  size="sm"
                  src={member.avatarUrl}
                />
                <Text className="max-w-48 truncate">{name}</Text>
              </Flex>
              <Flex align="center" className="shrink-0" gap={1}>
                {filters[field]?.includes(member.id) ? (
                  <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                ) : null}
                <Text color="muted">
                  {idx + (field === "assigneeIds" ? 1 : 0)}
                </Text>
              </Flex>
            </Command.Item>
          );
        })}
        {membersQuery.isFetchingNextPage ? (
          <Command.Loading className="p-2">
            <MenuLoadingSkeleton avatar rows={2} />
          </Command.Loading>
        ) : null}
      </Command.Group>
    </Command>
  );
};

const PriorityEditor = ({
  filters,
  setFilters,
}: {
  filters: StoriesFilter;
  setFilters: (value: StoriesFilter) => void;
}) => {
  const priorities = [
    "Urgent",
    "High",
    "Medium",
    "Low",
    "No Priority",
  ] as StoryPriority[];

  const togglePriority = (priority: StoryPriority) => {
    const selected = filters.priorities ?? [];
    const priorities = selected.includes(priority)
      ? selected.filter((value) => value !== priority)
      : [...selected, priority];
    setFilters({ ...filters, priorities: normalizeArrayFilter(priorities) });
  };

  return (
    <Command>
      <Command.Input autoFocus placeholder="Change priority..." />
      <Divider className="my-2" />
      <Command.Empty className="py-2">
        <Text color="muted">No priority found.</Text>
      </Command.Empty>
      <Command.Group>
        {priorities.map((priority, idx) => (
          <Command.Item
            active={Boolean(filters.priorities?.includes(priority))}
            className="justify-between gap-4"
            key={priority}
            onSelect={() => {
              togglePriority(priority);
            }}
            value={priority}
          >
            <Box className="grid min-w-0 flex-1 grid-cols-[24px_minmax(0,1fr)] items-center">
              <PriorityIcon priority={priority} />
              <Text className="truncate">{priority}</Text>
            </Box>
            <Flex align="center" className="shrink-0" gap={2}>
              {filters.priorities?.includes(priority) ? (
                <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
              ) : null}
              <Text color="muted">{idx}</Text>
            </Flex>
          </Command.Item>
        ))}
      </Command.Group>
    </Command>
  );
};

const TeamEditor = ({
  filters,
  setFilters,
}: {
  filters: StoriesFilter;
  setFilters: (value: StoriesFilter) => void;
}) => {
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);
  const { data, fetchNextPage, hasNextPage, isFetching, isFetchingNextPage } =
    useTeamsInfinite(deferredQuery, TEAM_MENU_PAGE_SIZE);
  const teams = data?.pages.flatMap((page) => page.teams) ?? [];
  const isLoadingTeams = isFetching && !isFetchingNextPage;

  const handleScroll = (event: UIEvent<HTMLDivElement>) => {
    const target = event.currentTarget;
    const distanceToBottom =
      target.scrollHeight - target.scrollTop - target.clientHeight;

    if (distanceToBottom <= 80 && hasNextPage && !isFetchingNextPage) {
      void fetchNextPage();
    }
  };

  const toggleTeam = (teamId: string) => {
    const selected = filters.teamIds ?? [];
    const teamIds = selected.includes(teamId)
      ? selected.filter((id) => id !== teamId)
      : [...selected, teamId];
    setFilters({ ...filters, teamIds: normalizeArrayFilter(teamIds) });
  };

  return (
    <Command>
      <Command.Input
        autoFocus
        onValueChange={setQuery}
        placeholder="Search teams..."
        value={query}
      />
      <Divider className="my-2" />
      {!isLoadingTeams ? (
        <Command.Empty className="py-2">
          <Text color="muted">No teams found.</Text>
        </Command.Empty>
      ) : null}
      <Command.Group
        className="max-h-80 overflow-y-auto"
        onScroll={handleScroll}
      >
        {isLoadingTeams ? (
          <Command.Loading className="p-2">
            <MenuLoadingSkeleton rows={5} />
          </Command.Loading>
        ) : null}
        {teams.map((team, idx) => (
          <Command.Item
            active={Boolean(filters.teamIds?.includes(team.id))}
            className="justify-between gap-4"
            key={team.id}
            onSelect={() => {
              toggleTeam(team.id);
            }}
            value={team.name}
          >
            <Flex align="center" className="min-w-0 flex-1" gap={2}>
              <TeamColor color={team.color} />
              <Text className="max-w-48 truncate">{team.name}</Text>
            </Flex>
            <Flex align="center" className="shrink-0" gap={2}>
              {filters.teamIds?.includes(team.id) ? (
                <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
              ) : null}
              <Text color="muted">{idx}</Text>
            </Flex>
          </Command.Item>
        ))}
        {isFetchingNextPage ? (
          <Command.Loading className="p-2">
            <MenuLoadingSkeleton rows={2} />
          </Command.Loading>
        ) : null}
      </Command.Group>
    </Command>
  );
};

const SprintEditor = ({
  filters,
  setFilters,
}: {
  filters: StoriesFilter;
  setFilters: (value: StoriesFilter) => void;
}) => {
  const { teamId } = useParams<{ teamId?: string }>();
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);
  const scopedTeamId = getScopedStoriesFilterTeamId(teamId, filters.teamIds);
  const { data, fetchNextPage, hasNextPage, isFetching, isFetchingNextPage } =
    useTeamSprintsInfinite(
      scopedTeamId ?? "",
      deferredQuery,
      SPRINT_MENU_PAGE_SIZE,
    );
  const sprints = data?.pages.flatMap((page) => page.sprints) ?? [];
  const isLoadingSprints =
    Boolean(scopedTeamId) && isFetching && !isFetchingNextPage;
  const needsSingleTeam = !scopedTeamId;

  const toggleSprint = (sprintId: string) => {
    const selected = filters.sprintIds ?? [];
    const sprintIds = selected.includes(sprintId)
      ? selected.filter((id) => id !== sprintId)
      : [...selected, sprintId];
    setFilters({ ...filters, sprintIds: normalizeArrayFilter(sprintIds) });
  };

  const handleScroll = (event: UIEvent<HTMLDivElement>) => {
    const target = event.currentTarget;
    const distanceToBottom =
      target.scrollHeight - target.scrollTop - target.clientHeight;

    if (distanceToBottom <= 80 && hasNextPage && !isFetchingNextPage) {
      void fetchNextPage();
    }
  };

  return (
    <Command>
      <Command.Input
        autoFocus
        onValueChange={setQuery}
        placeholder="Search sprints..."
        value={query}
      />
      <Divider className="my-2" />
      {!isLoadingSprints ? (
        <Command.Empty className="py-2">
          <Text color="muted">
            {needsSingleTeam ? "Select one team first." : "No sprints found."}
          </Text>
        </Command.Empty>
      ) : null}
      <Command.Group
        className="max-h-80 overflow-y-auto"
        onScroll={handleScroll}
      >
        {isLoadingSprints ? (
          <Command.Loading className="p-2">
            <MenuLoadingSkeleton rows={5} />
          </Command.Loading>
        ) : null}
        {!needsSingleTeam
          ? sprints.map((sprint, idx) => (
              <Command.Item
                active={Boolean(filters.sprintIds?.includes(sprint.id))}
                className="justify-between gap-4"
                key={sprint.id}
                onSelect={() => {
                  toggleSprint(sprint.id);
                }}
                value={sprint.name}
              >
                <Flex align="center" className="min-w-0 flex-1" gap={2}>
                  <SprintsIcon className="text-text-secondary h-4 w-auto" />
                  <Text className="max-w-48 truncate">{sprint.name}</Text>
                </Flex>
                <Flex align="center" className="shrink-0" gap={2}>
                  {filters.sprintIds?.includes(sprint.id) ? (
                    <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                  ) : null}
                  <Text color="muted">{idx}</Text>
                </Flex>
              </Command.Item>
            ))
          : null}
        {isFetchingNextPage ? (
          <Command.Loading className="p-2">
            <MenuLoadingSkeleton rows={2} />
          </Command.Loading>
        ) : null}
      </Command.Group>
    </Command>
  );
};

const ObjectiveEditor = ({
  filters,
  setFilters,
}: {
  filters: StoriesFilter;
  setFilters: (value: StoriesFilter) => void;
}) => {
  const { teamId } = useParams<{ teamId?: string }>();
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);
  const scopedTeamId = getScopedStoriesFilterTeamId(teamId, filters.teamIds);
  const { data, fetchNextPage, hasNextPage, isFetching, isFetchingNextPage } =
    useTeamObjectivesInfinite(
      scopedTeamId ?? "",
      deferredQuery,
      OBJECTIVE_MENU_PAGE_SIZE,
    );
  const objectives = data?.pages.flatMap((page) => page.objectives) ?? [];
  const isLoadingObjectives =
    Boolean(scopedTeamId) && isFetching && !isFetchingNextPage;
  const needsSingleTeam = !scopedTeamId;

  const handleScroll = (event: UIEvent<HTMLDivElement>) => {
    const target = event.currentTarget;
    const distanceToBottom =
      target.scrollHeight - target.scrollTop - target.clientHeight;

    if (distanceToBottom <= 80 && hasNextPage && !isFetchingNextPage) {
      void fetchNextPage();
    }
  };

  return (
    <Command>
      <Command.Input
        autoFocus
        onValueChange={setQuery}
        placeholder="Search objectives..."
        value={query}
      />
      <Divider className="my-2" />
      {!isLoadingObjectives ? (
        <Command.Empty className="py-2">
          <Text color="muted">
            {needsSingleTeam
              ? "Select one team first."
              : "No objectives found."}
          </Text>
        </Command.Empty>
      ) : null}
      <Command.Group
        className="max-h-80 overflow-y-auto"
        onScroll={handleScroll}
      >
        {isLoadingObjectives ? (
          <Command.Loading className="p-2">
            <MenuLoadingSkeleton rows={5} />
          </Command.Loading>
        ) : null}
        {!needsSingleTeam
          ? objectives.map((objective, idx) => (
              <Command.Item
                active={filters.objectiveId === objective.id}
                className="justify-between gap-4"
                key={objective.id}
                onSelect={() => {
                  setFilters({
                    ...filters,
                    objectiveId:
                      filters.objectiveId === objective.id
                        ? null
                        : objective.id,
                  });
                }}
                value={objective.name}
              >
                <Flex align="center" className="min-w-0 flex-1" gap={2}>
                  <ObjectiveIcon className="text-text-secondary h-4 w-auto" />
                  <Text className="max-w-64 truncate">{objective.name}</Text>
                </Flex>
                <Flex align="center" className="shrink-0" gap={2}>
                  {filters.objectiveId === objective.id ? (
                    <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                  ) : null}
                  <Text color="muted">{idx}</Text>
                </Flex>
              </Command.Item>
            ))
          : null}
        {isFetchingNextPage ? (
          <Command.Loading className="p-2">
            <MenuLoadingSkeleton rows={2} />
          </Command.Loading>
        ) : null}
      </Command.Group>
    </Command>
  );
};

const LabelEditor = ({
  filters,
  setFilters,
}: {
  filters: StoriesFilter;
  setFilters: (value: StoriesFilter) => void;
}) => {
  const { teamId } = useParams<{ teamId?: string }>();
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);
  const { data, fetchNextPage, hasNextPage, isFetching, isFetchingNextPage } =
    useLabelsInfinite({ search: deferredQuery, teamId }, LABEL_MENU_PAGE_SIZE);
  const labels = data?.pages.flatMap((page) => page.labels) ?? [];
  const isLoadingLabels = isFetching && !isFetchingNextPage;

  const toggleLabel = (labelId: string) => {
    const selected = filters.labelIds ?? [];
    const labelIds = selected.includes(labelId)
      ? selected.filter((id) => id !== labelId)
      : [...selected, labelId];
    setFilters({ ...filters, labelIds: normalizeArrayFilter(labelIds) });
  };

  const handleScroll = (event: UIEvent<HTMLDivElement>) => {
    const target = event.currentTarget;
    const distanceToBottom =
      target.scrollHeight - target.scrollTop - target.clientHeight;

    if (distanceToBottom <= 80 && hasNextPage && !isFetchingNextPage) {
      void fetchNextPage();
    }
  };

  return (
    <Command>
      <Command.Input
        autoFocus
        onValueChange={setQuery}
        placeholder="Search labels..."
        value={query}
      />
      <Divider className="my-2" />
      {!isLoadingLabels ? (
        <Command.Empty className="py-2">
          <Text color="muted">No labels found.</Text>
        </Command.Empty>
      ) : null}
      <Command.Group
        className="max-h-80 overflow-y-auto"
        onScroll={handleScroll}
      >
        {isLoadingLabels ? (
          <Command.Loading className="p-2">
            <MenuLoadingSkeleton rows={5} />
          </Command.Loading>
        ) : null}
        {labels.map((label, idx) => (
          <Command.Item
            active={Boolean(filters.labelIds?.includes(label.id))}
            className="justify-between gap-4"
            key={label.id}
            onSelect={() => {
              toggleLabel(label.id);
            }}
            value={label.name}
          >
            <Flex align="center" className="min-w-0 flex-1" gap={2}>
              <TagsIcon className="h-4 w-auto" style={{ color: label.color }} />
              <Text className="max-w-48 truncate">{label.name}</Text>
            </Flex>
            <Flex align="center" className="shrink-0" gap={2}>
              {filters.labelIds?.includes(label.id) ? (
                <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
              ) : null}
              <Text color="muted">{idx}</Text>
            </Flex>
          </Command.Item>
        ))}
        {isFetchingNextPage ? (
          <Command.Loading className="p-2">
            <MenuLoadingSkeleton rows={2} />
          </Command.Loading>
        ) : null}
      </Command.Group>
    </Command>
  );
};

const EstimateEditor = ({
  estimateScheme,
  filters,
  setFilters,
}: {
  estimateScheme: EstimateScheme;
  filters: StoriesFilter;
  setFilters: (value: StoriesFilter) => void;
}) => {
  const options = getEstimateOptions(estimateScheme);

  const toggleEstimate = (estimateValue: number) => {
    const selected = filters.estimateValues ?? [];
    const estimateValues = selected.includes(estimateValue)
      ? selected.filter((value) => value !== estimateValue)
      : [...selected, estimateValue];
    setFilters({
      ...filters,
      estimateValues: normalizeNumberArrayFilter(estimateValues),
    });
  };

  return (
    <Command>
      <Command.Input autoFocus placeholder="Change estimate..." />
      <Divider className="my-2" />
      <Command.Empty className="py-2">
        <Text color="muted">No estimate found.</Text>
      </Command.Empty>
      <Command.Group>
        {options.map(({ label, value }, idx) => (
          <Command.Item
            active={Boolean(filters.estimateValues?.includes(value))}
            className="justify-between gap-4"
            key={value}
            onSelect={() => {
              toggleEstimate(value);
            }}
            value={label}
          >
            <Box className="grid min-w-0 flex-1 grid-cols-[24px_minmax(0,1fr)] items-center">
              <EstimateIcon className="text-text-secondary h-4 w-auto" />
              <Text className="truncate">
                {formatEstimate(estimateScheme, value, "full")}
              </Text>
            </Box>
            <Flex align="center" className="shrink-0" gap={2}>
              {filters.estimateValues?.includes(value) ? (
                <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
              ) : null}
              <Text color="muted">{idx}</Text>
            </Flex>
          </Command.Item>
        ))}
      </Command.Group>
    </Command>
  );
};

const DateEditor = ({
  field,
  filters,
  setFilters,
}: {
  field: "startDate" | "endDate";
  filters: StoriesFilter;
  setFilters: (value: StoriesFilter) => void;
}) => {
  const selectedDate = filters[field] ? new Date(filters[field]) : undefined;

  return (
    <Box>
      <Calendar
        className="px-3 py-3 shadow-none"
        mode="single"
        onDayClick={(date) => {
          setFilters({
            ...filters,
            [field]: formatISO(date, { representation: "date" }),
          });
        }}
        selected={selectedDate}
      />
      {selectedDate ? (
        <Button
          className="mx-3 mb-2 w-[calc(100%-1.5rem)] justify-start"
          color="tertiary"
          onClick={() => {
            setFilters({ ...filters, [field]: null });
          }}
          size="sm"
          variant="naked"
        >
          Clear date
        </Button>
      ) : null}
    </Box>
  );
};

const FilterValueEditor = ({
  field,
  estimateScheme,
  filters,
  setFilters,
}: {
  field: StoriesFilterField;
  estimateScheme: EstimateScheme;
  filters: StoriesFilter;
  setFilters: (value: StoriesFilter) => void;
}) => {
  if (field === "statusIds") {
    return <StatusEditor filters={filters} setFilters={setFilters} />;
  }

  if (field === "assigneeIds" || field === "reporterIds") {
    return (
      <PeopleEditor field={field} filters={filters} setFilters={setFilters} />
    );
  }

  if (field === "priorities") {
    return <PriorityEditor filters={filters} setFilters={setFilters} />;
  }

  if (field === "teamIds") {
    return <TeamEditor filters={filters} setFilters={setFilters} />;
  }

  if (field === "sprintIds") {
    return <SprintEditor filters={filters} setFilters={setFilters} />;
  }

  if (field === "objectiveId") {
    return <ObjectiveEditor filters={filters} setFilters={setFilters} />;
  }

  if (field === "labelIds") {
    return <LabelEditor filters={filters} setFilters={setFilters} />;
  }

  if (field === "estimateValues") {
    return (
      <EstimateEditor
        estimateScheme={estimateScheme}
        filters={filters}
        setFilters={setFilters}
      />
    );
  }

  if (field === "startDate" || field === "endDate") {
    return (
      <DateEditor field={field} filters={filters} setFilters={setFilters} />
    );
  }

  return null;
};

const Chip = ({
  chip,
  estimateScheme,
  filters,
  setFilters,
  onEditTitle,
  onRemove,
}: {
  chip: FilterChip;
  estimateScheme: EstimateScheme;
  filters: StoriesFilter;
  setFilters: (value: StoriesFilter) => void;
  onEditTitle: () => void;
  onRemove: () => void;
}) => {
  const isEditable =
    chip.field !== "assignedToMe" &&
    chip.field !== "createdByMe" &&
    chip.field !== "hasNoAssignee";
  const shouldUseDialog = chip.field === "contentContains";
  const valueContent = (
    <div className="flex min-w-0 items-center truncate">{chip.value}</div>
  );
  let valueControl: ReactNode = (
    <div className="flex h-full items-center px-2.5">{chip.value}</div>
  );

  if (shouldUseDialog) {
    valueControl = (
      <button
        className="hover:bg-state-hover flex h-full max-w-72 items-center truncate px-2.5 text-left transition"
        onClick={onEditTitle}
        type="button"
      >
        {valueContent}
      </button>
    );
  } else if (isEditable) {
    valueControl = (
      <Popover>
        <Popover.Trigger asChild>
          <button
            className="hover:bg-state-hover flex h-full max-w-72 items-center truncate px-2.5 text-left transition"
            type="button"
          >
            {valueContent}
          </button>
        </Popover.Trigger>
        <Popover.Content
          align="start"
          className={getEditorContentClassName(chip.field)}
        >
          <FilterValueEditor
            estimateScheme={estimateScheme}
            field={chip.field}
            filters={filters}
            setFilters={setFilters}
          />
        </Popover.Content>
      </Popover>
    );
  }

  return (
    <Flex
      align="center"
      className="border-border bg-surface h-[2.1rem] shrink-0 overflow-hidden rounded-xl border"
      gap={0}
    >
      <span className="border-border text-text-secondary flex h-full items-center gap-1.5 border-r px-2.5">
        {chip.icon}
        {chip.label}
      </span>
      <span className="border-border text-text-secondary flex h-full items-center border-r px-2.5">
        {chip.operator}
      </span>
      {valueControl}
      <button
        aria-label={`Remove ${chip.label} filter`}
        className="hover:bg-state-hover border-border flex h-full w-9 items-center justify-center border-l transition"
        onClick={onRemove}
        type="button"
      >
        <CloseIcon className="text-text-secondary h-3.5 w-auto" />
      </button>
    </Flex>
  );
};

export const StoriesFilterBar = ({
  filters,
  setFilters,
  resetFilters,
  hiddenFields = EMPTY_FILTER_FIELDS,
  showWhenEmpty = false,
}: StoriesFilterBarProps) => {
  const { teamId } = useParams<{ teamId?: string }>();
  const [titleDialogOpen, setTitleDialogOpen] = useState(false);
  const scopedTeamId = getScopedStoriesFilterTeamId(teamId, filters.teamIds);
  const { data: allStatuses = [] } = useStatuses();
  const { data: allUsers = [] } = useMembers();
  const resolvedTeamId = scopedTeamId ?? "";
  const { data: teamMembers = [] } = useTeamMembers(resolvedTeamId);
  const { data: teams = [] } = useTeams();
  const { data: sprints = [] } = useTeamSprints(resolvedTeamId);
  const { data: objectives = [] } = useTeamObjectives(resolvedTeamId);
  const { data: allLabels = [] } = useLabels();
  const { data: teamSettings } = useTeamSettings(scopedTeamId);
  const estimateScheme = teamSettings?.estimationSettings.scheme ?? "points";

  const users = scopedTeamId ? teamMembers : allUsers;
  const statuses = scopedTeamId
    ? allStatuses.filter((status) => status.teamId === scopedTeamId)
    : allStatuses;

  const statusById = useMemo(
    () => new Map(statuses.map((status) => [status.id, status.name])),
    [statuses],
  );
  const userById = useMemo(
    () =>
      new Map(
        users.map((user) => {
          const username = user.username || user.email || "Unknown user";
          return [
            user.id,
            {
              avatarUrl: user.avatarUrl ?? null,
              id: user.id,
              name: user.fullName || username,
              username,
            },
          ];
        }),
      ),
    [users],
  );
  const teamById = useMemo(
    () => new Map(teams.map((team) => [team.id, team.name])),
    [teams],
  );
  const teamColorById = useMemo(
    () => new Map(teams.map((team) => [team.id, team.color])),
    [teams],
  );
  const sprintById = useMemo(
    () => new Map(sprints.map((sprint) => [sprint.id, sprint.name])),
    [sprints],
  );
  const objectiveById = useMemo(
    () =>
      new Map(objectives.map((objective) => [objective.id, objective.name])),
    [objectives],
  );
  const labelById = useMemo(
    () =>
      new Map(
        allLabels.map((label) => [
          label.id,
          {
            color: label.color,
            id: label.id,
            name: label.name,
          },
        ]),
      ),
    [allLabels],
  );

  const chips = useMemo(() => {
    const items: FilterChip[] = [];

    if (filters.contentContains?.trim()) {
      items.push({
        field: "contentContains",
        label: "Content",
        operator: "contains",
        value: filters.contentContains.trim(),
        icon: <ListIcon className="h-4 w-auto" />,
      });
    }

    if (filters.startDate) {
      items.push({
        field: "startDate",
        label: "Start date",
        operator: "is",
        value: format(new Date(filters.startDate), "MMM d, yyyy"),
        icon: <CalendarIcon className="h-4 w-auto" />,
      });
    }

    if (filters.endDate) {
      items.push({
        field: "endDate",
        label: "End date",
        operator: "is",
        value: format(new Date(filters.endDate), "MMM d, yyyy"),
        icon: <CalendarIcon className="h-4 w-auto" />,
      });
    }

    if (filters.statusIds?.length) {
      const selectedStatuses = filters.statusIds.map((id) => ({
        id,
        name: statusById.get(id) ?? id,
      }));

      items.push({
        field: "statusIds",
        label: "Status",
        operator: "is any of",
        value: <StatusChipValue statuses={selectedStatuses} />,
        icon: <StoryStatusIcon statusId={filters.statusIds[0]} />,
      });
    }

    if (filters.assigneeIds?.length) {
      const selectedUsers = filters.assigneeIds
        .map((id) => userById.get(id))
        .filter((user): user is UserChipSummary => Boolean(user));

      items.push({
        field: "assigneeIds",
        label: "Assignee",
        operator: "is any of",
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
        .map((id) => userById.get(id))
        .filter((user): user is UserChipSummary => Boolean(user));

      items.push({
        field: "reporterIds",
        label: "Creator",
        operator: "is any of",
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
        label: "Priority",
        operator: "is any of",
        value: <PriorityChipValue priorities={selectedPriorities} />,
        icon: (
          <PriorityIcon priority={filters.priorities[0] as StoryPriority} />
        ),
      });
    }

    if (filters.teamIds?.length) {
      items.push({
        field: "teamIds",
        label: "Team",
        operator: "is any of",
        value: getNames(filters.teamIds, teamById),
        icon: <TeamColor color={teamColorById.get(filters.teamIds[0])} />,
      });
    }

    if (filters.sprintIds?.length) {
      items.push({
        field: "sprintIds",
        label: "Sprint",
        operator: "is any of",
        value: getNames(filters.sprintIds, sprintById),
      });
    }

    if (filters.labelIds?.length) {
      const selectedLabels = filters.labelIds
        .map((id) => labelById.get(id))
        .filter((label): label is LabelChipSummary => Boolean(label));

      items.push({
        field: "labelIds",
        label: "Label",
        operator: "is any of",
        value: <LabelChipValue labels={selectedLabels} />,
        icon: (
          <TagsIcon
            className="h-4 w-auto"
            style={{ color: selectedLabels[0]?.color }}
          />
        ),
      });
    }

    if (filters.estimateValues?.length) {
      items.push({
        field: "estimateValues",
        label: "Estimate",
        operator: "is any of",
        value: (
          <EstimateChipValue
            estimateScheme={estimateScheme}
            estimateValues={filters.estimateValues}
          />
        ),
        icon: <EstimateIcon className="h-4 w-auto" />,
      });
    }

    if (filters.objectiveId) {
      items.push({
        field: "objectiveId",
        label: "Objective",
        operator: "is",
        value: objectiveById.get(filters.objectiveId) ?? filters.objectiveId,
        icon: <ObjectiveIcon className="h-4 w-auto" />,
      });
    }

    if (filters.hasNoAssignee) {
      items.push({
        field: "hasNoAssignee",
        label: "Assignee",
        operator: "is",
        value: "empty",
      });
    }

    return items;
  }, [
    filters,
    estimateScheme,
    labelById,
    objectiveById,
    sprintById,
    statusById,
    teamById,
    teamColorById,
    userById,
  ]);

  const hiddenFieldSet = useMemo(() => new Set(hiddenFields), [hiddenFields]);

  const removeFilter = (field: StoriesFilterField) => {
    if (field === "assignedToMe" || field === "createdByMe") {
      setFilters({ ...filters, [field]: false });
      return;
    }

    setFilters({ ...filters, [field]: null });
  };

  const baseFilterOptions: FilterOption[] = [
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
      field: "priorities",
      icon: <PriorityIcon priority="No Priority" />,
      label: "Priority",
    },
    ...(!teamId
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
      label: "Estimate",
    },
    {
      field: "objectiveId",
      icon: <ObjectiveIcon className="h-5 w-auto" />,
      label: "Objective",
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
    {
      field: "contentContains",
      icon: <ListIcon className="h-5 w-auto" />,
      label: "Content",
    },
  ];
  const filterOptions = baseFilterOptions.filter(
    (option) => !hiddenFieldSet.has(option.field),
  );

  if (!showWhenEmpty && !hasActiveStoriesFilters(filters)) {
    return null;
  }

  return (
    <Flex
      align="center"
      className="border-border bg-background h-[3.6rem] border-b px-4"
      gap={3}
      justify="between"
    >
      <Flex align="center" className="min-w-0 flex-1 overflow-x-auto" gap={2}>
        {chips.map((chip) => (
          <Chip
            chip={chip}
            estimateScheme={estimateScheme}
            filters={filters}
            key={chip.field}
            onEditTitle={() => {
              setTitleDialogOpen(true);
            }}
            onRemove={() => {
              removeFilter(chip.field);
            }}
            setFilters={setFilters}
          />
        ))}
        <Menu>
          <Menu.Button>
            <Button
              aria-label="Add filter"
              color="tertiary"
              leftIcon={<PlusIcon className="h-4 w-auto" />}
              size="sm"
              variant="outline"
            />
          </Menu.Button>
          <Menu.Items align="start" className="w-80 py-1">
            <Box className="px-4 py-2">
              <Menu.Input autoFocus placeholder="Add filter..." />
            </Box>
            <Menu.Separator className="my-0" />
            <Menu.Group className="max-h-96 overflow-y-auto px-1 py-1.5">
              {filterOptions.map((option) => {
                const isActive = chips.some(
                  (chip) => chip.field === option.field,
                );
                const shouldOpenDialog = option.field === "contentContains";

                if (shouldOpenDialog) {
                  return (
                    <Menu.Item
                      active={isActive}
                      className="justify-between gap-4"
                      key={option.field}
                      onSelect={() => {
                        setTitleDialogOpen(true);
                      }}
                    >
                      <Box className="grid min-w-0 flex-1 grid-cols-[24px_minmax(0,1fr)] items-center">
                        <span className="text-text-secondary flex h-6 w-6 shrink-0 items-center">
                          {option.icon}
                        </span>
                        <Text className="truncate">{option.label}</Text>
                      </Box>
                    </Menu.Item>
                  );
                }

                return (
                  <Menu.SubMenu key={option.field}>
                    <Menu.SubTrigger
                      active={isActive}
                      className="justify-between gap-4"
                    >
                      <Box className="grid min-w-0 flex-1 grid-cols-[24px_minmax(0,1fr)] items-center">
                        <span className="text-text-secondary flex h-6 w-6 shrink-0 items-center">
                          {option.icon}
                        </span>
                        <Text className="truncate">{option.label}</Text>
                      </Box>
                      <Flex align="center" className="shrink-0" gap={1}>
                        <ArrowRightIcon
                          className="text-text-muted h-3.5 w-auto"
                          strokeWidth={2.8}
                        />
                      </Flex>
                    </Menu.SubTrigger>
                    <Menu.SubItems
                      alignOffset={-6}
                      className={getEditorContentClassName(option.field)}
                      sideOffset={8}
                    >
                      <FilterValueEditor
                        estimateScheme={estimateScheme}
                        field={option.field}
                        filters={filters}
                        setFilters={setFilters}
                      />
                    </Menu.SubItems>
                  </Menu.SubMenu>
                );
              })}
            </Menu.Group>
          </Menu.Items>
        </Menu>
        {titleDialogOpen ? (
          <TitleFilterDialog
            filters={filters}
            key={filters.contentContains ?? ""}
            onOpenChange={setTitleDialogOpen}
            open={titleDialogOpen}
            setFilters={setFilters}
          />
        ) : null}
      </Flex>
      {hasActiveStoriesFilters(filters) ? (
        <Flex align="center" className="shrink-0" gap={2}>
          <Button
            color="tertiary"
            onClick={resetFilters}
            size="sm"
            variant="outline"
          >
            Clear all
          </Button>
        </Flex>
      ) : null}
    </Flex>
  );
};
