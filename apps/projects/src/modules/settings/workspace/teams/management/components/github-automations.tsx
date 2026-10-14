"use client";

import { Box, Flex, Select, Text } from "ui";
import { useParams } from "next/navigation";
import { SectionHeader } from "@/modules/settings/components/section-header";
import { useGitHubTeamSettings } from "@/lib/hooks/github/use-team-settings";
import { useUpdateGitHubTeamSettings } from "@/lib/hooks/github/use-update-team-settings";
import { useTeamStatuses } from "@/lib/hooks/statuses";
import { StoryStatusIcon } from "@/components/ui/story-status-icon";
import { useTerminology } from "@/hooks";

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
  label,
  value,
  statuses,
  disabled,
  placeholder,
  onChange,
}: {
  label: string;
  value: string;
  statuses: { id: string; name: string }[];
  disabled: boolean;
  placeholder?: string;
  onChange: (statusId: string) => void;
}) => {
  const selected = statuses.find((status) => status.id === value);
  const selectedLabel =
    placeholder ??
    (value === "none" ? "No action" : selected?.name ?? "Unavailable status");

  return (
    <Select disabled={disabled} onValueChange={onChange} value={value}>
      <Select.Trigger
        aria-label={label}
        className="w-full shrink-0 text-base disabled:cursor-not-allowed disabled:opacity-50 sm:w-48"
      >
        <Select.Input>
          <span className="flex min-w-0 items-center gap-2">
            {selected && !placeholder ? (
              <StoryStatusIcon
                className="size-4 shrink-0"
                statusId={selected.id}
              />
            ) : (
              <StoryStatusIcon className="size-4 shrink-0 opacity-30" />
            )}
            <span className="truncate">{selectedLabel}</span>
          </span>
        </Select.Input>
      </Select.Trigger>
      <Select.Content>
        <Select.Option className="text-base" value="none">
          <span className="flex min-w-0 items-center gap-2">
            <StoryStatusIcon className="size-4 shrink-0 opacity-30" />
            <span className="truncate">No action</span>
          </span>
        </Select.Option>
        {statuses.map((status) => (
          <Select.Option
            className="text-base"
            key={status.id}
            value={status.id}
          >
            <span className="flex min-w-0 items-center gap-2">
              <StoryStatusIcon
                className="size-4 shrink-0"
                statusId={status.id}
              />
              <span className="truncate">{status.name}</span>
            </span>
          </Select.Option>
        ))}
      </Select.Content>
    </Select>
  );
};

export const GitHubAutomations = () => {
  const { getTermDisplay } = useTerminology();
  const storyTerm = getTermDisplay("storyTerm");
  const storyTermPlural = getTermDisplay("storyTerm", { variant: "plural" });
  const { teamId } = useParams<{ teamId: string }>();
  const settingsQuery = useGitHubTeamSettings(teamId);
  const statusesQuery = useTeamStatuses(teamId);
  const teamSettings = settingsQuery.data;
  const statuses = statusesQuery.data ?? [];
  const updateTeamSettings = useUpdateGitHubTeamSettings(teamId);

  const rules = teamSettings?.rules ?? [];
  const loading = settingsQuery.isPending || statusesQuery.isPending;
  const unavailable =
    settingsQuery.isError ||
    statusesQuery.isError ||
    !teamSettings ||
    !statusesQuery.data;
  const disabled = loading || unavailable || updateTeamSettings.isPending;
  let placeholder: string | undefined;
  if (loading) placeholder = "Loading settings...";
  else if (unavailable) placeholder = "Settings unavailable";

  const handleChange = (eventKey: string, nextValue: string) => {
    if (disabled) return;

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
    <Box className="border-border bg-surface rounded-2xl border">
      <SectionHeader
        description={`Configure how GitHub pull requests and issues move ${storyTermPlural} through this team's workflow.`}
        title="GitHub automations"
      />

      {!loading && unavailable ? (
        <Text className="px-6 py-4" color="danger" role="alert">
          Could not load GitHub automation settings. Reload the page to try
          again.
        </Text>
      ) : null}
      <Box aria-busy={loading} className="divide-border divide-y-[0.5px]">
        {orderedEvents.map((event) => {
          const rule = rules.find((item) => item.eventKey === event.key);
          const value = rule?.targetStatusId ?? "none";
          return (
            <Flex
              align="center"
              className="flex-col items-stretch gap-4 px-6 py-4 sm:flex-row sm:items-center"
              justify="between"
              key={event.key}
            >
              <Box className="min-w-0">
                <Text className="font-medium">{event.label}</Text>
                <Text color="muted">
                  Leave as no action to keep the {storyTerm} in its current
                  status.
                </Text>
              </Box>
              <StatusPicker
                disabled={disabled}
                label={event.label}
                onChange={(statusId) => {
                  handleChange(event.key, statusId);
                }}
                placeholder={placeholder}
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
