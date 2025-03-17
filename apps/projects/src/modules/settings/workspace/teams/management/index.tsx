"use client";

import { useParams } from "next/navigation";
import { Box, Text, Tabs } from "ui";
import { FilterIcon, TeamIcon, WarningIcon, WorkflowIcon } from "icons";
import { useTeam } from "@/modules/teams/hooks/use-team";
import { TeamColor } from "@/components/ui";
import { GeneralSettings } from "./components/general";
import { MembersSettings } from "./components/members";
import { WorkflowSettings } from "./components/workflows";
import { DeleteTeam } from "./components/delete";

export const TeamManagement = () => {
  const { teamId } = useParams<{ teamId: string }>();
  const { data: team } = useTeam(teamId);

  if (!team) return null;

  return (
    <Box>
      <Text
        as="h1"
        className="mb-6 flex items-center gap-2 text-2xl font-medium"
      >
        <TeamColor color={team.color} />
        {team.name}
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
            leftIcon={<WorkflowIcon className="h-[1.1rem]" />}
            value="workflows"
          >
            Workflow
          </Tabs.Tab>
          <Tabs.Tab
            leftIcon={<WarningIcon className="h-[1.1rem]" />}
            value="delete"
          >
            Danger Zone
          </Tabs.Tab>
        </Tabs.List>

        <Box className="mt-5">
          <Tabs.Panel value="general">
            <GeneralSettings team={team} />
          </Tabs.Panel>
          <Tabs.Panel value="members">
            <MembersSettings team={team} />
          </Tabs.Panel>
          <Tabs.Panel value="workflows">
            <WorkflowSettings />
          </Tabs.Panel>
          <Tabs.Panel value="delete">
            <DeleteTeam team={team} />
          </Tabs.Panel>
        </Box>
      </Tabs>
    </Box>
  );
};
