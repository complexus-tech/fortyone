"use client";

import { useParams } from "next/navigation";
import { Box, Text, Tabs } from "ui";
import { FilterIcon, TeamIcon } from "icons";
import { useTeam } from "@/modules/teams/hooks/use-team";
import { StoryStatusIcon } from "@/components/ui";
import { GeneralSettings } from "./components/general";
import { MembersSettings } from "./components/members";

export const TeamManagement = () => {
  const { teamId } = useParams<{ teamId: string }>();
  const { data: team } = useTeam(teamId);

  if (!team) return null;

  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-medium">
        Team Settings
      </Text>

      <Tabs defaultValue="general">
        <Tabs.List className="mx-0">
          <Tabs.Tab
            leftIcon={<FilterIcon className="h-[1.1rem]" />}
            value="general"
          >
            General
          </Tabs.Tab>
          <Tabs.Tab
            leftIcon={<TeamIcon className="h-[1.1rem]" />}
            value="members"
          >
            Members
          </Tabs.Tab>
          <Tabs.Tab
            leftIcon={<StoryStatusIcon className="h-[1.1rem]" />}
            value="workflows"
          >
            Statuses workflow
          </Tabs.Tab>
        </Tabs.List>

        <Box className="mt-6">
          <Tabs.Panel value="general">
            <GeneralSettings team={team} />
          </Tabs.Panel>
          <Tabs.Panel value="members">
            <MembersSettings team={team} />
          </Tabs.Panel>
          <Tabs.Panel value="workflows">Coming soon</Tabs.Panel>
        </Box>
      </Tabs>
    </Box>
  );
};
