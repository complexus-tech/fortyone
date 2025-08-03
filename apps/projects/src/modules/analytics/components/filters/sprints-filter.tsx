"use client";
import { useState } from "react";
import { parseAsArrayOf, parseAsString, useQueryStates } from "nuqs";
import { Command, Divider, Flex, Text } from "ui";
import { SprintsIcon, CheckIcon } from "icons";
import { format } from "date-fns";
import { useTeamSprints } from "@/modules/sprints/hooks/team-sprints";
import { FilterButton } from "./filter-button";

const SprintsSelector = () => {
  const { data: sprints = [] } = useTeamSprints("");
  const [filters, setFilters] = useQueryStates({
    sprintIds: parseAsArrayOf(parseAsString),
  });
  const [query, setQuery] = useState("");

  const selectedSprints = filters.sprintIds || [];

  const toggleSprint = (sprintId: string) => {
    const newSprintIds = selectedSprints.includes(sprintId)
      ? selectedSprints.filter((id) => id !== sprintId)
      : [...selectedSprints, sprintId];

    setFilters({ sprintIds: newSprintIds.length > 0 ? newSprintIds : null });
  };

  const filteredSprints = sprints.filter((sprint) =>
    sprint.name.toLowerCase().includes(query.toLowerCase()),
  );

  return (
    <Command>
      <Command.Input
        onValueChange={setQuery}
        placeholder="Search sprints..."
        value={query}
      />
      <Divider className="my-2" />
      <Command.Group>
        <Command.Item
          className="justify-between"
          onSelect={() => {
            setFilters({ sprintIds: null });
          }}
        >
          <Flex align="center" gap={2}>
            <SprintsIcon className="h-4 w-auto" />
            <Text>All sprints</Text>
          </Flex>
          {selectedSprints.length === 0 && <CheckIcon className="h-4 w-auto" />}
        </Command.Item>
        {filteredSprints.map((sprint) => (
          <Command.Item
            className="justify-between"
            key={sprint.id}
            onSelect={() => {
              toggleSprint(sprint.id);
            }}
          >
            <Flex align="center" gap={2}>
              <SprintsIcon className="h-4 w-auto" />
              <Text>
                {sprint.name}
                <Text as="span" color="muted">
                  {" "}
                  {format(new Date(sprint.startDate), "MMM d")} -{" "}
                  {format(new Date(sprint.endDate), "MMM d")}
                </Text>
              </Text>
            </Flex>
            {selectedSprints.includes(sprint.id) && (
              <CheckIcon className="h-4 w-auto" />
            )}
          </Command.Item>
        ))}
      </Command.Group>
    </Command>
  );
};

export const SprintsFilter = () => {
  const [filters] = useQueryStates({
    sprintIds: parseAsArrayOf(parseAsString),
  });
  const { data: sprints = [] } = useTeamSprints("");

  const selectedSprints = filters.sprintIds || [];

  const getSprintsText = () => {
    if (selectedSprints.length === 0) return "All";
    if (selectedSprints.length === 1) {
      const sprint = sprints.find((s) => s.id === selectedSprints[0]);
      return sprint?.name || "Unknown";
    }
    return `${selectedSprints.length} sprints`;
  };

  return (
    <FilterButton
      icon={<SprintsIcon className="h-4 w-auto" />}
      isActive={selectedSprints.length > 0}
      label="Sprints"
      popover={<SprintsSelector />}
      text={getSprintsText()}
    />
  );
};
