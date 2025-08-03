"use client";
import { useState } from "react";
import { parseAsArrayOf, parseAsString, useQueryStates } from "nuqs";
import { Command, Divider, Flex, Text } from "ui";
import { ObjectiveIcon, CheckIcon } from "icons";
import { useObjectives } from "@/modules/objectives/hooks/use-objectives";
import { FilterButton } from "./filter-button";

const ObjectivesSelector = () => {
  const { data: objectives = [] } = useObjectives();
  const [filters, setFilters] = useQueryStates({
    objectiveIds: parseAsArrayOf(parseAsString),
  });
  const [query, setQuery] = useState("");

  const selectedObjectives = filters.objectiveIds || [];

  const toggleObjective = (objectiveId: string) => {
    const newObjectiveIds = selectedObjectives.includes(objectiveId)
      ? selectedObjectives.filter((id) => id !== objectiveId)
      : [...selectedObjectives, objectiveId];

    setFilters({
      objectiveIds: newObjectiveIds.length > 0 ? newObjectiveIds : null,
    });
  };

  const filteredObjectives = objectives.filter((objective) =>
    objective.name.toLowerCase().includes(query.toLowerCase()),
  );

  return (
    <Command>
      <Command.Input
        onValueChange={setQuery}
        placeholder="Search objectives..."
        value={query}
      />
      <Divider className="my-2" />
      <Command.Group>
        <Command.Item
          className="justify-between"
          onSelect={() => {
            setFilters({ objectiveIds: null });
          }}
        >
          <Flex align="center" gap={2}>
            <ObjectiveIcon className="h-4 w-auto" />
            <Text>All objectives</Text>
          </Flex>
          {selectedObjectives.length === 0 && (
            <CheckIcon className="h-4 w-auto" />
          )}
        </Command.Item>
        {filteredObjectives.map((objective) => (
          <Command.Item
            className="justify-between"
            key={objective.id}
            onSelect={() => {
              toggleObjective(objective.id);
            }}
          >
            <Flex align="center" gap={2}>
              <ObjectiveIcon className="h-4 w-auto" />
              <Text>{objective.name}</Text>
            </Flex>
            {selectedObjectives.includes(objective.id) && (
              <CheckIcon className="h-4 w-auto" />
            )}
          </Command.Item>
        ))}
      </Command.Group>
    </Command>
  );
};

export const ObjectivesFilter = () => {
  const [filters] = useQueryStates({
    objectiveIds: parseAsArrayOf(parseAsString),
  });
  const { data: objectives = [] } = useObjectives();

  const selectedObjectives = filters.objectiveIds || [];

  const getObjectivesText = () => {
    if (selectedObjectives.length === 0) return "All";
    if (selectedObjectives.length === 1) {
      const objective = objectives.find((o) => o.id === selectedObjectives[0]);
      return objective?.name || "Unknown";
    }
    return `${selectedObjectives.length} objectives`;
  };

  return (
    <FilterButton
      icon={<ObjectiveIcon className="h-4 w-auto" />}
      isActive={selectedObjectives.length > 0}
      label="Objectives"
      popover={<ObjectivesSelector />}
      text={getObjectivesText()}
    />
  );
};
