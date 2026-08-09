"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import {
  Badge,
  Box,
  Button,
  Command,
  Dialog,
  Divider,
  Flex,
  Menu,
  Popover,
  Switch,
  Text,
  TextArea,
} from "ui";
import {
  CheckIcon,
  LockIcon,
  MoreHorizontalIcon,
  RefreshIcon,
  SlackIcon,
  UnlinkIcon,
} from "icons";
import { useWorkspacePath } from "@/hooks";
import { SectionHeader } from "@/modules/settings/components";
import { useTeams } from "@/modules/teams/hooks/teams";
import type { Team } from "@/modules/teams/types";
import { TeamColor } from "@/components/ui/team-color";
import {
  useCreateSlackInstallSession,
  useDisconnectSlackWorkspace,
  useResyncSlackChannels,
  useSlackAccountLinkToken,
  useSlackAgentSettings,
  useSlackChannelAudiences,
  useSlackIntegration,
  useUpdateSlackAgentSettings,
  useUpdateSlackChannelAudience,
} from "@/lib/hooks/slack";
import type { SlackAgentSettings, SlackChannelAudience } from "./types";

const ChannelAudienceSelector = ({
  audience,
  teams,
}: {
  audience: SlackChannelAudience;
  teams: Team[];
}) => {
  const updateAudience = useUpdateSlackChannelAudience();
  const [query, setQuery] = useState("");

  const selectedTeamIds = audience.teamIds;
  const selectedTeams = teams.filter((team) =>
    selectedTeamIds.includes(team.id),
  );
  const filteredTeams = useMemo(() => {
    const normalizedQuery = query.trim().toLowerCase();
    if (!normalizedQuery) return teams;
    return teams.filter((team) =>
      `${team.name} ${team.code}`.toLowerCase().includes(normalizedQuery),
    );
  }, [query, teams]);

  const updateTeamIds = (teamIds: string[]) => {
    updateAudience.mutate({
      channelId: audience.channel.slackChannelId,
      teamIds,
    });
  };

  return (
    <Popover>
      <Popover.Trigger asChild>
        <Button
          className="min-w-44 justify-between"
          color="tertiary"
          loading={updateAudience.isPending}
        >
          {selectedTeams.length === 0
            ? "Public teams only"
            : selectedTeams.length === 1
              ? selectedTeams[0].name
              : `${selectedTeams.length} teams`}
        </Button>
      </Popover.Trigger>
      <Popover.Content align="end" className="w-80 p-0">
        <Command>
          <Command.Input
            onValueChange={setQuery}
            placeholder="Search teams..."
            value={query}
          />
          <Divider className="my-2" />
          <Command.List className="max-h-72 w-full overflow-y-auto border-0 bg-transparent py-0 shadow-none">
            <Command.Empty>No teams found</Command.Empty>
            <Command.Group>
              <Command.Item
                className="justify-between"
                onSelect={() => {
                  updateTeamIds([]);
                }}
              >
                <Box>
                  <Text>Public teams only</Text>
                  <Text color="muted">
                    Use the requesting member&apos;s public teams
                  </Text>
                </Box>
                {selectedTeamIds.length === 0 ? (
                  <CheckIcon className="h-4 w-auto shrink-0" />
                ) : null}
              </Command.Item>
              {filteredTeams.map((team) => {
                const isSelected = selectedTeamIds.includes(team.id);
                return (
                  <Command.Item
                    className="justify-between"
                    key={team.id}
                    onSelect={() => {
                      updateTeamIds(
                        isSelected
                          ? selectedTeamIds.filter((id) => id !== team.id)
                          : [...selectedTeamIds, team.id],
                      );
                    }}
                  >
                    <Flex align="center" gap={2}>
                      <TeamColor color={team.color} />
                      <Text>{team.name}</Text>
                      {team.isPrivate ? (
                        <LockIcon className="text-icon-muted h-4 w-auto" />
                      ) : null}
                    </Flex>
                    {isSelected ? (
                      <CheckIcon className="h-4 w-auto shrink-0" />
                    ) : null}
                  </Command.Item>
                );
              })}
            </Command.Group>
          </Command.List>
        </Command>
      </Popover.Content>
    </Popover>
  );
};

const SlackChannelAccess = () => {
  const { data: audiences = [], isLoading } = useSlackChannelAudiences();
  const { data: teams = [] } = useTeams();
  const resyncChannels = useResyncSlackChannels();

  return (
    <Box className="border-border bg-surface mt-6 rounded-2xl border">
      <SectionHeader
        action={
          <Button
            color="tertiary"
            leftIcon={<RefreshIcon className="h-4 w-auto" />}
            loading={resyncChannels.isPending}
            onClick={() => {
              resyncChannels.mutate();
            }}
          >
            Refresh channels
          </Button>
        }
        description="Choose which FortyOne teams the assistant may use in each Slack channel. Without an explicit mapping, only public teams the requesting member belongs to are available."
        title="Channel access"
      />

      {isLoading ? (
        <Box className="px-6 py-8">
          <Text color="muted">Loading Slack channels...</Text>
        </Box>
      ) : audiences.length === 0 ? (
        <Box className="px-6 py-8">
          <Text className="font-medium">No Slack channels found</Text>
          <Text className="mt-1" color="muted">
            Invite FortyOne to a channel, then refresh the channel list.
          </Text>
        </Box>
      ) : (
        <Box>
          {audiences.map((audience, index) => (
            <Flex
              align="center"
              className={
                index === 0 ? "px-6 py-4" : "border-border border-t px-6 py-4"
              }
              gap={4}
              justify="between"
              key={audience.channel.slackChannelId}
            >
              <Box className="min-w-0">
                <Flex align="center" gap={2}>
                  <Text className="truncate font-medium">
                    #{audience.channel.name}
                  </Text>
                  {audience.channel.isPrivate ? (
                    <LockIcon className="text-icon-muted h-4 w-auto shrink-0" />
                  ) : null}
                </Flex>
                <Text color="muted">
                  {audience.teamIds.length === 0
                    ? "Uses public teams the requesting member can access"
                    : "Restricted to the selected teams and each member's permissions"}
                </Text>
              </Box>
              <ChannelAudienceSelector audience={audience} teams={teams} />
            </Flex>
          ))}
        </Box>
      )}
    </Box>
  );
};

const SlackAgentConfiguration = () => {
  const { data: settings, isLoading } = useSlackAgentSettings();
  const updateSettings = useUpdateSlackAgentSettings();
  const [draft, setDraft] = useState<SlackAgentSettings | null>(null);

  const effectiveDraft = draft ?? settings ?? null;

  const updateDraft = (updates: Partial<SlackAgentSettings>) => {
    setDraft((current) => {
      const base = current ?? settings;
      return base ? { ...base, ...updates } : current;
    });
  };

  const hasChanges = Boolean(
    settings &&
      effectiveDraft &&
      (settings.assistantEnabled !== effectiveDraft.assistantEnabled ||
        settings.workflowActionsEnabled !==
          effectiveDraft.workflowActionsEnabled ||
        settings.guidance !== effectiveDraft.guidance),
  );

  return (
    <Box className="border-border bg-surface mt-6 rounded-2xl border">
      <SectionHeader
        description="Control Maya's Slack behavior and give it workspace-specific guidance. Team and channel permissions are still enforced for every message and action."
        title="Slack agent"
      />

      {isLoading || !effectiveDraft ? (
        <Box className="px-6 py-8">
          <Text color="muted">Loading Slack agent settings...</Text>
        </Box>
      ) : (
        <Box className="px-6 py-5">
          <Flex align="center" gap={6} justify="between">
            <Box>
              <Text className="font-medium">Enable Maya in Slack</Text>
              <Text color="muted">
                Answer mentions, direct messages, and subscribed thread replies.
              </Text>
            </Box>
            <Switch
              aria-label="Enable Maya in Slack"
              checked={effectiveDraft.assistantEnabled}
              onCheckedChange={(checked) => {
                updateDraft({ assistantEnabled: checked });
              }}
            />
          </Flex>

          <Divider className="my-5" />

          <Flex align="center" gap={6} justify="between">
            <Box>
              <Text className="font-medium">Allow confirmed story actions</Text>
              <Text color="muted">
                Maya may prepare story creation and updates, but a person must
                confirm every change in Slack.
              </Text>
            </Box>
            <Switch
              aria-label="Allow confirmed story actions"
              checked={effectiveDraft.workflowActionsEnabled}
              disabled={!effectiveDraft.assistantEnabled}
              onCheckedChange={(checked) => {
                updateDraft({ workflowActionsEnabled: checked });
              }}
            />
          </Flex>

          <Divider className="my-5" />

          <TextArea
            className="min-h-32 resize-y py-3 leading-6"
            label="Workspace guidance"
            maxLength={4000}
            onChange={(event) => {
              updateDraft({ guidance: event.target.value });
            }}
            placeholder="Explain terminology, response style, or operating rules Maya should follow in Slack."
            value={effectiveDraft.guidance}
          />
          <Flex align="center" className="mt-3" justify="between">
            <Text color="muted">
              {effectiveDraft.guidance.length}/4000 characters
            </Text>
            <Button
              color="invert"
              disabled={!hasChanges}
              loading={updateSettings.isPending}
              onClick={() => {
                updateSettings.mutate(effectiveDraft, {
                  onSuccess: (response) => {
                    if (response.data) setDraft(response.data);
                  },
                });
              }}
            >
              Save agent settings
            </Button>
          </Flex>
        </Box>
      )}
    </Box>
  );
};

export const SlackIntegrationSettings = () => {
  const { data: integration } = useSlackIntegration();
  const { withWorkspace } = useWorkspacePath();

  const createInstallSession = useCreateSlackInstallSession();
  const disconnectWorkspace = useDisconnectSlackWorkspace();
  useSlackAccountLinkToken();
  const [isDisconnectOpen, setIsDisconnectOpen] = useState(false);

  const isConnected = Boolean(integration?.slackWorkspace?.isActive);
  const slackWorkspace = integration?.slackWorkspace;

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
                color="invert"
                onClick={() => {
                  createInstallSession.mutate();
                }}
              >
                {isConnected ? "Reconnect Slack" : "Connect Slack"}
              </Button>
            </Flex>
          }
          description="Connect Slack so people can create FortyOne stories and requests from Slack."
          title="Connected workspace"
        />

        {!slackWorkspace ? (
          <Box className="px-6 py-8">
            <Text className="font-medium">No Slack workspace connected</Text>
            <Text className="mt-1" color="muted">
              Connect Slack to create stories from slash commands and message
              actions.
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
                  {slackWorkspace.slackTeamName}
                </Text>
                <Text color="muted">{slackWorkspace.slackTeamDomain}</Text>
              </Box>
            </Flex>
            <Flex align="center" gap={2}>
              <Badge
                className="uppercase"
                color={isConnected ? "success" : "tertiary"}
              >
                {isConnected ? "Connected" : "Disconnected"}
              </Badge>
              {isConnected ? (
                <Menu>
                  <Menu.Button>
                    <Button
                      className="px-2"
                      color="tertiary"
                      leftIcon={<MoreHorizontalIcon />}
                    />
                  </Menu.Button>
                  <Menu.Items align="end">
                    <Menu.Group>
                      <Menu.Item
                        onSelect={() => {
                          createInstallSession.mutate();
                        }}
                      >
                        <SlackIcon className="h-4 w-4" />
                        Update connection
                      </Menu.Item>
                      <Menu.Item
                        className="text-danger"
                        onSelect={() => {
                          setIsDisconnectOpen(true);
                        }}
                      >
                        <UnlinkIcon className="text-danger" />
                        Disconnect workspace
                      </Menu.Item>
                    </Menu.Group>
                  </Menu.Items>
                </Menu>
              ) : null}
            </Flex>
          </Flex>
        )}
      </Box>

      {isConnected ? (
        <>
          <SlackChannelAccess />
          <SlackAgentConfiguration />
        </>
      ) : null}

      <Box className="mt-6">
        <Link href={withWorkspace("/settings/integrations")}>
          <Text color="muted">Back to integrations</Text>
        </Link>
      </Box>

      <Dialog onOpenChange={setIsDisconnectOpen} open={isDisconnectOpen}>
        <Dialog.Content>
          <Dialog.Header>
            <Dialog.Title className="px-6 pt-0.5 text-lg">
              Disconnect Slack workspace
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body>
            <Text color="muted">
              Slash commands and message actions from this Slack workspace will
              stop creating FortyOne stories and requests.
            </Text>
          </Dialog.Body>
          <Dialog.Footer className="justify-end gap-3 border-0 pt-2">
            <Button
              className="px-4"
              color="tertiary"
              onClick={() => {
                setIsDisconnectOpen(false);
              }}
            >
              Cancel
            </Button>
            <Button
              className="px-4"
              loading={disconnectWorkspace.isPending}
              onClick={() => {
                disconnectWorkspace.mutate(undefined, {
                  onSuccess: (res) => {
                    if (!res.error) {
                      setIsDisconnectOpen(false);
                    }
                  },
                });
              }}
            >
              Disconnect
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>
    </Box>
  );
};
