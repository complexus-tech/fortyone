"use client";

import { useEffect } from "react";
import { useParams } from "next/navigation";
import { Box, Flex, Text, Tabs } from "ui";
import {
  CustomFieldsIcon,
  FilterIcon,
  GitIcon,
  SprintsIcon,
  TeamIcon,
  WorkflowIcon,
} from "icons";
import { useQueryStates, parseAsString, parseAsStringLiteral } from "nuqs";
import { useTerminology, useWorkspacePath } from "@/hooks";
import { useTeam } from "@/modules/teams/hooks/use-team";
import { TeamColor } from "@/components/ui";
import { SettingsBackButton } from "@/modules/settings/components";
import { TeamCustomFieldSettings } from "@/modules/custom-fields/public/settings";
import type {
  TeamSettingsSection as Section,
  TeamSettingsTab,
} from "./navigation";
import { GeneralSettings } from "./components/general";
import { MembersSettings } from "./components/members";
import { WorkflowSettings } from "./components/workflows";
import { DeleteTeam } from "./components/delete";
import { Automations } from "./components/automations";
import { EstimationSettings } from "./components/estimation";
import { SprintSettings } from "./components/sprints";
import { TeamSettingsSection } from "./components/settings-sections";
import {
  getTeamSettingsSectionId,
  resolveTeamSettingsLocation,
  TEAM_SETTINGS_QUERY_TABS,
} from "./navigation";

const focusSection = (section: Section) => {
  const element = document.getElementById(getTeamSettingsSectionId(section));
  element?.focus({ preventScroll: true });
  element?.scrollIntoView({ block: "start" });
};

export const TeamManagement = () => {
  const { teamId } = useParams<{ teamId: string }>();
  const { withWorkspace } = useWorkspacePath();
  const { getTermDisplay } = useTerminology();
  const { data: team } = useTeam(teamId);
  const [query, setQuery] = useQueryStates({
    tab: parseAsStringLiteral(TEAM_SETTINGS_QUERY_TABS).withDefault("general"),
    section: parseAsString,
  });
  const location = resolveTeamSettingsLocation(query.tab, query.section);
  const loadedTeamId = team?.id;

  useEffect(() => {
    if (loadedTeamId && location.section) focusSection(location.section);
  }, [loadedTeamId, location.tab, location.section]);

  if (!team) return null;

  return (
    <Box>
      <Flex align="center" className="mb-6" gap={2}>
        <SettingsBackButton
          href={withWorkspace("/settings/workspace/teams")}
          label="Back to teams"
        />
        <Text as="h1" className="flex items-center gap-2 text-2xl font-medium">
          <TeamColor color={team.color} />
          {team.name}
        </Text>
      </Flex>
      <Tabs
        defaultValue="general"
        onValueChange={(value) => {
          void setQuery({ tab: value as TeamSettingsTab, section: null });
        }}
        value={location.tab}
      >
        <Box className="overflow-x-auto">
          <Tabs.List
            aria-label="Team settings"
            className="mx-0 flex-nowrap md:mx-0"
          >
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
              leftIcon={<SprintsIcon className="h-[1.1rem]" />}
              value="planning"
            >
              Planning
            </Tabs.Tab>
            <Tabs.Tab
              leftIcon={<CustomFieldsIcon className="h-[1.1rem]" />}
              value="fields"
            >
              Custom fields
            </Tabs.Tab>
            <Tabs.Tab
              leftIcon={<GitIcon className="h-[1.1rem]" />}
              value="automations"
            >
              Automations
            </Tabs.Tab>
          </Tabs.List>
        </Box>
        <Box className="mt-5">
          <Tabs.Panel value="general">
            <Box className="space-y-6">
              <TeamSettingsSection label="Team details" value="general">
                <GeneralSettings team={team} />
              </TeamSettingsSection>
              <TeamSettingsSection label="Danger zone" value="danger">
                <DeleteTeam team={team} />
              </TeamSettingsSection>
            </Box>
          </Tabs.Panel>
          <Tabs.Panel value="members">
            <MembersSettings team={team} />
          </Tabs.Panel>
          <Tabs.Panel value="workflows">
            <TeamSettingsSection label="Statuses and limits" value="workflows">
              <WorkflowSettings />
            </TeamSettingsSection>
          </Tabs.Panel>
          <Tabs.Panel value="planning">
            <Box className="space-y-6">
              <TeamSettingsSection label="Complexity" value="complexity">
                <EstimationSettings teamId={teamId} />
              </TeamSettingsSection>
              <TeamSettingsSection
                label={`${getTermDisplay("sprintTerm", { capitalize: true, variant: "plural" })} scheduling`}
                value="sprints"
              >
                <SprintSettings />
              </TeamSettingsSection>
            </Box>
          </Tabs.Panel>
          <Tabs.Panel value="fields">
            <TeamSettingsSection label="Custom fields" value="fields">
              <TeamCustomFieldSettings teamId={teamId} />
            </TeamSettingsSection>
          </Tabs.Panel>
          <Tabs.Panel value="automations">
            <Automations />
          </Tabs.Panel>
        </Box>
      </Tabs>
    </Box>
  );
};
