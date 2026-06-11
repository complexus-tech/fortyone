"use client";
import { useMemo, useState, type ReactNode } from "react";
import {
  Avatar,
  Box,
  Button,
  Command,
  DatePicker,
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
  ListIcon,
  ObjectiveIcon,
  PlusIcon,
  SprintsIcon,
  TeamIcon,
  UserIcon,
} from "icons";
import { format, formatISO } from "date-fns";
import { useParams } from "next/navigation";
import { useStatuses } from "@/lib/hooks/statuses";
import { useMembers } from "@/lib/hooks/members";
import { useTeamMembers } from "@/lib/hooks/team-members";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useTeamSprints } from "@/modules/sprints/hooks/team-sprints";
import { useTeamObjectives } from "@/modules/objectives/hooks/use-objectives";
import type { StoryPriority } from "@/modules/stories/types";
import type { StoriesFilter } from "./stories-filter-button";
import { PriorityIcon } from "./priority-icon";
import { StoryStatusIcon } from "./story-status-icon";
import { TeamColor } from "./team-color";
import { hasActiveStoriesFilters } from "./stories-filter-utils";

type FilterField =
  | "titleContains"
  | "statusIds"
  | "assigneeIds"
  | "reporterIds"
  | "priorities"
  | "teamIds"
  | "sprintIds"
  | "objectiveId"
  | "startDate"
  | "endDate"
  | "assignedToMe"
  | "createdByMe"
  | "hasNoAssignee";

type FilterChip = {
  field: FilterField;
  label: string;
  operator: string;
  value: string;
  icon?: ReactNode;
};

type StoriesFilterBarProps = {
  filters: StoriesFilter;
  setFilters: (value: StoriesFilter) => void;
  resetFilters: () => void;
};

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

const getEditorContentClassName = (field: FilterField) => {
  if (field === "titleContains") {
    return "w-80 overflow-hidden py-2";
  }

  if (field === "objectiveId") {
    return "w-96 overflow-hidden py-2";
  }

  if (field === "assigneeIds" || field === "reporterIds") {
    return "w-80 overflow-hidden py-2";
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
  const [draft, setDraft] = useState(filters.titleContains ?? "");

  const applyTitleFilter = () => {
    const titleContains = draft.trim();
    setFilters({
      ...filters,
      titleContains: titleContains ? titleContains : null,
    });
    onOpenChange(false);
  };

  return (
    <Dialog onOpenChange={onOpenChange} open={open}>
      <Dialog.Content className="max-w-2xl" hideClose>
        <Dialog.Header className="px-6 pt-6">
          <Dialog.Title className="text-lg">Filter by title</Dialog.Title>
        </Dialog.Header>
        <Dialog.Body className="pt-0">
          <Input
            autoFocus
            leftIcon={<ListIcon className="text-text-secondary h-4 w-auto" />}
            onChange={(event) => {
              setDraft(event.target.value);
            }}
            onKeyDown={(event) => {
              if (event.key === "Enter") {
                applyTitleFilter();
              }
            }}
            placeholder="Title contains..."
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
  const { data: workspaceMembers = [] } = useMembers(query);
  const { data: teamMembers = [] } = useTeamMembers(teamId, query);
  const members = teamId ? teamMembers : workspaceMembers;

  const toggleMember = (memberId: string) => {
    const selected = filters[field] ?? [];
    const memberIds = selected.includes(memberId)
      ? selected.filter((id) => id !== memberId)
      : [...selected, memberId];
    setFilters({ ...filters, [field]: normalizeArrayFilter(memberIds) });
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
      <Command.Empty className="py-2">
        <Text color="muted">No user found.</Text>
      </Command.Empty>
      <Command.Group className="max-h-80 overflow-y-auto md:max-h-100">
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
  const { data: teams = [] } = useTeams();
  const filteredTeams = teams.filter((team) =>
    team.name.toLowerCase().includes(query.toLowerCase()),
  );

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
      <Command.Empty className="py-2">
        <Text color="muted">No teams found.</Text>
      </Command.Empty>
      <Command.Group className="max-h-80 overflow-y-auto">
        {filteredTeams.map((team, idx) => (
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
  const { data: sprints = [] } = useTeamSprints(teamId ?? "", query);

  const toggleSprint = (sprintId: string) => {
    const selected = filters.sprintIds ?? [];
    const sprintIds = selected.includes(sprintId)
      ? selected.filter((id) => id !== sprintId)
      : [...selected, sprintId];
    setFilters({ ...filters, sprintIds: normalizeArrayFilter(sprintIds) });
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
      <Command.Empty className="py-2">
        <Text color="muted">No sprints found.</Text>
      </Command.Empty>
      <Command.Group className="max-h-80 overflow-y-auto">
        {sprints.map((sprint, idx) => (
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
        ))}
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
  const { data: objectives = [] } = useTeamObjectives(teamId ?? "", query);

  return (
    <Command>
      <Command.Input
        autoFocus
        onValueChange={setQuery}
        placeholder="Search objectives..."
        value={query}
      />
      <Divider className="my-2" />
      <Command.Empty className="py-2">
        <Text color="muted">No objectives found.</Text>
      </Command.Empty>
      <Command.Group className="max-h-80 overflow-y-auto">
        {objectives.map((objective, idx) => (
          <Command.Item
            active={filters.objectiveId === objective.id}
            className="justify-between gap-4"
            key={objective.id}
            onSelect={() => {
              setFilters({
                ...filters,
                objectiveId:
                  filters.objectiveId === objective.id ? null : objective.id,
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
    <Box className="px-3 py-1">
      <DatePicker>
        <DatePicker.Trigger>
          <Button
            className="w-full justify-start"
            color="tertiary"
            leftIcon={<CalendarIcon className="h-4 w-auto" />}
            size="sm"
            variant="outline"
          >
            {selectedDate ? format(selectedDate, "MMM d, yyyy") : "Select date"}
          </Button>
        </DatePicker.Trigger>
        <DatePicker.Calendar
          mode="single"
          onDayClick={(date) => {
            setFilters({
              ...filters,
              [field]: formatISO(date, { representation: "date" }),
            });
          }}
          selected={selectedDate}
        />
      </DatePicker>
      {selectedDate ? (
        <Button
          className="mt-2 w-full justify-start"
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
  filters,
  setFilters,
}: {
  field: FilterField;
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

  if (field === "startDate" || field === "endDate") {
    return (
      <DateEditor field={field} filters={filters} setFilters={setFilters} />
    );
  }

  return null;
};

const Chip = ({
  chip,
  filters,
  setFilters,
  onEditTitle,
  onRemove,
}: {
  chip: FilterChip;
  filters: StoriesFilter;
  setFilters: (value: StoriesFilter) => void;
  onEditTitle: () => void;
  onRemove: () => void;
}) => {
  const isEditable =
    chip.field !== "assignedToMe" &&
    chip.field !== "createdByMe" &&
    chip.field !== "hasNoAssignee";
  const shouldUseDialog = chip.field === "titleContains";

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
      {shouldUseDialog ? (
        <button
          className="hover:bg-state-hover flex h-full max-w-72 items-center truncate px-2.5 text-left transition"
          onClick={onEditTitle}
          type="button"
        >
          <span className="truncate">{chip.value}</span>
        </button>
      ) : isEditable ? (
        <Popover>
          <Popover.Trigger asChild>
            <button
              className="hover:bg-state-hover flex h-full max-w-72 items-center truncate px-2.5 text-left transition"
              type="button"
            >
              <span className="truncate">{chip.value}</span>
            </button>
          </Popover.Trigger>
          <Popover.Content
            align="start"
            className={getEditorContentClassName(chip.field)}
          >
            <FilterValueEditor
              field={chip.field}
              filters={filters}
              setFilters={setFilters}
            />
          </Popover.Content>
        </Popover>
      ) : (
        <span className="flex h-full items-center px-2.5">{chip.value}</span>
      )}
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
}: StoriesFilterBarProps) => {
  const { teamId } = useParams<{ teamId?: string }>();
  const [titleDialogOpen, setTitleDialogOpen] = useState(false);
  const { data: allStatuses = [] } = useStatuses();
  const { data: allUsers = [] } = useMembers();
  const resolvedTeamId = teamId ?? "";
  const { data: teamMembers = [] } = useTeamMembers(resolvedTeamId);
  const { data: teams = [] } = useTeams();
  const { data: sprints = [] } = useTeamSprints(resolvedTeamId);
  const { data: objectives = [] } = useTeamObjectives(resolvedTeamId);

  const users = teamId ? teamMembers : allUsers;
  const statuses = teamId
    ? allStatuses.filter((status) => status.teamId === teamId)
    : allStatuses;

  const statusById = useMemo(
    () => new Map(statuses.map((status) => [status.id, status.name])),
    [statuses],
  );
  const userById = useMemo(
    () =>
      new Map(
        users.map((user) => [
          user.id,
          user.fullName || user.username || "Unknown user",
        ]),
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

  const chips = useMemo(() => {
    const items: FilterChip[] = [];

    if (filters.titleContains?.trim()) {
      items.push({
        field: "titleContains",
        label: "Title",
        operator: "contains",
        value: filters.titleContains.trim(),
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
      items.push({
        field: "statusIds",
        label: "Status",
        operator: "is any of",
        value: getNames(filters.statusIds, statusById),
        icon: <StoryStatusIcon statusId={filters.statusIds[0]} />,
      });
    }

    if (filters.assigneeIds?.length) {
      items.push({
        field: "assigneeIds",
        label: "Assignee",
        operator: "is any of",
        value: getNames(filters.assigneeIds, userById),
      });
    }

    if (filters.reporterIds?.length) {
      items.push({
        field: "reporterIds",
        label: "Reporter",
        operator: "is any of",
        value: getNames(filters.reporterIds, userById),
      });
    }

    if (filters.priorities?.length) {
      items.push({
        field: "priorities",
        label: "Priority",
        operator: "is any of",
        value: filters.priorities.join(", "),
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
    objectiveById,
    sprintById,
    statusById,
    teamById,
    teamColorById,
    userById,
  ]);

  const removeFilter = (field: FilterField) => {
    if (field === "assignedToMe" || field === "createdByMe") {
      setFilters({ ...filters, [field]: false });
      return;
    }

    setFilters({ ...filters, [field]: null });
  };

  const filterOptions: {
    field: FilterField;
    icon: ReactNode;
    label: string;
  }[] = [
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
    ...(teamId
      ? []
      : [
          {
            field: "teamIds" as const,
            icon: <TeamIcon className="h-5 w-auto" />,
            label: "Team",
          },
        ]),
    {
      field: "sprintIds",
      icon: <SprintsIcon className="h-5 w-auto" />,
      label: "Sprint",
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
      field: "titleContains",
      icon: <ListIcon className="h-5 w-auto" />,
      label: "Title",
    },
  ];

  if (!hasActiveStoriesFilters(filters)) {
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
            filters={filters}
            key={chip.field}
            onRemove={() => {
              removeFilter(chip.field);
            }}
            onEditTitle={() => {
              setTitleDialogOpen(true);
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
                const shouldOpenDialog = option.field === "titleContains";

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
            key={filters.titleContains ?? ""}
            onOpenChange={setTitleDialogOpen}
            open={titleDialogOpen}
            setFilters={setFilters}
          />
        ) : null}
      </Flex>
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
    </Flex>
  );
};
