"use client";

import { Box, Flex, Select, Text } from "ui";
import { useParams } from "next/navigation";
import { SectionHeader } from "@/modules/settings/components/section-header";
import { useGitHubTeamSettings } from "@/lib/hooks/github/use-team-settings";
import { useUpdateGitHubTeamSettings } from "@/lib/hooks/github/use-update-team-settings";
import { useTeamStatuses } from "@/lib/hooks/statuses";

const orderedEvents = [
  { key: "draft_pr_open", label: "On draft PR open, move to..." },
  { key: "pr_open", label: "On PR open, move to..." },
  { key: "pr_review_activity", label: "On PR review request or activity, move to..." },
  { key: "pr_ready_for_merge", label: "On PR ready for merge, move to..." },
  { key: "pr_merge", label: "On PR merge, move to..." },
  { key: "issue_open", label: "On GitHub issue open, move to..." },
  { key: "issue_reopen", label: "On GitHub issue reopen, move to..." },
  { key: "issue_close", label: "On GitHub issue close, move to..." },
] as const;

export const GitHubAutomations = () => {
  const { teamId } = useParams<{ teamId: string }>();
  const { data: teamSettings } = useGitHubTeamSettings(teamId);
  const { data: statuses = [] } = useTeamStatuses(teamId);
  const updateTeamSettings = useUpdateGitHubTeamSettings(teamId);

  const rules = teamSettings?.rules ?? [];

  return (
    <Box className="border-border bg-surface mt-6 rounded-2xl border">
      <SectionHeader
        description="Configure how GitHub pull requests and issues move stories through this team’s workflow."
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
              <Select
                onValueChange={(nextValue) => {
                  const nextRules = orderedEvents.map((item) => {
                    const existing =
                      rules.find((entry) => entry.eventKey === item.key) ?? {
                        id: `${item.key}-draft`,
                        eventKey: item.key,
                        isActive: true,
                      };
                    if (item.key !== event.key) {
                      return {
                        eventKey: existing.eventKey,
                        targetStatusId: existing.targetStatusId ?? null,
                        baseBranchPattern: existing.baseBranchPattern ?? null,
                        isActive: existing.isActive,
                      };
                    }
                    return {
                      eventKey: existing.eventKey,
                      targetStatusId:
                        nextValue === "none" ? null : nextValue,
                      baseBranchPattern: existing.baseBranchPattern ?? null,
                      isActive: existing.isActive,
                    };
                  });
                  updateTeamSettings.mutate({ rules: nextRules });
                }}
                value={value}
              >
                <Select.Trigger className="w-64 text-[0.9rem] md:text-base">
                  <Select.Input />
                </Select.Trigger>
                <Select.Content>
                  <Select.Option value="none">No action</Select.Option>
                  {statuses.map((status) => (
                    <Select.Option key={status.id} value={status.id}>
                      {status.name}
                    </Select.Option>
                  ))}
                </Select.Content>
              </Select>
            </Flex>
          );
        })}
      </Box>
    </Box>
  );
};
