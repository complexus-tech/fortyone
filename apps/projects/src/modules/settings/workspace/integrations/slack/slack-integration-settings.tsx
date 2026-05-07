"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { Badge, Box, Button, Flex, Menu, Switch, Text } from "ui";
import { PlusIcon } from "icons";
import { TeamColor } from "@/components/ui/team-color";
import { useWorkspacePath } from "@/hooks";
import { SectionHeader } from "@/modules/settings/components";
import { useTeams } from "@/modules/teams/hooks/teams";
import {
  useCreateSlackChannelLink,
  useCreateSlackInstallSession,
  useDeleteSlackChannelLink,
  useResyncSlackChannels,
  useSlackIntegration,
  useUpdateSlackWorkspaceSettings,
} from "@/lib/hooks/slack";

const SlackIcon = ({ className = "h-5 w-5" }: { className?: string }) => (
  <svg
    className={className}
    viewBox="0 0 54 54"
    xmlns="http://www.w3.org/2000/svg"
  >
    <path
      d="M19.712.133a5.381 5.381 0 0 0-5.376 5.387 5.381 5.381 0 0 0 5.376 5.386h5.376V5.52A5.381 5.381 0 0 0 19.712.133m0 14.365H5.376A5.381 5.381 0 0 0 0 19.884a5.381 5.381 0 0 0 5.376 5.387h14.336a5.381 5.381 0 0 0 5.376-5.387 5.381 5.381 0 0 0-5.376-5.386"
      fill="#36C5F0"
    />
    <path
      d="M53.76 19.884a5.381 5.381 0 0 0-5.376-5.386 5.381 5.381 0 0 0-5.376 5.386v5.387h5.376a5.381 5.381 0 0 0 5.376-5.387m-14.336 0V5.52A5.381 5.381 0 0 0 34.048.133a5.381 5.381 0 0 0-5.376 5.387v14.364a5.381 5.381 0 0 0 5.376 5.387 5.381 5.381 0 0 0 5.376-5.387"
      fill="#2EB67D"
    />
    <path
      d="M34.048 54a5.381 5.381 0 0 0 5.376-5.387 5.381 5.381 0 0 0-5.376-5.386h-5.376v5.386A5.381 5.381 0 0 0 34.048 54m0-14.365h14.336a5.381 5.381 0 0 0 5.376-5.386 5.381 5.381 0 0 0-5.376-5.387H34.048a5.381 5.381 0 0 0-5.376 5.387 5.381 5.381 0 0 0 5.376 5.386"
      fill="#ECB22E"
    />
    <path
      d="M0 34.249a5.381 5.381 0 0 0 5.376 5.386 5.381 5.381 0 0 0 5.376-5.386v-5.387H5.376A5.381 5.381 0 0 0 0 34.25m14.336 0v14.364A5.381 5.381 0 0 0 19.712 54a5.381 5.381 0 0 0 5.376-5.387V34.25a5.381 5.381 0 0 0-5.376-5.387 5.381 5.381 0 0 0-5.376 5.386"
      fill="#E01E5A"
    />
  </svg>
);

export const SlackIntegrationSettings = () => {
  const { data: integration } = useSlackIntegration();
  const { data: teams = [] } = useTeams();
  const { withWorkspace } = useWorkspacePath();

  const createInstallSession = useCreateSlackInstallSession();
  const resyncChannels = useResyncSlackChannels();
  const updateSettings = useUpdateSlackWorkspaceSettings();
  const createLink = useCreateSlackChannelLink();
  const deleteLink = useDeleteSlackChannelLink();

  const [channelMenuOpen, setChannelMenuOpen] = useState(false);
  const [teamMenuOpen, setTeamMenuOpen] = useState(false);
  const [selectedChannelId, setSelectedChannelId] = useState("");
  const [selectedTeamId, setSelectedTeamId] = useState("");

  const linkedChannelIds = useMemo(
    () =>
      new Set(
        (integration?.channelLinks ?? []).map((link) => link.slackChannelId),
      ),
    [integration?.channelLinks],
  );
  const linkedTeamIds = useMemo(
    () => new Set((integration?.channelLinks ?? []).map((link) => link.teamId)),
    [integration?.channelLinks],
  );

  const availableChannels = useMemo(
    () =>
      (integration?.channels ?? []).filter(
        (channel) =>
          channel.isActive && !linkedChannelIds.has(channel.slackChannelId),
      ),
    [integration?.channels, linkedChannelIds],
  );

  const availableTeams = useMemo(
    () => teams.filter((team) => !linkedTeamIds.has(team.id)),
    [teams, linkedTeamIds],
  );

  const selectedChannel = availableChannels.find(
    (channel) => channel.slackChannelId === selectedChannelId,
  );
  const selectedTeam = availableTeams.find(
    (team) => team.id === selectedTeamId,
  );

  const isConnected = Boolean(integration?.slackWorkspace?.isActive);
  const createMode =
    integration?.settings.defaultCreateMode ?? "create_task_now";

  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-medium">
        Slack
      </Text>

      <Box className="border-border bg-surface rounded-2xl border">
        <SectionHeader
          action={
            <Flex gap={2}>
              <Button
                color="tertiary"
                onClick={() => {
                  resyncChannels.mutate();
                }}
              >
                Resync channels
              </Button>
              <Button
                color="invert"
                onClick={() => {
                  createInstallSession.mutate();
                }}
              >
                {isConnected ? "Reconnect Slack" : "Connect Slack"}
              </Button>
            </Flex>
          }
          description="Connect a Slack workspace and map channels to teams in FortyOne."
          title="Connected workspace"
        />

        {!integration?.slackWorkspace ? (
          <Box className="px-6 py-8">
            <Text className="font-medium">No Slack workspace connected</Text>
            <Text className="mt-1" color="muted">
              Connect Slack to create tasks from Slack and route channel actions
              to teams.
            </Text>
          </Box>
        ) : (
          <Flex align="center" className="px-6 py-4" justify="between">
            <Flex align="center" gap={3}>
              <Flex
                align="center"
                className="bg-surface-muted size-9 shrink-0 rounded-lg"
                justify="center"
              >
                <SlackIcon className="h-4.5 w-4.5" />
              </Flex>
              <Box>
                <Text className="font-medium">
                  {integration.slackWorkspace.slackTeamName}
                </Text>
                <Text color="muted">
                  {integration.slackWorkspace.slackTeamDomain}
                </Text>
              </Box>
            </Flex>
            <Badge
              className="uppercase"
              color={isConnected ? "success" : "tertiary"}
            >
              {isConnected ? "Connected" : "Disconnected"}
            </Badge>
          </Flex>
        )}
      </Box>

      <Box className="border-border bg-surface mt-6 rounded-2xl border">
        <SectionHeader
          description="Choose what happens when someone creates a task from Slack."
          title="Create behavior"
        />
        <Flex align="center" className="px-6 py-4" justify="between">
          <Box>
            <Text className="font-medium">Create task immediately</Text>
            <Text color="muted">
              When off, Slack submissions are sent to requests for review first.
            </Text>
          </Box>
          <Switch
            checked={createMode === "create_task_now"}
            onCheckedChange={(checked) => {
              updateSettings.mutate({
                defaultCreateMode: checked
                  ? "create_task_now"
                  : "send_to_requests",
              });
            }}
          />
        </Flex>
      </Box>

      <Box className="border-border bg-surface mt-6 rounded-2xl border">
        <SectionHeader
          description="Map each Slack channel to one team. Slash commands and message actions use this mapping."
          title="Channel routing"
        />

        {(integration?.channelLinks.length ?? 0) === 0 ? (
          <Box className="px-6 py-8">
            <Text className="font-medium">No channel routes configured</Text>
            <Text className="mt-1" color="muted">
              Link a Slack channel to a team so task creation can route
              correctly.
            </Text>
          </Box>
        ) : (
          <Box className="divide-border divide-y">
            {integration?.channelLinks.map((link) => {
              const channel = integration.channels.find(
                (item) => item.slackChannelId === link.slackChannelId,
              );
              return (
                <Flex
                  align="center"
                  className="px-6 py-4"
                  justify="between"
                  key={link.id}
                >
                  <Flex align="center" gap={3}>
                    <Flex
                      align="center"
                      className="bg-surface-muted size-9 shrink-0 rounded-lg"
                      justify="center"
                    >
                      <SlackIcon className="h-4 w-4" />
                    </Flex>
                    <Box>
                      <Text className="font-medium">
                        #{channel?.name ?? link.slackChannelId}
                      </Text>
                      <Flex align="center" gap={2}>
                        <TeamColor color={link.teamColor} />
                        <Text color="muted">{link.teamName}</Text>
                      </Flex>
                    </Box>
                  </Flex>
                  <Button
                    color="tertiary"
                    onClick={() => {
                      deleteLink.mutate(link.id);
                    }}
                  >
                    Unlink
                  </Button>
                </Flex>
              );
            })}
          </Box>
        )}

        <Box className="border-border border-t px-6 py-4">
          <Flex className="flex-col gap-3 md:flex-row md:items-center">
            <Menu open={channelMenuOpen} onOpenChange={setChannelMenuOpen}>
              <Menu.Button>
                <Button
                  color="tertiary"
                  className="w-full justify-between md:w-64"
                >
                  {selectedChannel
                    ? `#${selectedChannel.name}`
                    : "Select channel"}
                </Button>
              </Menu.Button>
              <Menu.Items align="start" className="max-h-72 overflow-auto">
                <Menu.Group>
                  {availableChannels.map((channel) => (
                    <Menu.Item
                      key={channel.id}
                      onSelect={() => {
                        setSelectedChannelId(channel.slackChannelId);
                      }}
                    >
                      #{channel.name}
                    </Menu.Item>
                  ))}
                </Menu.Group>
              </Menu.Items>
            </Menu>

            <Menu open={teamMenuOpen} onOpenChange={setTeamMenuOpen}>
              <Menu.Button>
                <Button
                  color="tertiary"
                  className="w-full justify-between md:w-64"
                >
                  {selectedTeam ? selectedTeam.name : "Select team"}
                </Button>
              </Menu.Button>
              <Menu.Items align="start" className="max-h-72 overflow-auto">
                <Menu.Group>
                  {availableTeams.map((team) => (
                    <Menu.Item
                      key={team.id}
                      onSelect={() => {
                        setSelectedTeamId(team.id);
                      }}
                    >
                      <Flex align="center" gap={2}>
                        <TeamColor color={team.color} />
                        <Text>{team.name}</Text>
                      </Flex>
                    </Menu.Item>
                  ))}
                </Menu.Group>
              </Menu.Items>
            </Menu>

            <Button
              color="invert"
              leftIcon={<PlusIcon />}
              onClick={() => {
                if (!selectedChannelId || !selectedTeamId) return;
                createLink.mutate({
                  slackChannelId: selectedChannelId,
                  teamId: selectedTeamId,
                });
                setSelectedChannelId("");
                setSelectedTeamId("");
              }}
            >
              Link channel
            </Button>
          </Flex>
        </Box>
      </Box>

      <Box className="mt-6">
        <Link href={withWorkspace("/settings/workspace/integrations")}>
          <Text color="muted">Back to integrations</Text>
        </Link>
      </Box>
    </Box>
  );
};
