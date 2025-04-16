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
      <Flex align="center" className="mb-5" justify="between">
        <Text as="h1" className="text-2xl font-medium">
          Teams
        </Text>
        {teams.length > 1 && (
          <Input
            className="w-72 rounded-lg"
            leftIcon={<SearchIcon className="h-4" />}
            onChange={(e) => {
              setSearch(e.target.value);
            }}
            placeholder="Search teams..."
            type="search"
            value={search}
            variant="solid"
          />
        )}
      </Flex>
      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          action={
            <Button
              className="shrink-0"
              color="tertiary"
              href="/settings/workspace/teams/create"
              leftIcon={
                <PlusIcon className="h-[1.1rem] text-black dark:text-white" />
              }
            >
              Create Team
            </Button>
          }
          description="Manage your teams and their members."
          title="Team Management"
        />

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

        <Flex
          align="center"
          className="hidden border-b-[0.5px] border-gray-100 px-6 py-5 dark:border-dark-100 md:flex"
          justify="between"
        >
          <Text>Name</Text>
          <Flex align="center" gap={3} justify="between">
            <Text className="w-32">Members</Text>
            <Text className="w-32">Created</Text>
            <Box className="w-[2.1rem]" />
          </Flex>
        </Flex>

        {filteredTeams.map((team) => (
          <WorkspaceTeam {...team} key={team.id} />
        ))}
      </Box>
    </Box>
  );
};
