"use client";
import { Avatar, Box, Button, Divider, Flex, Popover, Text, Tooltip } from "ui";
import {
  ArrowDownIcon,
  AssigneeIcon,
  CheckIcon,
  FilterIcon,
  UserIcon,
} from "icons";
import { useRef, type ReactNode } from "react";
import { cn } from "lib";
import { useHotkeys } from "react-hotkeys-hook";
import { useParams } from "next/navigation";
import { format } from "date-fns";
import { useStatuses } from "@/lib/hooks/statuses";
import type { StoryPriority } from "@/modules/stories/types";
import { hexToRgba } from "@/utils";
import { useMembers } from "@/lib/hooks/members";
import { useTeamMembers } from "@/lib/hooks/team-members";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useTeamSprints } from "@/modules/sprints/hooks/team-sprints";
import { PriorityIcon } from "./priority-icon";
import { StoryStatusIcon } from "./story-status-icon";
import { TeamColor } from "./team-color";

export type StoriesFilter = {
  statusIds: string[] | null;
  assigneeIds: string[] | null;
  reporterIds: string[] | null;
  priorities: string[] | null;
  teamIds: string[] | null;
  sprintIds: string[] | null;
  labelIds: string[] | null;
  parentId: string | null;
  objectiveId: string | null;
  epicId: string | null;
  keyResultId: string | null;
  hasNoAssignee: boolean | null;
  assignedToMe: boolean;
  createdByMe: boolean;
};

type StoriesFilterButtonProps = {
  filters: StoriesFilter;
  setFilters: (v: StoriesFilter) => void;
  resetFilters: () => void;
};

const FilterSection = ({
  title,
  children,
}: {
  title: string;
  children: ReactNode;
}) => (
  <Box className="px-4 py-3">
    <Text className="mb-2.5" fontWeight="medium">
      {title}
    </Text>
    {children}
  </Box>
);

const ToggleButton = ({
  label,
  icon,
  isActive,
  onClick,
}: {
  label: string;
  isActive: boolean;
  icon: ReactNode;
  onClick: () => void;
}) => {
  return (
    <button
      className={cn(
        "flex w-full items-center justify-between px-4 py-3 transition hover:bg-gray-50 hover:dark:bg-dark-50/40",
      )}
      onClick={onClick}
      type="button"
    >
      <span className="flex items-center gap-2">
        {icon}
        {label}
      </span>
      {isActive ? <CheckIcon className="h-5 w-auto text-primary" /> : null}
    </button>
  );
};

const StatusSelector = ({
  selected,
  onChange,
}: {
  selected: string[] | null;
  onChange: (ids: string[]) => void;
}) => {
  const { teamId } = useParams<{ teamId?: string }>();
  const { data: allStatuses = [] } = useStatuses();
  const statuses = teamId
    ? allStatuses.filter((status) => status.teamId === teamId)
    : allStatuses;

  const toggleStatus = (statusId: string) => {
    const current = selected || [];
    if (current.includes(statusId)) {
      onChange(current.filter((id) => id !== statusId));
    } else {
      onChange([...current, statusId]);
    }
  };

  return (
    <Flex gap={2} wrap>
      {statuses.map((status) => (
        <Button
          className={cn("ring-2 ring-transparent", {
            "ring-primary": selected?.includes(status.id),
          })}
          color="tertiary"
          key={status.id}
          leftIcon={<StoryStatusIcon statusId={status.id} />}
          onClick={() => {
            toggleStatus(status.id);
          }}
          size="sm"
          style={{
            backgroundColor: hexToRgba(status.color, 0.1),
            borderColor: hexToRgba(status.color, 0.2),
          }}
        >
          {status.name}
        </Button>
      ))}
    </Flex>
  );
};

const UserSelector = ({
  selected,
  onChange,
}: {
  selected: string[] | null;
  onChange: (ids: string[]) => void;
}) => {
  const { teamId } = useParams<{ teamId?: string }>();
  const { data: allUsers = [] } = useMembers();
  const { data: teamMembers = [] } = useTeamMembers(teamId);
  const users = teamId ? teamMembers : allUsers;

  const toggleUser = (userId: string) => {
    const current = selected || [];
    if (current.includes(userId)) {
      onChange(current.filter((id) => id !== userId));
    } else {
      onChange([...current, userId]);
    }
  };

  return (
    <Flex gap={2} wrap>
      {users.map((user) => (
        <button
          className={cn("relative rounded-full ring-2 ring-transparent", {
            "ring-primary": selected?.includes(user.id),
          })}
          key={user.id}
          onClick={() => {
            toggleUser(user.id);
          }}
          type="button"
        >
          <Avatar className="h-10" name={user.fullName} src={user.avatarUrl} />
          <span className="sr-only">{user.fullName || user.username}</span>
        </button>
      ))}
    </Flex>
  );
};

const PrioritySelector = ({
  selected,
  onChange,
}: {
  selected: string[] | null;
  onChange: (priorities: string[]) => void;
}) => {
  const priorities = [
    "Urgent",
    "High",
    "Medium",
    "Low",
    "No Priority",
  ] as StoryPriority[];

  const togglePriority = (priority: string) => {
    const current = selected || [];
    if (current.includes(priority)) {
      onChange(current.filter((p) => p !== priority));
    } else {
      onChange([...current, priority]);
    }
  };

  return (
    <Flex gap={2} wrap>
      {priorities.map((priority) => (
        <Button
          className={cn("ring-2 ring-transparent", {
            "ring-primary": selected?.includes(priority),
          })}
          color="tertiary"
          key={priority}
          leftIcon={<PriorityIcon priority={priority} />}
          onClick={() => {
            togglePriority(priority);
          }}
          size="sm"
        >
          {priority}
        </Button>
      ))}
    </Flex>
  );
};

const TeamSelector = ({
  selected,
  onChange,
}: {
  selected: string[] | null;
  onChange: (teamIds: string[]) => void;
}) => {
  const { data: teams = [] } = useTeams();

  const toggleTeam = (teamId: string) => {
    const current = selected || [];
    if (current.includes(teamId)) {
      onChange(current.filter((id) => id !== teamId));
    } else {
      onChange([...current, teamId]);
    }
  };

  return (
    <Flex gap={2} wrap>
      {teams.map((team) => (
        <Button
          className={cn({
            "ring-2 ring-primary": selected?.includes(team.id),
          })}
          color="tertiary"
          key={team.id}
          leftIcon={<TeamColor color={team.color} />}
          onClick={() => {
            toggleTeam(team.id);
          }}
          size="sm"
          style={{
            backgroundColor: hexToRgba(team.color, 0.1),
            borderColor: hexToRgba(team.color, 0.2),
          }}
          variant={selected?.includes(team.id) ? "solid" : "outline"}
        >
          {team.name}
        </Button>
      ))}
    </Flex>
  );
};

const SprintSelector = ({
  selected,
  onChange,
}: {
  selected: string[] | null;
  onChange: (sprintIds: string[]) => void;
}) => {
  const { teamId } = useParams<{ teamId: string }>();
  const { data: sprints = [] } = useTeamSprints(teamId);
  const toggleSprint = (sprintId: string) => {
    const current = selected || [];
    if (current.includes(sprintId)) {
      onChange(current.filter((id) => id !== sprintId));
    } else {
      onChange([...current, sprintId]);
    }
  };

  return (
    <Flex gap={2} wrap>
      {sprints.map((sprint) => {
        const startDate = format(new Date(sprint.startDate), "MMM d");
        const endDate = format(new Date(sprint.endDate), "MMM d");
        const sprintName = `${sprint.name} (${startDate} - ${endDate})`;
        return (
          <Tooltip key={sprint.id} title={sprintName}>
            <Button
              className={cn("ring-2 ring-transparent", {
                "ring-primary": selected?.includes(sprint.id),
              })}
              color="tertiary"
              onClick={() => {
                toggleSprint(sprint.id);
              }}
              size="sm"
            >
              {sprint.name}
            </Button>
          </Tooltip>
        );
      })}
    </Flex>
  );
};

export const StoriesFilterButton = ({
  filters,
  setFilters,
  resetFilters,
}: StoriesFilterButtonProps) => {
  const buttonRef = useRef<HTMLButtonElement>(null);

  // filtersCount returns the number of filters applied.
  const filtersCount = () => {
    let count = 0;

    // Count array filters
    const arrayFilters = [
      "statusIds",
      "assigneeIds",
      "reporterIds",
      "priorities",
      "teamIds",
      "sprintIds",
      "labelIds",
    ] as const;
    arrayFilters.forEach((key) => {
      if (filters[key] && filters[key].length > 0) count++;
    });

    // Count string filters
    const stringFilters = [
      "parentId",
      "objectiveId",
      "epicId",
      "keyResultId",
    ] as const;
    stringFilters.forEach((key) => {
      if (filters[key]) count++;
    });

    // Count boolean filters
    if (filters.hasNoAssignee) count++;
    if (filters.assignedToMe) count++;
    if (filters.createdByMe) count++;

    return count;
  };

  const getButtonLabel = () => {
    if (filtersCount()) {
      return `${filtersCount()} filter${filtersCount() > 1 ? "s" : ""} applied`;
    }
    return "Filters";
  };

  useHotkeys("v+f", (e) => {
    e.preventDefault();
    buttonRef.current?.click();
  });

  return (
    <Popover>
      <Popover.Trigger asChild>
        <Button
          color="tertiary"
          leftIcon={<FilterIcon className="h-4 w-auto" />}
          ref={buttonRef}
          rightIcon={<ArrowDownIcon className="h-3.5 w-auto" />}
          size="sm"
          variant="outline"
        >
          <span className="hidden md:inline">{getButtonLabel()}</span>
        </Button>
      </Popover.Trigger>
      <Popover.Content className="max-h-[87vh] w-80 overflow-y-auto rounded-[1.25rem] pb-2 dark:bg-dark-200/90 md:w-[35rem]">
        <Flex align="center" className="h-11 px-4" justify="between">
          <Text
            color="muted"
            fontSize="sm"
            fontWeight="semibold"
            transform="uppercase"
          >
            Apply Filters
          </Text>
          {filtersCount() > 0 && (
            <Button
              className="text-primary dark:text-primary"
              color="tertiary"
              onClick={resetFilters}
              size="sm"
              variant="naked"
            >
              Clear filters
            </Button>
          )}
        </Flex>
        <Divider className="mt-1.5" />
        <Box>
          <ToggleButton
            icon={<AssigneeIcon />}
            isActive={filters.assignedToMe}
            label="Assigned to me"
            onClick={() => {
              setFilters({ ...filters, assignedToMe: !filters.assignedToMe });
            }}
          />
          <ToggleButton
            icon={<UserIcon />}
            isActive={filters.createdByMe}
            label="Created by me"
            onClick={() => {
              setFilters({ ...filters, createdByMe: !filters.createdByMe });
            }}
          />
          <ToggleButton
            icon={<AssigneeIcon />}
            isActive={filters.hasNoAssignee || false}
            label="Has no assignee"
            onClick={() => {
              setFilters({ ...filters, hasNoAssignee: !filters.hasNoAssignee });
            }}
          />
        </Box>
        <Divider />
        <FilterSection title="Status">
          <StatusSelector
            onChange={(statusIds) => {
              setFilters({ ...filters, statusIds });
            }}
            selected={filters.statusIds}
          />
        </FilterSection>
        <Divider />
        <FilterSection title="Assignee">
          <UserSelector
            onChange={(assigneeIds) => {
              setFilters({ ...filters, assigneeIds });
            }}
            selected={filters.assigneeIds}
          />
        </FilterSection>
        <FilterSection title="Reporter">
          <UserSelector
            onChange={(reporterIds) => {
              setFilters({ ...filters, reporterIds });
            }}
            selected={filters.reporterIds}
          />
        </FilterSection>
        <Divider />
        <FilterSection title="Priority">
          <PrioritySelector
            onChange={(priorities) => {
              setFilters({ ...filters, priorities });
            }}
            selected={filters.priorities}
          />
        </FilterSection>
        <Divider />
        <FilterSection title="Team">
          <TeamSelector
            onChange={(teamIds) => {
              setFilters({ ...filters, teamIds });
            }}
            selected={filters.teamIds}
          />
        </FilterSection>
        <Divider />
        <FilterSection title="Sprint">
          <SprintSelector
            onChange={(sprintIds) => {
              setFilters({ ...filters, sprintIds });
            }}
            selected={filters.sprintIds}
          />
        </FilterSection>
      </Popover.Content>
    </Popover>
  );
};
