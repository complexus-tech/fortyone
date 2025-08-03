"use client";
import { useState } from "react";
import { parseAsArrayOf, parseAsString, useQueryStates } from "nuqs";
import { Box, Command, Divider, Flex, Text } from "ui";
import { TeamIcon, CheckIcon } from "icons";
import { useTeams } from "@/modules/teams/hooks/teams";
import { FilterButton } from "./filter-button";

const TeamsSelector = () => {
  const { data: teams = [] } = useTeams();
  const [filters, setFilters] = useQueryStates({
    teamIds: parseAsArrayOf(parseAsString),
  });
  const [query, setQuery] = useState("");

  const selectedTeams = filters.teamIds || [];

  const toggleTeam = (teamId: string) => {
    const newTeamIds = selectedTeams.includes(teamId)
      ? selectedTeams.filter((id) => id !== teamId)
      : [...selectedTeams, teamId];

    setFilters({ teamIds: newTeamIds.length > 0 ? newTeamIds : null });
  };

  const filteredTeams = teams.filter((team) =>
    team.name.toLowerCase().includes(query.toLowerCase()),
  );

  return (
    <Command>
      <Command.Input
        onValueChange={setQuery}
        placeholder="Search teams..."
        value={query}
      />
      <Divider className="my-2" />
      <Command.Group>
        <Command.Item
          className="justify-between"
          onSelect={() => {
            setFilters({ teamIds: null });
          }}
        >
          <Flex align="center" gap={2}>
            <TeamIcon className="h-4 w-auto" />
            <Text>All teams</Text>
          </Flex>
          {selectedTeams.length === 0 && <CheckIcon className="h-4 w-auto" />}
        </Command.Item>
        {filteredTeams.map((team) => (
          <Command.Item
            className="justify-between"
            key={team.id}
            onSelect={() => {
              toggleTeam(team.id);
            }}
          >
            <Flex align="center" gap={2}>
              <Box
                className="size-3 rounded"
                style={{ backgroundColor: team.color }}
              />
              <Text>{team.name}</Text>
            </Flex>
            {selectedTeams.includes(team.id) && (
              <CheckIcon className="h-4 w-auto" />
            )}
          </Command.Item>
        ))}
      </Command.Group>
    </Command>
  );
};

export const TeamsFilter = () => {
  const [filters] = useQueryStates({
    teamIds: parseAsArrayOf(parseAsString),
  });
  const { data: teams = [] } = useTeams();

  const selectedTeams = filters.teamIds || [];

  const getTeamsText = () => {
    if (selectedTeams.length === 0) return "All";
    if (selectedTeams.length === 1) {
      const team = teams.find((t) => t.id === selectedTeams[0]);
      return team?.name || "Unknown";
    }
    return `${selectedTeams.length} teams`;
  };

  return (
    <FilterButton
      icon={<TeamIcon className="h-4 w-auto" />}
      isActive={selectedTeams.length > 0}
      label="Teams"
      popover={<TeamsSelector />}
      text={getTeamsText()}
    />
  );
};
