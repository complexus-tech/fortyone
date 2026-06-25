"use client";

import { parseAsArrayOf, parseAsString, useQueryStates } from "nuqs";
import { ArrowDownIcon, CheckIcon, PreferencesIcon } from "icons";
import { Badge, Box, Button, Divider, Flex, Popover, Text } from "ui";
import { useTerminology } from "@/hooks";
import { useObjectives } from "@/modules/objectives/hooks/use-objectives";
import { useSprints } from "@/modules/sprints/hooks/sprints";
import { useTeams } from "@/modules/teams/hooks/teams";
import { DateRangeFilter } from "./date-range-filter";

type FilterOption = {
  id: string;
  label: string;
  color?: string | null;
};

const toggleValue = (selected: string[], value: string) =>
  selected.includes(value)
    ? selected.filter((item) => item !== value)
    : [...selected, value];

const FilterSection = ({
  emptyText,
  label,
  options,
  selected,
  onChange,
}: {
  emptyText: string;
  label: string;
  options: FilterOption[];
  selected: string[];
  onChange: (values: string[] | null) => void;
}) => {
  return (
    <Box className="px-4 py-3">
      <Flex align="center" className="mb-2" justify="between">
        <Text fontWeight="medium">{label}</Text>
        {selected.length ? (
          <Button
            color="tertiary"
            onClick={() => {
              onChange(null);
            }}
            size="xs"
            variant="naked"
          >
            Clear
          </Button>
        ) : null}
      </Flex>

      {options.length ? (
        <Flex className="gap-2" wrap>
          {options.map((option) => {
            const isSelected = selected.includes(option.id);

            return (
              <Button
                className="max-w-full rounded-lg px-2"
                color={isSelected ? "primary" : "tertiary"}
                key={option.id}
                onClick={() => {
                  const next = toggleValue(selected, option.id);
                  onChange(next.length ? next : null);
                }}
                size="xs"
                variant={isSelected ? "solid" : "outline"}
              >
                <Flex align="center" className="min-w-0 gap-1.5">
                  {option.color ? (
                    <Box
                      className="size-2.5 shrink-0 rounded-sm"
                      style={{ backgroundColor: option.color }}
                    />
                  ) : null}
                  <span className="max-w-36 truncate">{option.label}</span>
                  {isSelected ? <CheckIcon className="size-3" /> : null}
                </Flex>
              </Button>
            );
          })}
        </Flex>
      ) : (
        <Text color="muted">{emptyText}</Text>
      )}
    </Box>
  );
};

const AnalyticsFilterMenu = () => {
  const { getTermDisplay } = useTerminology();
  const { data: teams = [] } = useTeams();
  const { data: objectives = [] } = useObjectives();
  const { data: sprints = [] } = useSprints();
  const [filters, setFilters] = useQueryStates({
    objectiveIds: parseAsArrayOf(parseAsString),
    sprintIds: parseAsArrayOf(parseAsString),
    teamIds: parseAsArrayOf(parseAsString),
  });

  const selectedTeams = filters.teamIds ?? [];
  const selectedObjectives = filters.objectiveIds ?? [];
  const selectedSprints = filters.sprintIds ?? [];
  const activeFilterCount =
    selectedTeams.length + selectedObjectives.length + selectedSprints.length;

  return (
    <Popover>
      <Popover.Trigger asChild>
        <Button
          className="gap-2"
          color="tertiary"
          leftIcon={<PreferencesIcon className="text-text-muted h-4 w-auto" />}
          rightIcon={<ArrowDownIcon className="text-text-muted h-3.5 w-auto" />}
          size="sm"
          variant="outline"
        >
          Filters
          {activeFilterCount ? (
            <Badge color="primary" rounded="full" size="sm">
              {activeFilterCount}
            </Badge>
          ) : null}
        </Button>
      </Popover.Trigger>
      <Popover.Content align="end" className="w-96 pb-2">
        <Flex align="center" className="px-4 pt-3 pb-2" justify="between">
          <Box>
            <Text fontWeight="semibold">Report filters</Text>
            <Text className="mt-1" color="muted">
              Narrow the analytics dataset.
            </Text>
          </Box>
          {activeFilterCount ? (
            <Button
              color="tertiary"
              onClick={() => {
                setFilters({
                  objectiveIds: null,
                  sprintIds: null,
                  teamIds: null,
                });
              }}
              size="sm"
              variant="naked"
            >
              Reset
            </Button>
          ) : null}
        </Flex>
        <Divider className="my-1" />
        <FilterSection
          emptyText="No teams available."
          label="Teams"
          onChange={(teamIds) => {
            setFilters({ teamIds });
          }}
          options={teams.map((team) => ({
            color: team.color,
            id: team.id,
            label: team.name,
          }))}
          selected={selectedTeams}
        />
        <Divider className="my-1" />
        <FilterSection
          emptyText={`No ${getTermDisplay("objectiveTerm", { variant: "plural" })} available.`}
          label={getTermDisplay("objectiveTerm", {
            capitalize: true,
            variant: "plural",
          })}
          onChange={(objectiveIds) => {
            setFilters({ objectiveIds });
          }}
          options={objectives.map((objective) => ({
            id: objective.id,
            label: objective.name,
          }))}
          selected={selectedObjectives}
        />
        <Divider className="my-1" />
        <FilterSection
          emptyText={`No ${getTermDisplay("sprintTerm", { variant: "plural" })} available.`}
          label={getTermDisplay("sprintTerm", {
            capitalize: true,
            variant: "plural",
          })}
          onChange={(sprintIds) => {
            setFilters({ sprintIds });
          }}
          options={sprints.map((sprint) => ({
            id: sprint.id,
            label: sprint.name,
          }))}
          selected={selectedSprints}
        />
      </Popover.Content>
    </Popover>
  );
};

export const Filters = () => {
  return (
    <Flex align="center" className="gap-2">
      <DateRangeFilter showLabel={false} />
      <AnalyticsFilterMenu />
    </Flex>
  );
};
