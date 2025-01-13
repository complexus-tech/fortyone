"use client";

import { Box, Text, Button, Input } from "ui";
import { SearchIcon } from "icons";
import { useTeams } from "@/modules/teams/hooks/teams";
import { SectionHeader } from "../../components/section-header";
import { WorkspaceTeam } from "./components/team";

export const TeamsSettings = () => {
  const { data: teams = [] } = useTeams();
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Teams
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          action={<Button color="primary">Create Team</Button>}
          description="Create and manage teams in your workspace."
          title="Team Management"
        />

        <Box className="border-b border-gray-100 px-6 py-4 dark:border-dark-100">
          <Box className="relative max-w-md">
            <SearchIcon className="text-gray-400 absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2" />
            <Input
              className="pl-9"
              placeholder="Search teams..."
              type="search"
            />
          </Box>
        </Box>

        <Box>
          {teams.map((team) => (
            <WorkspaceTeam key={team.id} {...team} />
          ))}
        </Box>
      </Box>
    </Box>
  );
};
