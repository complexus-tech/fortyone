"use client";

import { useState } from "react";
import {
  Badge,
  Box,
  Button,
  Command,
  Divider,
  Flex,
  Menu,
  Popover,
  Skeleton,
  Text,
} from "ui";
import {
  ArrowDownIcon,
  CheckIcon,
  LoadingIcon,
  MoreHorizontalIcon,
  PlusIcon,
  UnlinkIcon,
} from "icons";
import { TeamColor } from "@/components/ui/team-color";
import {
  useSlackChannelAudiences,
  useUpdateSlackChannelAudience,
} from "@/lib/hooks/slack";
import { SectionHeader } from "@/modules/settings/components";
import { useTeams } from "@/modules/teams/hooks/teams";
import type { Team } from "@/modules/teams/types";
import type { SlackChannelAudience } from "./types";

const getAudienceSummary = (teams: Team[], selectedTeamIds: string[]) => {
  if (selectedTeamIds.length === 0) return "Personal only";

  const selectedIds = new Set(selectedTeamIds);
  const selectedTeams = teams.filter((team) => selectedIds.has(team.id));
  if (selectedTeams.length === 0) return "Personal only";
  if (selectedTeams.length === 1) return selectedTeams[0].name;
  return `${selectedTeams[0].name} +${selectedTeams.length - 1}`;
};

const TeamAudiencePicker = ({
  channelName,
  disabled,
  onChange,
  privateTeams,
  publicTeams,
  selectedTeamIds,
}: {
  channelName: string;
  disabled: boolean;
  onChange: (teamIds: string[]) => void;
  privateTeams: Team[];
  publicTeams: Team[];
  selectedTeamIds: string[];
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const selectedIds = new Set(selectedTeamIds);
  const summary = getAudienceSummary(publicTeams, selectedTeamIds);

  return (
    <Popover onOpenChange={setIsOpen} open={isOpen}>
      <Popover.Trigger asChild>
        <Button
          aria-label={`Choose work access for #${channelName}`}
          className="w-max max-w-64 min-w-0 justify-between"
          color="tertiary"
          disabled={disabled}
          rightIcon={
            <ArrowDownIcon
              aria-hidden="true"
              className="h-3.5 w-auto shrink-0"
            />
          }
          size="sm"
          title={summary}
          variant="outline"
        >
          <span className="max-w-48 truncate">{summary}</span>
        </Button>
      </Popover.Trigger>
      <Popover.Content align="end" className="w-80">
        <Command>
          <Command.Input autoFocus placeholder="Search teams..." />
          <Divider className="my-2" />
          <Command.Empty className="py-2">
            <Text color="muted">No teams found.</Text>
          </Command.Empty>
          <Command.List className="mt-0 max-h-72 w-full overflow-y-auto border-0 bg-transparent py-0 shadow-none backdrop-blur-none dark:bg-transparent">
            <Command.Group>
              <Command.Item
                active={selectedTeamIds.length === 0}
                aria-checked={selectedTeamIds.length === 0}
                className="min-h-10 justify-between py-2.5"
                disabled={disabled}
                onSelect={() => {
                  if (selectedTeamIds.length > 0) {
                    onChange([]);
                  }
                  setIsOpen(false);
                }}
                value="Personal only"
              >
                <Text>Personal only</Text>
                {selectedTeamIds.length === 0 ? (
                  <CheckIcon
                    aria-hidden="true"
                    className="h-5 w-auto shrink-0"
                    strokeWidth={2.1}
                  />
                ) : null}
              </Command.Item>

              {publicTeams.map((team) => {
                const isSelected = selectedIds.has(team.id);
                return (
                  <Command.Item
                    active={isSelected}
                    aria-checked={isSelected}
                    className="min-h-10 justify-between gap-4 py-2.5"
                    disabled={disabled}
                    key={team.id}
                    onSelect={() => {
                      onChange(
                        isSelected
                          ? selectedTeamIds.filter(
                              (selectedId) => selectedId !== team.id,
                            )
                          : [...selectedTeamIds, team.id],
                      );
                    }}
                    value={`${team.name} ${team.code}`}
                  >
                    <Flex align="center" className="min-w-0" gap={2}>
                      <TeamColor
                        className="size-2.5 shrink-0 rounded-sm"
                        color={team.color}
                      />
                      <Text className="truncate">{team.name}</Text>
                    </Flex>
                    {isSelected ? (
                      <CheckIcon
                        aria-hidden="true"
                        className="h-5 w-auto shrink-0"
                        strokeWidth={2.1}
                      />
                    ) : null}
                  </Command.Item>
                );
              })}
            </Command.Group>
            {privateTeams.length > 0 ? (
              <>
                <Command.Separator />
                <Box className="px-3 pb-1">
                  <Text className="font-medium" color="muted">
                    Private teams
                  </Text>
                </Box>
                <Command.Group>
                  {privateTeams.map((team) => (
                    <Command.Item
                      className="min-h-11 gap-2 py-2.5"
                      disabled
                      key={team.id}
                      value={`${team.name} ${team.code} private`}
                    >
                      <TeamColor
                        className="size-2.5 shrink-0 rounded-sm"
                        color={team.color}
                      />
                      <Text className="truncate">{team.name}</Text>
                    </Command.Item>
                  ))}
                </Command.Group>
              </>
            ) : null}
          </Command.List>
        </Command>
      </Popover.Content>
    </Popover>
  );
};

const AddChannelPicker = ({
  audiences,
}: {
  audiences: SlackChannelAudience[];
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const updateAudience = useUpdateSlackChannelAudience();

  return (
    <Popover onOpenChange={setIsOpen} open={isOpen}>
      <Popover.Trigger asChild>
        <Button
          className="shrink-0"
          color="tertiary"
          disabled={audiences.length === 0 || updateAudience.isPending}
          leftIcon={
            <PlusIcon aria-hidden="true" className="size-3.5 shrink-0" />
          }
          size="sm"
        >
          Add channel
        </Button>
      </Popover.Trigger>
      <Popover.Content align="end" className="w-80">
        <Command>
          <Command.Input autoFocus placeholder="Search Slack channels..." />
          <Divider className="my-2" />
          <Command.Empty className="py-2">
            <Text color="muted">No channels found.</Text>
          </Command.Empty>
          <Command.List className="mt-0 max-h-80 w-full overflow-y-auto border-0 bg-transparent py-0 shadow-none backdrop-blur-none md:max-h-100 dark:bg-transparent">
            <Command.Group>
              {audiences.map((audience) => (
                <Command.Item
                  className="min-h-11 justify-between gap-4 py-2.5"
                  key={audience.channel.id}
                  onSelect={() => {
                    setIsOpen(false);
                    updateAudience.mutate({
                      channelId: audience.channel.slackChannelId,
                      isConfigured: true,
                      teamIds: audience.teamIds,
                    });
                  }}
                  value={`${audience.channel.name} ${
                    audience.channel.isPrivate ? "private" : "public"
                  }`}
                >
                  <Text className="min-w-0 truncate font-medium">
                    #{audience.channel.name}
                  </Text>
                  <Text className="shrink-0" color="muted">
                    {audience.channel.isPrivate ? "Private" : "Public"}
                  </Text>
                </Command.Item>
              ))}
            </Command.Group>
          </Command.List>
        </Command>
      </Popover.Content>
    </Popover>
  );
};

const SlackChannelAudienceRow = ({
  audience,
  privateTeams,
  publicTeams,
}: {
  audience: SlackChannelAudience;
  privateTeams: Team[];
  publicTeams: Team[];
}) => {
  const { channel } = audience;
  const publicTeamIds = new Set(publicTeams.map((team) => team.id));
  const selectedTeamIds = audience.teamIds.filter((teamId) =>
    publicTeamIds.has(teamId),
  );
  const updateAudience = useUpdateSlackChannelAudience(channel.slackChannelId);

  return (
    <Flex align="center" className="gap-4 px-6 py-4" justify="between" wrap>
      <Flex align="center" className="min-w-64 flex-1" gap={2} wrap>
        <Text className="max-w-full truncate font-medium">#{channel.name}</Text>
        <Badge color="tertiary" variant="outline">
          {channel.isPrivate ? "Private" : "Public"}
        </Badge>
        {channel.isArchived ? (
          <Badge color="tertiary" variant="outline">
            Archived
          </Badge>
        ) : null}
      </Flex>

      <Flex
        align="center"
        aria-label={`Actions for #${channel.name}`}
        className="min-w-0 gap-1"
        role="group"
      >
        <TeamAudiencePicker
          channelName={channel.name}
          disabled={updateAudience.isPending}
          onChange={(teamIds) => {
            updateAudience.mutate({
              channelId: channel.slackChannelId,
              isConfigured: true,
              teamIds,
            });
          }}
          privateTeams={privateTeams}
          publicTeams={publicTeams}
          selectedTeamIds={selectedTeamIds}
        />
        {updateAudience.isPending ? (
          <span
            aria-live="polite"
            className="inline-flex size-5 shrink-0 items-center justify-center"
          >
            <LoadingIcon
              aria-label={`Saving #${channel.name}`}
              className="size-4 animate-spin"
            />
          </span>
        ) : null}
        <Menu>
          <Menu.Button>
            <Button
              aria-label={`Channel options for #${channel.name}`}
              asIcon
              color="tertiary"
              disabled={updateAudience.isPending}
              size="sm"
              variant="naked"
            >
              <MoreHorizontalIcon aria-hidden="true" className="h-5 w-auto" />
            </Button>
          </Menu.Button>
          <Menu.Items align="end">
            <Menu.Group>
              <Menu.Item
                className="text-danger"
                onSelect={() => {
                  updateAudience.mutate({
                    channelId: channel.slackChannelId,
                    isConfigured: false,
                    teamIds: [],
                  });
                }}
              >
                <UnlinkIcon aria-hidden="true" className="text-danger" />
                Remove channel
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>
    </Flex>
  );
};

export const SlackChannelAudienceSettings = () => {
  const audiencesQuery = useSlackChannelAudiences();
  const teamsQuery = useTeams();
  const audiences = audiencesQuery.data ?? [];
  const publicTeams = (teamsQuery.data ?? []).filter((team) => !team.isPrivate);
  const privateTeams = (teamsQuery.data ?? []).filter((team) => team.isPrivate);
  const configuredAudiences = audiences.filter(
    (audience) => audience.isConfigured,
  );
  const availableAudiences = audiences.filter(
    (audience) => !audience.isConfigured && !audience.channel.isArchived,
  );
  const isPending = audiencesQuery.isPending || teamsQuery.isPending;
  const isError = audiencesQuery.isError || teamsQuery.isError;

  return (
    <Box className="border-border bg-surface mt-6 rounded-2xl border">
      <SectionHeader
        action={
          !isPending && !isError && audiences.length > 0 ? (
            <AddChannelPicker audiences={availableAudiences} />
          ) : null
        }
        description="Choose the Slack channels where Maya can share team work."
        title="Channel access"
      />

      {isPending ? (
        <Box
          aria-label="Loading Slack channel access"
          className="space-y-4 px-6 py-7"
          role="status"
        >
          <Skeleton className="h-10 w-full" />
          <Skeleton className="h-10 w-full" />
        </Box>
      ) : null}

      {!isPending && isError ? (
        <Flex
          align="center"
          className="gap-4 px-6 py-7"
          justify="between"
          role="alert"
          wrap
        >
          <Box>
            <Text className="font-medium">
              Couldn&apos;t load Slack channel access
            </Text>
            <Text className="mt-1" color="muted">
              Try again before changing which teams Maya can use in Slack.
            </Text>
          </Box>
          <Button
            color="tertiary"
            loading={audiencesQuery.isFetching || teamsQuery.isFetching}
            onClick={() => {
              void Promise.all([
                audiencesQuery.refetch(),
                teamsQuery.refetch(),
              ]);
            }}
            size="sm"
            variant="outline"
          >
            Try again
          </Button>
        </Flex>
      ) : null}

      {!isPending && !isError && audiences.length === 0 ? (
        <Box className="px-6 py-8">
          <Text className="font-medium">No Slack channels synced</Text>
          <Text className="mt-1" color="muted">
            Select Resync channels above to load workspace channels.
          </Text>
        </Box>
      ) : null}

      {!isPending &&
      !isError &&
      audiences.length > 0 &&
      configuredAudiences.length === 0 ? (
        <Box className="px-6 py-8">
          <Text className="font-medium">No channels configured</Text>
          <Text className="mt-1" color="muted">
            Add a channel to choose the work Maya can share there.
          </Text>
        </Box>
      ) : null}

      {!isPending && !isError && configuredAudiences.length > 0 ? (
        <Box className="divide-border divide-y">
          {configuredAudiences.map((audience) => (
            <SlackChannelAudienceRow
              audience={audience}
              key={audience.channel.id}
              privateTeams={privateTeams}
              publicTeams={publicTeams}
            />
          ))}
        </Box>
      ) : null}
    </Box>
  );
};
