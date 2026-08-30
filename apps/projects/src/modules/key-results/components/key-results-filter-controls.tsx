"use client";

import type { ReactNode } from "react";
import { format, formatISO } from "date-fns";
import { cn } from "lib";
import { CheckIcon, ObjectiveIcon, OKRIcon } from "icons";
import { Avatar, Box, Button, Flex, Input, Text } from "ui";
import { TeamColor } from "@/components/ui/team-color";
import { MemberTooltip } from "@/components/ui/member-tooltip";
import { useMembers } from "@/lib/hooks/members";
import { useObjectives } from "@/modules/objectives/public/client";
import { useTeams } from "@/modules/teams/public/client";
import { hexToRgba } from "@/utils/color";
import type { KeyResultFilters } from "../types";
import {
  getKeyResultsMemberName,
  KEY_RESULT_MEASUREMENT_OPTIONS,
  normalizeKeyResultsFilterValues,
  toggleKeyResultsFilterValue,
} from "./key-results-filter-values";
import type {
  KeyResultsFilterField,
  KeyResultsMeasurementType,
} from "./key-results-filter-values";
import type { KeyResultsMember } from "./key-results-member";

type KeyResultsFilterTeams = NonNullable<ReturnType<typeof useTeams>["data"]>;
type KeyResultsFilterObjectives = NonNullable<
  ReturnType<typeof useObjectives>["data"]
>;

export type KeyResultsFilterData = {
  members: KeyResultsMember[];
  objectives: KeyResultsFilterObjectives;
  teams: KeyResultsFilterTeams;
};

type SetKeyResultsFilters = (filters: KeyResultFilters) => void;

export const useKeyResultsFilterData = (): KeyResultsFilterData => {
  const { data: teams = [] } = useTeams();
  const { data: objectives = [] } = useObjectives();
  const { data: members = [] } = useMembers();

  return { members, objectives, teams };
};

export const KeyResultsFilterSection = ({
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

const KeyResultsLeadSelector = ({
  members,
  onChange,
  selected,
}: {
  members: KeyResultsMember[];
  onChange: (leadIds: string[]) => void;
  selected: string[] | undefined;
}) => {
  const selectedSet = new Set(selected);

  return (
    <Flex gap={2} wrap>
      {members.map((member) => {
        const active = selectedSet.has(member.id);
        const memberName = getKeyResultsMemberName(member);

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
                onChange(toggleKeyResultsFilterValue(selected, member.id));
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

const KeyResultsTeamSelector = ({
  onChange,
  selected,
  teams,
}: {
  onChange: (teamIds: string[]) => void;
  selected: string[] | undefined;
  teams: KeyResultsFilterTeams;
}) => {
  const selectedSet = new Set(selected);

  return (
    <Flex gap={2} wrap>
      {teams.map((team) => {
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
              onChange(toggleKeyResultsFilterValue(selected, team.id));
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

const KeyResultsObjectiveSelector = ({
  objectiveLabel,
  objectives,
  onChange,
  selected,
}: {
  objectiveLabel: string;
  objectives: KeyResultsFilterObjectives;
  onChange: (objectiveIds: string[]) => void;
  selected: string[] | undefined;
}) => {
  const selectedSet = new Set(selected);

  return (
    <Flex gap={2} wrap>
      {objectives.map((objective) => {
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
              onChange(toggleKeyResultsFilterValue(selected, objective.id));
            }}
            size="sm"
            title={objective.name}
            variant={active ? "solid" : "outline"}
          >
            <span className="truncate">{objective.name}</span>
          </Button>
        );
      })}
      {objectives.length === 0 ? (
        <Text color="muted">No {objectiveLabel}</Text>
      ) : null}
    </Flex>
  );
};

const KeyResultsMeasurementSelector = ({
  onChange,
  selected,
}: {
  onChange: (measurementTypes: KeyResultsMeasurementType[]) => void;
  selected: KeyResultsMeasurementType[] | undefined;
}) => {
  const selectedSet = new Set(selected);

  return (
    <Flex gap={2} wrap>
      {KEY_RESULT_MEASUREMENT_OPTIONS.map((option) => {
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
                toggleKeyResultsFilterValue(
                  selected,
                  option.value,
                ) as KeyResultsMeasurementType[],
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

const KeyResultsDeliveryDateSelector = ({
  filters,
  setFilters,
}: {
  filters: KeyResultFilters;
  setFilters: SetKeyResultsFilters;
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

export const KeyResultsFilterValueEditor = ({
  field,
  filters,
  members,
  objectiveLabel,
  objectives,
  setFilters,
  teams,
}: {
  field: KeyResultsFilterField;
  filters: KeyResultFilters;
  members: KeyResultsMember[];
  objectiveLabel: string;
  objectives: KeyResultsFilterObjectives;
  setFilters: SetKeyResultsFilters;
  teams: KeyResultsFilterTeams;
}) => {
  if (field === "leadIds") {
    return (
      <KeyResultsLeadSelector
        members={members}
        onChange={(leadIds) => {
          setFilters({
            ...filters,
            leadIds: normalizeKeyResultsFilterValues(leadIds),
          });
        }}
        selected={filters.leadIds}
      />
    );
  }

  if (field === "teamIds") {
    return (
      <KeyResultsTeamSelector
        onChange={(teamIds) => {
          setFilters({
            ...filters,
            teamIds: normalizeKeyResultsFilterValues(teamIds),
          });
        }}
        selected={filters.teamIds}
        teams={teams}
      />
    );
  }

  if (field === "objectiveIds") {
    return (
      <KeyResultsObjectiveSelector
        objectiveLabel={objectiveLabel}
        objectives={objectives}
        onChange={(objectiveIds) => {
          setFilters({
            ...filters,
            objectiveIds: normalizeKeyResultsFilterValues(objectiveIds),
          });
        }}
        selected={filters.objectiveIds}
      />
    );
  }

  if (field === "measurementTypes") {
    return (
      <KeyResultsMeasurementSelector
        onChange={(measurementTypes) => {
          setFilters({
            ...filters,
            measurementTypes: normalizeKeyResultsFilterValues(measurementTypes),
          });
        }}
        selected={filters.measurementTypes}
      />
    );
  }

  return (
    <KeyResultsDeliveryDateSelector filters={filters} setFilters={setFilters} />
  );
};
