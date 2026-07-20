"use client";
import { Avatar, Box, Button, Divider, Flex, Popover, Text, Tooltip } from "ui";
import { ArrowDownIcon, AssigneeIcon, CheckIcon, FilterIcon } from "icons";
import { useRef, type ReactNode } from "react";
import { cn } from "lib";
import { useHotkeys } from "react-hotkeys-hook";
import { useParams, usePathname } from "next/navigation";
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
import { MemberTooltip } from "./member-tooltip";
import type { StoriesFilter } from "./stories-filter-types";
import {
  getActiveStoriesFilterCount,
  hasActiveStoriesFilters,
} from "./stories-filter-utils";
import type { StoriesFilterField } from "./stories-filter-bar";
import {
  getVisibleStoriesFilterButtonFields,
  type StoriesFilterButtonField,
} from "./stories-filter-button-options";

type StoriesFilterButtonProps = {
  filters: StoriesFilter;
  setFilters: (v: StoriesFilter) => void;
  resetFilters: () => void;
  iconOnly?: boolean;
  hiddenFields?: readonly StoriesFilterField[];
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
        "hover:bg-state-hover flex w-full items-center justify-between px-4 py-3 transition",
      )}
      onClick={onClick}
      type="button"
    >
      <span className="flex items-center gap-2">
        {icon}
        {label}
      </span>
      {isActive ? <CheckIcon className="text-primary h-5 w-auto" /> : null}
    </button>
  );
};

export const StatusSelector = ({
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

export const UserSelector = ({
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
        <MemberTooltip key={user.id} member={user}>
          <button
            className={cn("relative rounded-full ring-2 ring-transparent", {
              "ring-primary": selected?.includes(user.id),
            })}
            onClick={() => {
              toggleUser(user.id);
            }}
            type="button"
          >
            <Avatar
              className="h-10"
              name={user.fullName}
              src={user.avatarUrl}
            />
            <span className="sr-only">{user.fullName || user.username}</span>
          </button>
        </MemberTooltip>
      ))}
    </Flex>
  );
};

export const PrioritySelector = ({
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

export const TeamSelector = ({
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
            "ring-primary ring-2": selected?.includes(team.id),
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

export const SprintSelector = ({
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
      {sprints.slice(0, 10).map((sprint) => {
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
      {sprints.length === 0 && <Text color="muted">No sprints</Text>}
    </Flex>
  );
};

export const StoriesFilterButton = ({
  filters,
  setFilters,
  resetFilters,
  iconOnly = false,
  hiddenFields = [],
}: StoriesFilterButtonProps) => {
  const buttonRef = useRef<HTMLButtonElement>(null);
  const pathname = usePathname();
  const { teamId } = useParams<{ teamId: string }>();
  const isBacklog = pathname.includes("/backlog");
  const visibleFilters = { ...filters };
  hiddenFields.forEach((field) => {
    if (field === "assignedToMe" || field === "createdByMe") {
      visibleFilters[field] = false;
      return;
    }

    visibleFilters[field] = null;
  });
  const filtersCount = getActiveStoriesFilterCount(visibleFilters);
  const visibleFields = getVisibleStoriesFilterButtonFields({
    hasRouteTeam: Boolean(teamId),
    hiddenFields,
  });
  const shouldShow = (field: StoriesFilterButtonField) =>
    visibleFields.includes(field);

  const getButtonLabel = () => {
    if (filtersCount) {
      return `${filtersCount} filter${filtersCount > 1 ? "s" : ""} applied`;
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
          aria-label={getButtonLabel()}
          className="relative"
          color="tertiary"
          leftIcon={<FilterIcon className="h-4 w-auto" />}
          ref={buttonRef}
          rightIcon={
            iconOnly ? undefined : <ArrowDownIcon className="h-3.5 w-auto" />
          }
          size="sm"
          variant="outline"
        >
          {hasActiveStoriesFilters(visibleFilters) ? (
            <span
              aria-hidden="true"
              className="bg-primary absolute -top-0.5 -right-0.5 h-2.5 w-2.5 rounded-full"
            >
              <span className="bg-primary absolute inset-0 animate-ping rounded-full opacity-75" />
            </span>
          ) : null}
          {iconOnly ? null : (
            <span className="hidden md:inline">{getButtonLabel()}</span>
          )}
        </Button>
      </Popover.Trigger>
      <Popover.Content
        align="end"
        className="bg-surface-elevated dark:bg-surface-elevated/80 mr-0 max-h-[87vh] w-80 overflow-y-auto rounded-2xl pb-2 md:w-140"
      >
        <Flex align="center" className="h-11 px-4" justify="between">
          <Text
            color="muted"
            fontSize="sm"
            fontWeight="semibold"
            transform="uppercase"
          >
            Apply Filters
          </Text>
          {filtersCount > 0 && (
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
        {shouldShow("hasNoAssignee") ? (
          <Box>
            <ToggleButton
              icon={<AssigneeIcon />}
              isActive={filters.hasNoAssignee || false}
              label="Has no assignee"
              onClick={() => {
                setFilters({
                  ...filters,
                  hasNoAssignee: !filters.hasNoAssignee,
                });
              }}
            />
          </Box>
        ) : null}
        {!isBacklog && shouldShow("statusIds") ? (
          <>
            <Divider />
            <FilterSection title="Status">
              <StatusSelector
                onChange={(statusIds) => {
                  setFilters({ ...filters, statusIds });
                }}
                selected={filters.statusIds}
              />
            </FilterSection>
          </>
        ) : null}

        {shouldShow("assigneeIds") || shouldShow("reporterIds") ? (
          <>
            <Divider />
            {shouldShow("assigneeIds") ? (
              <FilterSection title="Assignee">
                <UserSelector
                  onChange={(assigneeIds) => {
                    setFilters({ ...filters, assigneeIds });
                  }}
                  selected={filters.assigneeIds}
                />
              </FilterSection>
            ) : null}
            {shouldShow("reporterIds") ? (
              <FilterSection title="Reporter">
                <UserSelector
                  onChange={(reporterIds) => {
                    setFilters({ ...filters, reporterIds });
                  }}
                  selected={filters.reporterIds}
                />
              </FilterSection>
            ) : null}
          </>
        ) : null}
        {shouldShow("priorities") ? (
          <>
            <Divider />
            <FilterSection title="Priority">
              <PrioritySelector
                onChange={(priorities) => {
                  setFilters({ ...filters, priorities });
                }}
                selected={filters.priorities}
              />
            </FilterSection>
          </>
        ) : null}
        {shouldShow("teamIds") ? (
          <>
            <Divider />
            <FilterSection title="Team">
              <TeamSelector
                onChange={(teamIds) => {
                  setFilters({ ...filters, teamIds });
                }}
                selected={filters.teamIds}
              />
            </FilterSection>
          </>
        ) : null}
      </Popover.Content>
    </Popover>
  );
};
