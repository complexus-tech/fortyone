"use client";

import { useState } from "react";
import { Box, Text, Input, Flex, Button } from "ui";
import { PlusIcon, SearchIcon, TeamIcon } from "icons";
import { useTeams } from "@/modules/teams/hooks/teams";
import { SectionHeader } from "@/modules/settings/components";
import { WorkspaceTeam } from "../components/team";

export const TeamsList = () => {
  const { data: teams = [] } = useTeams();
  const [search, setSearch] = useState("");

  const filteredTeams = teams.filter(
    (team) =>
      team.name.toLowerCase().includes(search.toLowerCase()) ||
      team.code.toLowerCase().includes(search.toLowerCase()),
  );

  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-medium">
        Teams
      </Text>
      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Manage your teams and their members."
          title="Team Management"
        />
        {teams.length > 10 && (
          <Box className="border-b border-gray-100 px-6 py-4 dark:border-dark-100">
            <Box className="relative max-w-md">
              <SearchIcon className="text-gray-400 absolute left-3 top-1/2 h-4 -translate-y-1/2" />
              <Input
                className="pl-9"
                onChange={(e) => {
                  setSearch(e.target.value);
                }}
                placeholder="Search teams..."
                type="search"
                value={search}
              />
            </Box>
          </Box>
        )}

        {teams.length === 0 && (
          <Flex
            align="center"
            className="px-6 py-10"
            direction="column"
            justify="center"
          >
            <TeamIcon className="h-12 w-auto" />
            <Text className="mt-4 text-lg font-semibold">No teams found</Text>
            <Text className="mb-3" color="muted">
              Create a team to get started.
            </Text>
            <Button
              color="tertiary"
              href="/settings/workspace/teams/create"
              leftIcon={<PlusIcon className="h-[1.1rem]" />}
            >
              Create Team
            </Button>
          </Flex>
        )}

        {filteredTeams.map((team) => (
          <WorkspaceTeam {...team} key={team.id} />
        ))}
      </Box>
    </Box>
  );
};
