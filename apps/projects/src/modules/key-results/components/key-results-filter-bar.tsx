"use client";

import type { ReactNode } from "react";
import { format } from "date-fns";
import {
  AssigneeIcon,
  CalendarIcon,
  ChevronRightIcon,
  ObjectiveIcon,
  OKRIcon,
  PlusIcon,
  TeamIcon,
} from "icons";
import { Box, Button, Flex, Menu, Text } from "ui";
import { TeamColor } from "@/components/ui/team-color";
import { useTerminology } from "@/hooks/use-terminology-display";
import { getActiveKeyResultFilterCount } from "../key-results-filter-utils";
import type { KeyResultFilters } from "../types";
import {
  KeyResultsFilterChip,
  KeyResultsLeadChipValue,
} from "./key-results-filter-chips";
import {
  KeyResultsFilterValueEditor,
  useKeyResultsFilterData,
} from "./key-results-filter-controls";
import {
  getKeyResultsMemberName,
  KEY_RESULT_MEASUREMENT_OPTIONS,
} from "./key-results-filter-values";
import type { KeyResultsFilterField } from "./key-results-filter-values";

type KeyResultsFilterOption = {
  field: KeyResultsFilterField;
  icon: ReactNode;
  label: string;
};

export const KeyResultsFilterBar = ({
  filters,
  setFilters,
}: {
  filters: KeyResultFilters;
  setFilters: (filters: KeyResultFilters) => void;
}) => {
  const { members, objectives, teams } = useKeyResultsFilterData();
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
  const filterOptions: KeyResultsFilterOption[] = [
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
  const filterEditorProps = {
    filters,
    members,
    objectiveLabel: objectivePlural,
    objectives,
    setFilters,
    teams,
  };
  const editorFor = (field: KeyResultsFilterField) => (
    <KeyResultsFilterValueEditor {...filterEditorProps} field={field} />
  );
  const chips: ReactNode[] = [];

  if (filters.leadIds?.length) {
    const selectedMembers = filters.leadIds.map((id) => {
      const member = memberById.get(id);
      const fallbackName = member
        ? getKeyResultsMemberName(member)
        : "Unknown user";

      return {
        avatarUrl: member?.avatarUrl,
        id,
        name: fallbackName,
        username: member?.username || fallbackName,
      };
    });
    chips.push(
      <KeyResultsFilterChip
        editor={editorFor("leadIds")}
        icon={<AssigneeIcon className="h-4 w-auto" />}
        key="leadIds"
        label="Lead"
        onRemove={() => {
          setFilters({ ...filters, leadIds: undefined });
        }}
        operator="is any of"
        value={<KeyResultsLeadChipValue members={selectedMembers} />}
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
      <KeyResultsFilterChip
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
          KEY_RESULT_MEASUREMENT_OPTIONS.find((option) => option.value === type)
            ?.label ?? type,
      )
      .join(", ");
    chips.push(
      <KeyResultsFilterChip
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
      <KeyResultsFilterChip
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
      <KeyResultsFilterChip
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
                    <ChevronRightIcon
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
