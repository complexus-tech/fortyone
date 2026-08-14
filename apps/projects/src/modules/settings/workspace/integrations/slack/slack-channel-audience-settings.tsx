"use client";

import { useMemo, useState } from "react";
import { Badge, Box, Button, Command, Flex, Popover, Skeleton, Text } from "ui";
import { ArrowDownIcon, CheckIcon, LockIcon } from "icons";
import { TeamColor } from "@/components/ui/team-color";
import {
  useSlackChannelAudiences,
  useUpdateSlackChannelAudience,
} from "@/lib/hooks/slack";
import { SectionHeader } from "@/modules/settings/components";
import { useTeams } from "@/modules/teams/hooks/teams";
import type { Team } from "@/modules/teams/types";
import type { SlackChannelAudience } from "./types";

const haveSameTeamIds = (left: string[], right: string[]) => {
  if (left.length !== right.length) return false;
  const rightIds = new Set(right);
  return left.every((teamId) => rightIds.has(teamId));
};

const getAudienceSummary = (teams: Team[], selectedTeamIds: string[]) => {
  if (selectedTeamIds.length === 0) return "Personal only";

  const selectedTeams = teams.filter((team) =>
    selectedTeamIds.includes(team.id),
  );
  if (selectedTeams.length === 0) return "Select public teams";
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
          aria-label={`Choose public FortyOne teams for #${channelName}`}
          className="w-64 min-w-0 justify-between"
          color="tertiary"
          disabled={disabled}
          rightIcon={<ArrowDownIcon aria-hidden="true" className="shrink-0" />}
          title={summary}
          variant="outline"
        >
          <span className="truncate">{summary}</span>
        </Button>
      </Popover.Trigger>
      <Popover.Content align="end" className="w-80 p-0">
        <Command>
          <Command.Input
            className="h-10 pr-3"
            placeholder="Search FortyOne teams..."
          />
          <Command.Separator />
          <Command.Empty>
            <Text color="muted">No teams found.</Text>
          </Command.Empty>
          <Box className="max-h-72 overflow-y-auto pb-1">
            <Command.Group>
              {publicTeams.map((team) => {
                const isSelected = selectedIds.has(team.id);
                return (
                  <Command.Item
                    aria-checked={isSelected}
                    className="justify-between gap-4"
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
                    role="option"
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
                <Box className="px-3 pt-1 pb-1.5">
                  <Text className="font-medium" color="muted">
                    Private teams
                  </Text>
                </Box>
                <Command.Group>
                  {privateTeams.map((team) => (
                    <Command.Item
                      className="justify-between gap-4"
                      disabled
                      key={team.id}
                      value={`${team.name} ${team.code} private`}
                    >
                      <Flex align="center" className="min-w-0" gap={2}>
                        <LockIcon
                          aria-hidden="true"
                          className="h-4 w-4 shrink-0"
                        />
                        <Text className="truncate">{team.name}</Text>
                      </Flex>
                      <Badge color="tertiary" size="sm" variant="outline">
                        Private
                      </Badge>
                    </Command.Item>
                  ))}
                </Command.Group>
              </>
            ) : null}
          </Box>
          <Box className="border-border border-t px-3 py-2.5">
            <Text className="leading-5" color="muted">
              Private teams cannot be included in shared Slack reports. Existing
              private mappings stay unchanged for other Slack features.
            </Text>
          </Box>
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
  const publicTeamIds = useMemo(
    () => new Set(publicTeams.map((team) => team.id)),
    [publicTeams],
  );
  const privateTeamIds = useMemo(
    () => new Set(privateTeams.map((team) => team.id)),
    [privateTeams],
  );
  const [selectedTeamIds, setSelectedTeamIds] = useState(() =>
    audience.teamIds.filter((teamId) => publicTeamIds.has(teamId)),
  );
  const updateAudience = useUpdateSlackChannelAudience();
  const preservedPrivateTeamIds = audience.teamIds.filter(
    (teamId) => !publicTeamIds.has(teamId),
  );
  const configuredPublicTeamIds = audience.teamIds.filter((teamId) =>
    publicTeamIds.has(teamId),
  );
  const privateMappingCount = audience.teamIds.filter((teamId) =>
    privateTeamIds.has(teamId),
  ).length;
  const hasChanges = !haveSameTeamIds(configuredPublicTeamIds, selectedTeamIds);
  const { channel } = audience;

  return (
    <Flex align="center" className="gap-4 px-6 py-5" justify="between" wrap>
      <Box className="min-w-64 flex-1">
        <Flex align="center" gap={2} wrap>
          <Text className="font-medium">#{channel.name}</Text>
          <Badge color="tertiary" size="sm" variant="outline">
            {channel.isPrivate
              ? "Private Slack channel"
              : "Public Slack channel"}
          </Badge>
          {channel.isArchived ? (
            <Badge color="tertiary" size="sm" variant="outline">
              Archived
            </Badge>
          ) : null}
        </Flex>
        <Text className="mt-1" color="muted">
          {selectedTeamIds.length === 0
            ? "Shared reports are off. Personal questions use the asking person's joined public teams."
            : `${selectedTeamIds.length} public ${selectedTeamIds.length === 1 ? "team" : "teams"} selected.`}
        </Text>
        {privateMappingCount > 0 ? (
          <Text className="mt-1" color="muted" role="status">
            {privateMappingCount === 1
              ? "1 existing private team mapping is preserved for other Slack features, but Maya will not use it for shared reports."
              : `${privateMappingCount} existing private team mappings are preserved for other Slack features, but Maya will not use them for shared reports.`}
          </Text>
        ) : null}
      </Box>
      <Flex align="center" className="min-w-0" gap={2}>
        <TeamAudiencePicker
          channelName={channel.name}
          disabled={publicTeams.length === 0 || updateAudience.isPending}
          onChange={setSelectedTeamIds}
          privateTeams={privateTeams}
          publicTeams={publicTeams}
          selectedTeamIds={selectedTeamIds}
        />
        <Button
          color="invert"
          disabled={!hasChanges || updateAudience.isPending}
          loading={updateAudience.isPending}
          loadingText="Saving"
          onClick={() => {
            updateAudience.mutate({
              channelId: channel.slackChannelId,
              teamIds: [...preservedPrivateTeamIds, ...selectedTeamIds],
            });
          }}
        >
          Save
        </Button>
      </Flex>
    </Flex>
  );
};

export const SlackChannelAudienceSettings = () => {
  const audiencesQuery = useSlackChannelAudiences();
  const teamsQuery = useTeams();
  const publicTeams = useMemo(
    () => (teamsQuery.data ?? []).filter((team) => !team.isPrivate),
    [teamsQuery.data],
  );
  const privateTeams = useMemo(
    () => (teamsQuery.data ?? []).filter((team) => team.isPrivate),
    [teamsQuery.data],
  );
  const isPending = audiencesQuery.isPending || teamsQuery.isPending;
  const isError = audiencesQuery.isError || teamsQuery.isError;

  return (
    <Box className="border-border bg-surface mt-6 rounded-2xl border">
      <SectionHeader
        description="Choose which public FortyOne teams Maya can use for shared work reports in each synced Slack channel."
        title="Channel access"
      />
      <Box className="border-border bg-surface-muted/40 border-b px-6 py-3.5">
        <Flex align="start" gap={2}>
          <LockIcon aria-hidden="true" className="mt-0.5 h-4 w-4 shrink-0" />
          <Text className="leading-5" color="muted">
            Personal questions can use the asking person&apos;s joined public
            teams. Cross-person reports require at least one selected public
            team. Private FortyOne teams stay excluded from channel reports;
            Maya can use one in a DM only when the requester belongs to it.
          </Text>
        </Flex>
      </Box>

      {isPending ? (
        <Box
          aria-label="Loading Slack channel access"
          className="space-y-4 px-6 py-7"
          role="status"
        >
          <Skeleton className="h-5 w-44" />
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
            variant="outline"
          >
            Try again
          </Button>
        </Flex>
      ) : null}

      {!isPending && !isError && audiencesQuery.data.length === 0 ? (
        <Box className="px-6 py-8">
          <Text className="font-medium">No Slack channels synced</Text>
          <Text className="mt-1" color="muted">
            Synced channels will appear here after Slack finishes refreshing the
            workspace connection.
          </Text>
        </Box>
      ) : null}

      {!isPending && !isError && audiencesQuery.data.length ? (
        <Box className="divide-border divide-y">
          {audiencesQuery.data.map((audience) => (
            <SlackChannelAudienceRow
              audience={audience}
              key={`${audience.channel.id}:${audience.teamIds.join(",")}`}
              privateTeams={privateTeams}
              publicTeams={publicTeams}
            />
          ))}
        </Box>
      ) : null}
    </Box>
  );
};
