"use client";

import { useState } from "react";
import { Box, Command, Divider, Flex, Popover, Text } from "ui";
import { CheckIcon } from "icons";
import { useParams } from "next/navigation";
import { SectionHeader } from "@/modules/settings/components/section-header";
import { useGitHubTeamSettings } from "@/lib/hooks/github/use-team-settings";
import { useUpdateGitHubTeamSettings } from "@/lib/hooks/github/use-update-team-settings";
import { useTeamStatuses } from "@/lib/hooks/statuses";
import { StoryStatusIcon } from "@/components/ui/story-status-icon";

const orderedEvents = [
  { key: "draft_pr_open", label: "On draft PR open, move to..." },
  { key: "pr_open", label: "On PR open, move to..." },
  {
    key: "pr_review_activity",
    label: "On PR review request or activity, move to...",
  },
  { key: "pr_ready_for_merge", label: "On PR ready for merge, move to..." },
  { key: "pr_merge", label: "On PR merge, move to..." },
  { key: "issue_open", label: "On GitHub issue open, move to..." },
  { key: "issue_reopen", label: "On GitHub issue reopen, move to..." },
  { key: "issue_close", label: "On GitHub issue close, move to..." },
  {
    key: "commit_close",
    label: "On commit with closing keyword, move to...",
  },
] as const;

const StatusPicker = ({
  value,
  statuses,
  onChange,
}: {
  value: string;
  statuses: { id: string; name: string }[];
  onChange: (statusId: string) => void;
}) => {
  const [open, setOpen] = useState(false);
  const selected = statuses.find((s) => s.id === value);

  return (
    <Popover onOpenChange={setOpen} open={open}>
      <Popover.Trigger asChild>
        <button
          className="border-border hover:bg-surface-muted flex w-48 items-center gap-2 rounded-lg border px-3 py-1.5 transition-colors"
          type="button"
        >
          {value !== "none" && selected ? (
            <>
              <StoryStatusIcon statusId={selected.id} />
              <Text className="truncate">{selected.name}</Text>
            </>
          ) : (
            <>
              <StoryStatusIcon className="opacity-30" />
              <Text>No action</Text>
            </>
          )}
        </button>
      </Popover.Trigger>
      <Popover.Content align="end" className="w-64">
        <Command>
          <Command.Input autoFocus placeholder="Search statuses..." />
          <Divider className="my-2" />
          <Command.Empty className="py-2">
            <Text color="muted">No statuses found.</Text>
          </Command.Empty>
          <Command.Group className="max-h-60 overflow-y-auto">
            <Command.Item
              active={value === "none"}
              className="justify-between"
              onSelect={() => {
                onChange("none");
                setOpen(false);
              }}
              value="No action"
            >
              <Flex align="center" gap={2}>
                <StoryStatusIcon className="opacity-30" />
                <Text>No action</Text>
              </Flex>
              {value === "none" && (
                <CheckIcon className="h-5 w-auto shrink-0" strokeWidth={2.1} />
              )}
            </Command.Item>
            {statuses.map((status) => (
              <Command.Item
                active={value === status.id}
                className="justify-between"
                key={status.id}
                onSelect={() => {
                  onChange(status.id);
                  setOpen(false);
                }}
                value={status.name}
              >
                <Flex align="center" gap={2}>
                  <StoryStatusIcon statusId={status.id} />
                  <Text className="truncate">{status.name}</Text>
                </Flex>
                {value === status.id && (
                  <CheckIcon
                    className="h-5 w-auto shrink-0"
                    strokeWidth={2.1}
                  />
                )}
              </Command.Item>
            ))}
          </Command.Group>
        </Command>
      </Popover.Content>
    </Popover>
  );
};

export const GitHubAutomations = () => {
  const { teamId } = useParams<{ teamId: string }>();
  const { data: teamSettings } = useGitHubTeamSettings(teamId);
  const { data: statuses = [] } = useTeamStatuses(teamId);
  const updateTeamSettings = useUpdateGitHubTeamSettings(teamId);

  const rules = teamSettings?.rules ?? [];

  const handleChange = (eventKey: string, nextValue: string) => {
    const nextRules = orderedEvents.map((item) => {
      const existing = rules.find((entry) => entry.eventKey === item.key) ?? {
        id: `${item.key}-draft`,
        eventKey: item.key,
        isActive: true,
      };
      if (item.key !== eventKey) {
        return {
          eventKey: existing.eventKey,
          targetStatusId: existing.targetStatusId ?? null,
          baseBranchPattern: existing.baseBranchPattern ?? null,
          isActive: existing.isActive,
        };
      }
      return {
        eventKey: existing.eventKey,
        targetStatusId: nextValue === "none" ? null : nextValue,
        baseBranchPattern: existing.baseBranchPattern ?? null,
        isActive: existing.isActive,
      };
    });
    updateTeamSettings.mutate({ rules: nextRules });
  };

  return (
    <Box className="border-border bg-surface mt-6 rounded-2xl border">
      <SectionHeader
        description="Configure how GitHub pull requests and issues move stories through this team's workflow."
        title="GitHub automations"
      />

      <Box className="divide-border divide-y-[0.5px]">
        {orderedEvents.map((event) => {
          const rule = rules.find((item) => item.eventKey === event.key);
          const value = rule?.targetStatusId ?? "none";
          return (
            <Flex
              align="center"
              className="px-6 py-4"
              justify="between"
              key={event.key}
            >
              <Box>
                <Text className="font-medium">{event.label}</Text>
                <Text color="muted">
                  Leave as no action to keep the story in its current status.
                </Text>
              </Box>
              <StatusPicker
                onChange={(statusId) => handleChange(event.key, statusId)}
                statuses={statuses}
                value={value}
              />
            </Flex>
          );
        })}
      </Box>
    </Box>
  );
};
