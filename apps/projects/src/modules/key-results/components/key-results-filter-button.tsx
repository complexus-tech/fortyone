"use client";

import type { ReactNode } from "react";
import { format, formatISO } from "date-fns";
import { cn } from "lib";
import {
  ArrowDownIcon,
  ArrowRightIcon,
  AssigneeIcon,
  CalendarIcon,
  CheckIcon,
  CloseIcon,
  FilterIcon,
  ObjectiveIcon,
  OKRIcon,
  PlusIcon,
  TeamIcon,
} from "icons";
import {
  Avatar,
  Box,
  Button,
  Divider,
  Flex,
  Input,
  Menu,
  Popover,
  Text,
} from "ui";
import { TeamColor } from "@/components/ui";
import { MemberTooltip } from "@/components/ui/member-tooltip";
import { useTerminology } from "@/hooks";
import { useMembers } from "@/lib/hooks/members";
import { useObjectives } from "@/modules/objectives/hooks/use-objectives";
import { useTeams } from "@/modules/teams/hooks/teams";
import type { Member } from "@/types";
import { hexToRgba } from "@/utils";
import { getActiveKeyResultFilterCount } from "../key-results-filter-utils";
import type { KeyResultFilters } from "../types";

type FilterField =
  | "leadIds"
  | "endDate"
  | "measurementTypes"
  | "teamIds"
  | "objectiveIds";

type FilterOption = {
  field: FilterField;
  icon: ReactNode;
  label: string;
};

type MeasurementType = NonNullable<
  KeyResultFilters["measurementTypes"]
>[number];

const MEASUREMENT_OPTIONS = [
  { label: "Percentage", value: "percentage" },
  { label: "Number", value: "number" },
  { label: "Complete / incomplete", value: "boolean" },
] as const satisfies readonly {
  label: string;
  value: MeasurementType;
}[];

const toggleValue = (values: string[] | undefined, value: string) => {
  const selected = values ?? [];
  return selected.includes(value)
    ? selected.filter((item) => item !== value)
    : [...selected, value];
};

const normalizeValues = <T,>(values: T[]) =>
  values.length > 0 ? values : undefined;

const getMemberName = (member: Member) =>
  member.fullName.trim() || member.username || member.email || "Unknown user";

const FilterSection = ({
  children,
  title,
}: {
  children: ReactNode;
  title: string;
}) => (
  <Box className="px-4 py-3">
    <Text className="mb-2.5" fontWeight="medium">
      {title}
    </Text>
    {children}
  </Box>
);

const LeadSelector = ({
  members,
  onChange,
  selected,
}: {
  members: Member[];
  onChange: (leadIds: string[]) => void;
  selected: string[] | undefined;
}) => {
  const selectedSet = new Set(selected);

  return (
    <Flex gap={2} wrap>
      {members.map((member) => {
        const active = selectedSet.has(member.id);
        const memberName = getMemberName(member);

        return (
          <MemberTooltip key={member.id} member={member}>
            <button
              aria-pressed={active}
              className={cn(
                "relative rounded-full ring-2 ring-transparent transition",
                {
                  "ring-primary": active,
                },
              )}
              onClick={() => {
                onChange(toggleValue(selected, member.id));
              }}
              type="button"
            >
              <Avatar
                className="h-10"
                name={memberName}
                src={member.avatarUrl}
              />
              <span className="sr-only">{memberName}</span>
              {active ? (
                <span className="bg-primary text-on-primary absolute -right-0.5 -bottom-0.5 flex size-4 items-center justify-center rounded-full">
                  <CheckIcon className="h-3 w-auto" strokeWidth={3} />
                </span>
              ) : null}
            </button>
          </MemberTooltip>
        );
      })}
      {members.length === 0 ? <Text color="muted">No members</Text> : null}
    </Flex>
  );
};

const TeamSelector = ({
  onChange,
  selected,
  teams,
}: {
  onChange: (teamIds: string[]) => void;
  selected: string[] | undefined;
  teams: ReturnType<typeof useTeams>["data"];
}) => {
  const selectedSet = new Set(selected);

  return (
    <Flex gap={2} wrap>
      {(teams ?? []).map((team) => {
        const active = selectedSet.has(team.id);

        return (
          <Button
            className={cn("ring-2 ring-transparent", {
              "ring-primary": active,
            })}
            color="tertiary"
            key={team.id}
            leftIcon={<TeamColor color={team.color} />}
            onClick={() => {
              onChange(toggleValue(selected, team.id));
            }}
            size="sm"
            style={{
              backgroundColor: hexToRgba(team.color, 0.1),
              borderColor: hexToRgba(team.color, 0.2),
            }}
            variant={active ? "solid" : "outline"}
          >
            {team.name}
          </Button>
        );
      })}
    </Flex>
  );
};

const ObjectiveSelector = ({
  objectiveLabel,
  objectives,
  onChange,
  selected,
}: {
  objectiveLabel: string;
  objectives: ReturnType<typeof useObjectives>["data"];
  onChange: (objectiveIds: string[]) => void;
  selected: string[] | undefined;
}) => {
  const selectedSet = new Set(selected);

  return (
    <Flex gap={2} wrap>
      {(objectives ?? []).map((objective) => {
        const active = selectedSet.has(objective.id);

        return (
          <Button
            className={cn("max-w-full ring-2 ring-transparent", {
              "ring-primary": active,
            })}
            color="tertiary"
            key={objective.id}
            leftIcon={<ObjectiveIcon className="h-4 w-auto" />}
            onClick={() => {
              onChange(toggleValue(selected, objective.id));
            }}
            size="sm"
            title={objective.name}
            variant={active ? "solid" : "outline"}
          >
            <span className="truncate">{objective.name}</span>
          </Button>
        );
      })}
      {(objectives ?? []).length === 0 ? (
        <Text color="muted">No {objectiveLabel}</Text>
      ) : null}
    </Flex>
  );
};

const MeasurementSelector = ({
  onChange,
  selected,
}: {
  onChange: (measurementTypes: MeasurementType[]) => void;
  selected: MeasurementType[] | undefined;
}) => {
  const selectedSet = new Set(selected);

  return (
    <Flex gap={2} wrap>
      {MEASUREMENT_OPTIONS.map((option) => {
        const active = selectedSet.has(option.value);

        return (
          <Button
            className={cn("ring-2 ring-transparent", {
              "ring-primary": active,
            })}
            color="tertiary"
            key={option.value}
            leftIcon={<OKRIcon className="h-4 w-auto" />}
            onClick={() => {
              onChange(
                toggleValue(selected, option.value) as MeasurementType[],
              );
            }}
            size="sm"
            variant={active ? "solid" : "outline"}
          >
            {option.label}
          </Button>
        );
      })}
    </Flex>
  );
};

const DeliveryDateSelector = ({
  filters,
  setFilters,
}: {
  filters: KeyResultFilters;
  setFilters: (filters: KeyResultFilters) => void;
}) => (
  <Box className="grid grid-cols-2 gap-2">
    <Box>
      <Text className="mb-1.5" color="muted" fontSize="sm">
        From
      </Text>
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
    </Box>
    <Box>
      <Text className="mb-1.5" color="muted" fontSize="sm">
        To
      </Text>
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
  </Box>
);

const FilterValueEditor = ({
  field,
  filters,
  members,
  objectiveLabel,
  objectives,
  setFilters,
  teams,
}: {
  field: FilterField;
  filters: KeyResultFilters;
  members: Member[];
  objectiveLabel: string;
  objectives: ReturnType<typeof useObjectives>["data"];
  setFilters: (filters: KeyResultFilters) => void;
  teams: ReturnType<typeof useTeams>["data"];
}) => {
  if (field === "leadIds") {
    return (
      <LeadSelector
        members={members}
        onChange={(leadIds) => {
          setFilters({ ...filters, leadIds: normalizeValues(leadIds) });
        }}
        selected={filters.leadIds}
      />
    );
  }

  if (field === "teamIds") {
    return (
      <TeamSelector
        onChange={(teamIds) => {
          setFilters({ ...filters, teamIds: normalizeValues(teamIds) });
        }}
        selected={filters.teamIds}
        teams={teams}
      />
    );
  }

  if (field === "objectiveIds") {
    return (
      <ObjectiveSelector
        objectiveLabel={objectiveLabel}
        objectives={objectives}
        onChange={(objectiveIds) => {
          setFilters({
            ...filters,
            objectiveIds: normalizeValues(objectiveIds),
          });
        }}
        selected={filters.objectiveIds}
      />
    );
  }

  if (field === "measurementTypes") {
    return (
      <MeasurementSelector
        onChange={(measurementTypes) => {
          setFilters({
            ...filters,
            measurementTypes: normalizeValues(measurementTypes),
          });
        }}
        selected={filters.measurementTypes}
      />
    );
  }

  return <DeliveryDateSelector filters={filters} setFilters={setFilters} />;
};

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
  const { getTermDisplay } = useTerminology();
  const activeFilterCount = getActiveKeyResultFilterCount(filters);
  const keyResultLabel = getTermDisplay("keyResultTerm", {
    variant: "plural",
  });
  const objectiveLabel = getTermDisplay("objectiveTerm", {
    variant: "plural",
  });
  const buttonLabel =
    activeFilterCount > 0
      ? `${activeFilterCount} filter${activeFilterCount === 1 ? "" : "s"} applied`
      : "Filters";

  return (
    <Popover>
      <Popover.Trigger asChild>
        <Button
          aria-label={buttonLabel}
          className="relative"
          color="tertiary"
          leftIcon={<FilterIcon className="h-4 w-auto" />}
          rightIcon={<ArrowDownIcon className="h-3.5 w-auto" />}
          size="sm"
          variant="outline"
        >
          {activeFilterCount > 0 ? (
            <span
              aria-hidden="true"
              className="bg-primary absolute -top-0.5 -right-0.5 h-2.5 w-2.5 rounded-full"
            >
              <span className="bg-primary absolute inset-0 animate-ping rounded-full opacity-75" />
            </span>
          ) : null}
          <span className="hidden md:inline">{buttonLabel}</span>
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
          {activeFilterCount > 0 ? (
            <Button
              className="text-primary dark:text-primary"
              color="tertiary"
              onClick={() => {
                setFilters({});
              }}
              size="sm"
              variant="naked"
            >
              Clear filters
            </Button>
          ) : null}
        </Flex>
        <Divider className="mt-1.5" />
        <FilterSection title="Lead">
          <LeadSelector
            members={members}
            onChange={(leadIds) => {
              setFilters({ ...filters, leadIds: normalizeValues(leadIds) });
            }}
            selected={filters.leadIds}
          />
        </FilterSection>
        <Divider />
        <FilterSection title="Delivery date">
          <DeliveryDateSelector filters={filters} setFilters={setFilters} />
        </FilterSection>
        <Divider />
        <FilterSection title="Measurement type">
          <MeasurementSelector
            onChange={(measurementTypes) => {
              setFilters({
                ...filters,
                measurementTypes: normalizeValues(measurementTypes),
              });
            }}
            selected={filters.measurementTypes}
          />
        </FilterSection>
        {teams.length > 0 ? (
          <>
            <Divider />
            <FilterSection title="Team">
              <TeamSelector
                onChange={(teamIds) => {
                  setFilters({
                    ...filters,
                    teamIds: normalizeValues(teamIds),
                  });
                }}
                selected={filters.teamIds}
                teams={teams}
              />
            </FilterSection>
          </>
        ) : null}
        {objectives.length > 0 ? (
          <>
            <Divider />
            <FilterSection
              title={getTermDisplay("objectiveTerm", {
                variant: "plural",
                capitalize: true,
              })}
            >
              <ObjectiveSelector
                objectiveLabel={objectiveLabel}
                objectives={objectives}
                onChange={(objectiveIds) => {
                  setFilters({
                    ...filters,
                    objectiveIds: normalizeValues(objectiveIds),
                  });
                }}
                selected={filters.objectiveIds}
              />
            </FilterSection>
          </>
        ) : null}
        <Text className="sr-only">Filter {keyResultLabel}</Text>
      </Popover.Content>
    </Popover>
  );
};

type LeadChipMember = {
  avatarUrl?: string | null;
  id: string;
  name: string;
  username: string;
};

const LeadChipValue = ({ members }: { members: LeadChipMember[] }) => {
  const visibleMembers = members.slice(0, 2);

  if (members.length > 2) {
    return (
      <Flex align="center" gap={1}>
        <Flex align="center" className="-space-x-1">
          {visibleMembers.map((member) => (
            <Avatar
              className="ring-background ring-1"
              color="primary"
              key={member.id}
              name={member.name}
              size="xs"
              src={member.avatarUrl}
            />
          ))}
        </Flex>
        <span>{members.length} leads</span>
      </Flex>
    );
  }

  return (
    <Flex align="center" gap={2}>
      {visibleMembers.map((member) => (
        <Flex align="center" gap={1} key={member.id}>
          <Avatar
            color="primary"
            name={member.name}
            size="xs"
            src={member.avatarUrl}
          />
          <span>{member.username}</span>
        </Flex>
      ))}
    </Flex>
  );
};

const FilterChip = ({
  editor,
  icon,
  label,
  onRemove,
  operator,
  value,
}: {
  editor: ReactNode;
  icon: ReactNode;
  label: string;
  onRemove: () => void;
  operator: string;
  value: ReactNode;
}) => (
  <Flex
    align="center"
    className="border-border bg-surface h-[2.1rem] shrink-0 overflow-hidden rounded-xl border"
    gap={0}
  >
    <span className="border-border text-text-secondary flex h-full items-center gap-1.5 border-r px-2.5">
      {icon}
      {label}
    </span>
    <span className="border-border text-text-secondary flex h-full items-center border-r px-2.5">
      {operator}
    </span>
    <Popover>
      <Popover.Trigger asChild>
        <button
          className="hover:bg-state-hover flex h-full max-w-72 items-center truncate px-2.5 text-left transition"
          type="button"
        >
          <span className="flex min-w-0 items-center truncate">{value}</span>
        </button>
      </Popover.Trigger>
      <Popover.Content align="start" className="w-80 p-4">
        {editor}
      </Popover.Content>
    </Popover>
    <button
      aria-label={`Remove ${label} filter`}
      className="hover:bg-state-hover border-border flex h-full w-9 items-center justify-center border-l transition"
      onClick={onRemove}
      type="button"
    >
      <CloseIcon className="text-text-secondary h-3.5 w-auto" />
    </button>
  </Flex>
);

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
  const { getTermDisplay } = useTerminology();

  if (getActiveKeyResultFilterCount(filters) === 0) return null;

  const objectiveSingular = getTermDisplay("objectiveTerm", {
    capitalize: true,
  });
  const objectivePlural = getTermDisplay("objectiveTerm", {
    variant: "plural",
  });
  const memberById = new Map(members.map((member) => [member.id, member]));
  const teamById = new Map(teams.map((team) => [team.id, team]));
  const objectiveById = new Map(
    objectives.map((objective) => [objective.id, objective]),
  );
  const filterOptions: FilterOption[] = [
    {
      field: "leadIds",
      icon: <AssigneeIcon className="h-5 w-auto" />,
      label: "Lead",
    },
    {
      field: "endDate",
      icon: <CalendarIcon className="h-5 w-auto" />,
      label: "Delivery date",
    },
    {
      field: "measurementTypes",
      icon: <OKRIcon className="h-5 w-auto" />,
      label: "Measurement type",
    },
    {
      field: "teamIds",
      icon: <TeamIcon className="h-5 w-auto" />,
      label: "Team",
    },
    {
      field: "objectiveIds",
      icon: <ObjectiveIcon className="h-5 w-auto" />,
      label: objectiveSingular,
    },
  ];
  const editorFor = (field: FilterField) => (
    <FilterValueEditor
      field={field}
      filters={filters}
      members={members}
      objectiveLabel={objectivePlural}
      objectives={objectives}
      setFilters={setFilters}
      teams={teams}
    />
  );
  const chips: ReactNode[] = [];

  if (filters.leadIds?.length) {
    const selectedMembers = filters.leadIds.map((id) => {
      const member = memberById.get(id);
      const fallbackName = member ? getMemberName(member) : "Unknown user";

      return {
        avatarUrl: member?.avatarUrl,
        id,
        name: fallbackName,
        username: member?.username || fallbackName,
      };
    });
    chips.push(
      <FilterChip
        editor={editorFor("leadIds")}
        icon={<AssigneeIcon className="h-4 w-auto" />}
        key="leadIds"
        label="Lead"
        onRemove={() => {
          setFilters({ ...filters, leadIds: undefined });
        }}
        operator="is any of"
        value={<LeadChipValue members={selectedMembers} />}
      />,
    );
  }

  if (filters.endDateAfter || filters.endDateBefore) {
    const from = filters.endDateAfter
      ? format(new Date(filters.endDateAfter), "MMM d, yyyy")
      : "Any time";
    const to = filters.endDateBefore
      ? format(new Date(filters.endDateBefore), "MMM d, yyyy")
      : "Any time";
    chips.push(
      <FilterChip
        editor={editorFor("endDate")}
        icon={<CalendarIcon className="h-4 w-auto" />}
        key="endDate"
        label="Delivery date"
        onRemove={() => {
          setFilters({
            ...filters,
            endDateAfter: undefined,
            endDateBefore: undefined,
          });
        }}
        operator="is between"
        value={`${from} – ${to}`}
      />,
    );
  }

  if (filters.measurementTypes?.length) {
    const value = filters.measurementTypes
      .map(
        (type) =>
          MEASUREMENT_OPTIONS.find((option) => option.value === type)?.label ??
          type,
      )
      .join(", ");
    chips.push(
      <FilterChip
        editor={editorFor("measurementTypes")}
        icon={<OKRIcon className="h-4 w-auto" />}
        key="measurementTypes"
        label="Measurement type"
        onRemove={() => {
          setFilters({ ...filters, measurementTypes: undefined });
        }}
        operator="is any of"
        value={value}
      />,
    );
  }

  if (filters.teamIds?.length) {
    const selectedTeams = filters.teamIds.map((id) => teamById.get(id));
    chips.push(
      <FilterChip
        editor={editorFor("teamIds")}
        icon={<TeamColor color={selectedTeams[0]?.color} />}
        key="teamIds"
        label="Team"
        onRemove={() => {
          setFilters({ ...filters, teamIds: undefined });
        }}
        operator="is any of"
        value={selectedTeams
          .map((team, index) => team?.name ?? filters.teamIds?.[index] ?? "")
          .join(", ")}
      />,
    );
  }

  if (filters.objectiveIds?.length) {
    chips.push(
      <FilterChip
        editor={editorFor("objectiveIds")}
        icon={<ObjectiveIcon className="h-4 w-auto" />}
        key="objectiveIds"
        label={objectiveSingular}
        onRemove={() => {
          setFilters({ ...filters, objectiveIds: undefined });
        }}
        operator="is any of"
        value={filters.objectiveIds
          .map((id) => objectiveById.get(id)?.name ?? id)
          .join(", ")}
      />,
    );
  }

  return (
    <Flex
      align="center"
      className="border-border bg-background h-[3.6rem] border-b-[0.5px] px-4"
      gap={3}
      justify="between"
    >
      <Flex align="center" className="min-w-0 flex-1 overflow-x-auto" gap={2}>
        {chips}
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
            <Menu.Group className="max-h-[min(30rem,calc(100dvh-12rem))] overflow-y-auto px-1 py-1.5">
              {filterOptions.map((option) => (
                <Menu.SubMenu key={option.field}>
                  <Menu.SubTrigger
                    active={
                      option.field === "endDate"
                        ? Boolean(filters.endDateAfter || filters.endDateBefore)
                        : Boolean(filters[option.field]?.length)
                    }
                    className="justify-between gap-4"
                  >
                    <Box className="grid min-w-0 flex-1 grid-cols-[24px_minmax(0,1fr)] items-center">
                      <span className="text-text-secondary flex h-6 w-6 shrink-0 items-center">
                        {option.icon}
                      </span>
                      <Text className="truncate">{option.label}</Text>
                    </Box>
                    <ArrowRightIcon
                      className="text-text-muted h-3.5 w-auto"
                      strokeWidth={2.8}
                    />
                  </Menu.SubTrigger>
                  <Menu.SubItems
                    alignOffset={-6}
                    className="w-80 p-4"
                    sideOffset={8}
                  >
                    {editorFor(option.field)}
                  </Menu.SubItems>
                </Menu.SubMenu>
              ))}
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>
      <Button
        className="shrink-0"
        color="tertiary"
        onClick={() => {
          setFilters({});
        }}
        size="sm"
        variant="outline"
      >
        Clear all
      </Button>
    </Flex>
  );
};
