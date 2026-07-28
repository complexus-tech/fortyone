"use client";

import type { ReactNode } from "react";
import { format, formatISO } from "date-fns";
import { cn } from "lib";
import {
  CalendarIcon,
  CheckIcon,
  FilterIcon,
  ObjectiveIcon,
  OKRIcon,
  TeamIcon,
} from "icons";
import { Avatar, Box, Button, Divider, Flex, Input, Popover, Text } from "ui";
import { useMembers } from "@/lib/hooks/members";
import { useObjectives } from "@/modules/objectives/hooks/use-objectives";
import { useTeams } from "@/modules/teams/hooks/teams";
import type { KeyResultFilters } from "../types";

const toggleValue = (values: string[] | undefined, value: string) => {
  const selected = values ?? [];
  return selected.includes(value)
    ? selected.filter((item) => item !== value)
    : [...selected, value];
};

const FilterOption = ({
  active,
  icon,
  label,
  onClick,
}: {
  active: boolean;
  icon: ReactNode;
  label: string;
  onClick: () => void;
}) => (
  <button
    className="hover:bg-state-hover flex w-full items-center justify-between gap-3 px-4 py-2.5 text-left transition-colors"
    onClick={onClick}
    type="button"
  >
    <Flex align="center" className="min-w-0" gap={2}>
      <span className="text-text-muted flex shrink-0">{icon}</span>
      <Text className="truncate whitespace-nowrap">{label}</Text>
    </Flex>
    {active ? <CheckIcon className="text-primary h-4 shrink-0" /> : null}
  </button>
);

const FilterSection = ({
  children,
  title,
}: {
  children: ReactNode;
  title: string;
}) => (
  <Box>
    <Text className="px-4 pt-3 pb-1.5" color="muted" fontWeight="medium">
      {title}
    </Text>
    {children}
  </Box>
);

export const getActiveKeyResultFilterCount = (filters: KeyResultFilters) =>
  Number(Boolean(filters.teamIds?.length)) +
  Number(Boolean(filters.objectiveIds?.length)) +
  Number(Boolean(filters.measurementTypes?.length)) +
  Number(Boolean(filters.leadIds?.length)) +
  Number(Boolean(filters.endDateAfter || filters.endDateBefore));

export const KeyResultsFilterButton = ({
  filters,
  setFilters,
}: {
  filters: KeyResultFilters;
  setFilters: (filters: KeyResultFilters) => void;
}) => {
  const { data: teams = [] } = useTeams();
  const { data: objectives = [] } = useObjectives();
  const { data: members = [] } = useMembers();
  const activeFilterCount = getActiveKeyResultFilterCount(filters);

  return (
    <Popover>
      <Popover.Trigger asChild>
        <Button
          color="tertiary"
          leftIcon={<FilterIcon className="h-[1.1rem]" />}
          size="sm"
          variant="outline"
        >
          Filters
          {activeFilterCount > 0 ? (
            <span className="bg-primary text-on-primary ml-0.5 flex size-5 items-center justify-center rounded-full text-xs">
              {activeFilterCount}
            </span>
          ) : null}
        </Button>
      </Popover.Trigger>
      <Popover.Content align="end" className="w-80 overflow-hidden p-0">
        <Flex align="center" className="px-4 py-3" justify="between">
          <Text fontWeight="medium">Filter key results</Text>
          {activeFilterCount > 0 ? (
            <Button
              color="tertiary"
              onClick={() => {
                setFilters({});
              }}
              size="xs"
              variant="naked"
            >
              Clear all
            </Button>
          ) : null}
        </Flex>
        <Divider />
        <Box className="max-h-[30rem] overflow-y-auto pb-2">
          <FilterSection title="Lead">
            {members.map((member) => (
              <FilterOption
                active={Boolean(filters.leadIds?.includes(member.id))}
                icon={
                  <Avatar
                    name={member.fullName || member.username}
                    size="xs"
                    src={member.avatarUrl}
                  />
                }
                key={member.id}
                label={member.fullName || member.username}
                onClick={() => {
                  const leadIds = toggleValue(filters.leadIds, member.id);
                  setFilters({
                    ...filters,
                    leadIds: leadIds.length > 0 ? leadIds : undefined,
                  });
                }}
              />
            ))}
          </FilterSection>
          <Divider className="my-2" />
          <FilterSection title="Delivery date">
            <Box className="grid grid-cols-2 gap-2 px-4 pb-3">
              <Input
                aria-label="Delivery date from"
                onChange={(event) => {
                  setFilters({
                    ...filters,
                    endDateAfter: event.target.value
                      ? formatISO(new Date(event.target.value))
                      : undefined,
                  });
                }}
                type="date"
                value={
                  filters.endDateAfter
                    ? format(new Date(filters.endDateAfter), "yyyy-MM-dd")
                    : ""
                }
              />
              <Input
                aria-label="Delivery date to"
                onChange={(event) => {
                  setFilters({
                    ...filters,
                    endDateBefore: event.target.value
                      ? formatISO(new Date(event.target.value))
                      : undefined,
                  });
                }}
                type="date"
                value={
                  filters.endDateBefore
                    ? format(new Date(filters.endDateBefore), "yyyy-MM-dd")
                    : ""
                }
              />
            </Box>
          </FilterSection>
          <Divider className="my-2" />
          <FilterSection title="Measurement type">
            {(
              [
                ["percentage", "Percentage"],
                ["number", "Number"],
                ["boolean", "Complete / incomplete"],
              ] as const
            ).map(([value, label]) => (
              <FilterOption
                active={Boolean(filters.measurementTypes?.includes(value))}
                icon={<OKRIcon className="h-4" />}
                key={value}
                label={label}
                onClick={() => {
                  const measurementTypes = toggleValue(
                    filters.measurementTypes,
                    value,
                  ) as NonNullable<KeyResultFilters["measurementTypes"]>;
                  setFilters({
                    ...filters,
                    measurementTypes:
                      measurementTypes.length > 0
                        ? measurementTypes
                        : undefined,
                  });
                }}
              />
            ))}
          </FilterSection>
          {teams.length > 0 ? (
            <>
              <Divider className="my-2" />
              <FilterSection title="Teams">
                {teams.map((team) => (
                  <FilterOption
                    active={Boolean(filters.teamIds?.includes(team.id))}
                    icon={
                      <TeamIcon
                        className={cn("h-4", {
                          "text-foreground": filters.teamIds?.includes(team.id),
                        })}
                      />
                    }
                    key={team.id}
                    label={team.name}
                    onClick={() => {
                      const teamIds = toggleValue(filters.teamIds, team.id);
                      setFilters({
                        ...filters,
                        teamIds: teamIds.length > 0 ? teamIds : undefined,
                      });
                    }}
                  />
                ))}
              </FilterSection>
            </>
          ) : null}
          {objectives.length > 0 ? (
            <>
              <Divider className="my-2" />
              <FilterSection title="Objectives">
                {objectives.map((objective) => (
                  <FilterOption
                    active={Boolean(
                      filters.objectiveIds?.includes(objective.id),
                    )}
                    icon={<ObjectiveIcon className="h-4" />}
                    key={objective.id}
                    label={objective.name}
                    onClick={() => {
                      const objectiveIds = toggleValue(
                        filters.objectiveIds,
                        objective.id,
                      );
                      setFilters({
                        ...filters,
                        objectiveIds:
                          objectiveIds.length > 0 ? objectiveIds : undefined,
                      });
                    }}
                  />
                ))}
              </FilterSection>
            </>
          ) : null}
        </Box>
      </Popover.Content>
    </Popover>
  );
};

export const KeyResultsFilterBar = ({
  filters,
  setFilters,
}: {
  filters: KeyResultFilters;
  setFilters: (filters: KeyResultFilters) => void;
}) => {
  const { data: teams = [] } = useTeams();
  const { data: objectives = [] } = useObjectives();
  const { data: members = [] } = useMembers();

  if (getActiveKeyResultFilterCount(filters) === 0) return null;

  const labels = [
    ...(filters.teamIds ?? []).map(
      (id) => teams.find((team) => team.id === id)?.name ?? "Team",
    ),
    ...(filters.objectiveIds ?? []).map(
      (id) =>
        objectives.find((objective) => objective.id === id)?.name ??
        "Objective",
    ),
    ...(filters.leadIds ?? []).map((id) => {
      const member = members.find((item) => item.id === id);
      return member?.fullName || member?.username || "Lead";
    }),
    ...(filters.measurementTypes ?? []).map((type) =>
      type === "boolean"
        ? "Complete / incomplete"
        : `${type[0].toUpperCase()}${type.slice(1)}`,
    ),
  ];

  return (
    <Flex
      align="center"
      className="border-border bg-background h-[3.6rem] overflow-x-auto border-b-[0.5px] px-4"
      gap={2}
    >
      <FilterIcon className="text-text-muted h-[1.1rem] shrink-0" />
      {labels.map((label, index) => (
        <span
          className="border-border bg-surface-muted shrink-0 rounded-md border-[0.5px] px-2 py-1"
          key={`${label}-${index}`}
        >
          {label}
        </span>
      ))}
      {filters.endDateAfter || filters.endDateBefore ? (
        <span className="border-border bg-surface-muted flex shrink-0 items-center gap-1.5 rounded-md border-[0.5px] px-2 py-1">
          <CalendarIcon className="h-4" />
          Delivery date
        </span>
      ) : null}
      <Button
        className="ml-auto shrink-0"
        color="tertiary"
        onClick={() => {
          setFilters({});
        }}
        size="sm"
        variant="naked"
      >
        Clear all
      </Button>
    </Flex>
  );
};
