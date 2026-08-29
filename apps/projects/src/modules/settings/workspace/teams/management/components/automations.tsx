"use client";

import { useParams } from "next/navigation";
import { Box, Flex, Select, Switch, Text } from "ui";
import { useTerminology } from "@/hooks";
import { SectionHeader } from "@/modules/settings/components/section-header";
import { useTeamSettings } from "@/modules/teams/hooks/use-team-settings";
import { useUpdateStoryAutomationSettingsMutation } from "@/modules/teams/hooks/update-story-automation-settings-mutation";
import { GitHubAutomations } from "./github-automations";

export const Automations = () => {
  const { teamId } = useParams<{ teamId: string }>();
  const { getTermDisplay } = useTerminology();
  const { data: teamSettings } = useTeamSettings(teamId);
  const updateStorySettings = useUpdateStoryAutomationSettingsMutation(teamId);
  const storySettings = teamSettings?.storyAutomationSettings;

  return (
    <>
      <Box className="border-border bg-surface rounded-2xl border">
        <SectionHeader
          description={`Automate cleanup and management of ${getTermDisplay("storyTerm", { variant: "plural" })}.`}
          title={`${getTermDisplay("storyTerm", { capitalize: true, variant: "plural" })} automations`}
        />

        <Box className="divide-border divide-y-[0.5px]">
          <Flex align="center" className="px-6 py-4" justify="between">
            <Box>
              <Text className="font-medium">
                Auto-close {getTermDisplay("storyTerm", { variant: "plural" })}{" "}
                that are inactive for
              </Text>
              <Text className="line-clamp-2" color="muted">
                Automatically close{" "}
                {getTermDisplay("storyTerm", { variant: "plural" })} that
                haven&apos;t been updated
              </Text>
            </Box>
            <Switch
              checked={storySettings?.autoCloseInactiveEnabled ?? false}
              onCheckedChange={(checked) => {
                updateStorySettings.mutate({
                  autoCloseInactiveEnabled: checked,
                });
              }}
            />
          </Flex>

          {storySettings?.autoCloseInactiveEnabled ? (
            <Flex align="center" className="px-6 py-4" justify="between">
              <Box>
                <Text className="font-medium">
                  Close after being inactive for
                </Text>
                <Text className="line-clamp-2" color="muted">
                  Automatically close{" "}
                  {getTermDisplay("storyTerm", { variant: "plural" })} that
                  haven&apos;t been updated
                </Text>
              </Box>
              <Select
                onValueChange={(value) => {
                  updateStorySettings.mutate({
                    autoCloseInactiveMonths: Number.parseInt(value, 10),
                  });
                }}
                value={storySettings.autoCloseInactiveMonths.toString()}
              >
                <Select.Trigger className="w-32 text-[0.9rem] md:text-base">
                  <Select.Input />
                </Select.Trigger>
                <Select.Content>
                  <Select.Option className="text-base" value="3">
                    3 months
                  </Select.Option>
                  <Select.Option className="text-base" value="6">
                    6 months
                  </Select.Option>
                  <Select.Option className="text-base" value="12">
                    1 year
                  </Select.Option>
                </Select.Content>
              </Select>
            </Flex>
          ) : null}

          <Flex align="center" className="px-6 py-4" justify="between">
            <Box>
              <Text className="font-medium">
                Auto-archive{" "}
                {getTermDisplay("storyTerm", { variant: "plural" })}
              </Text>
              <Text className="line-clamp-2" color="muted">
                Automatically archive completed and cancelled{" "}
                {getTermDisplay("storyTerm", { variant: "plural" })} after a
                period of time
              </Text>
            </Box>
            <Switch
              checked={storySettings?.autoArchiveEnabled ?? false}
              onCheckedChange={(checked) => {
                updateStorySettings.mutate({ autoArchiveEnabled: checked });
              }}
            />
          </Flex>

          {storySettings?.autoArchiveEnabled ? (
            <Flex align="center" className="px-6 py-4" justify="between">
              <Box>
                <Text className="font-medium">Archive after</Text>
                <Text className="line-clamp-2" color="muted">
                  Automatically archive completed and cancelled{" "}
                  {getTermDisplay("storyTerm", { variant: "plural" })} that
                  haven&apos;t been updated
                </Text>
              </Box>
              <Select
                onValueChange={(value) => {
                  updateStorySettings.mutate({
                    autoArchiveMonths: Number.parseInt(value, 10),
                  });
                }}
                value={storySettings.autoArchiveMonths.toString()}
              >
                <Select.Trigger className="w-32 text-[0.9rem] md:text-base">
                  <Select.Input />
                </Select.Trigger>
                <Select.Content>
                  <Select.Option className="text-base" value="3">
                    3 months
                  </Select.Option>
                  <Select.Option className="text-base" value="6">
                    6 months
                  </Select.Option>
                  <Select.Option className="text-base" value="12">
                    1 year
                  </Select.Option>
                </Select.Content>
              </Select>
            </Flex>
          ) : null}
        </Box>
      </Box>

      <GitHubAutomations />
    </>
  );
};
