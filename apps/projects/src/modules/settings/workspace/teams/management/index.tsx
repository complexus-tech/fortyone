"use client";

import { useParams } from "next/navigation";
import { Box, Text, Tabs } from "ui";
import { useTeam } from "@/modules/teams/hooks/use-team";
import { GeneralSettings } from "./components/general";
import { MembersSettings } from "./components/members";

export const TeamManagement = () => {
  const { teamId } = useParams<{ teamId: string }>();
  const { data: team } = useTeam(teamId);

  if (!team) return null;

  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Team Settings
      </Text>

      <Tabs defaultValue="general">
        <Tabs.List>
          <Tabs.Tab value="general">General</Tabs.Tab>
          <Tabs.Tab value="members">Members</Tabs.Tab>
        </Tabs.List>

        <Box className="mt-6">
          <Tabs.Panel value="general">
            <GeneralSettings team={team} />
          </Tabs.Panel>
          <Tabs.Panel value="members">
            <MembersSettings team={team} />
          </Tabs.Panel>
        </Box>
      </Tabs>
    </Box>
  );
};
