"use client";

import { useParams } from "next/navigation";
import { Box, Text, Switch, Select, Flex } from "ui";
import { SectionHeader } from "@/modules/settings/components/section-header";
import { useTerminology } from "@/hooks";
import { useTeamSettings } from "@/modules/teams/hooks/use-team-settings";
import { useUpdateSprintSettingsMutation } from "@/modules/teams/hooks/update-sprint-settings-mutation";
import { useUpdateStoryAutomationSettingsMutation } from "@/modules/teams/hooks/update-story-automation-settings-mutation";

export const Automations = () => {
  const { teamId } = useParams<{ teamId: string }>();
  const { getTermDisplay } = useTerminology();
  const { data: teamSettings } = useTeamSettings(teamId);
  const updateSprintSettings = useUpdateSprintSettingsMutation(teamId);
  const updateStorySettings = useUpdateStoryAutomationSettingsMutation(teamId);
  const sprintSettings = teamSettings?.sprintSettings;
  const storySettings = teamSettings?.storyAutomationSettings;

  return (
    <>
      {/* Sprint Configuration Section */}
      <Box className="border-border bg-surface rounded-2xl border">
        <SectionHeader
          description={`Configure ${getTermDisplay("sprintTerm", { variant: "plural" })} and automation settings for your team.`}
          title={getTermDisplay("sprintTerm", {
            capitalize: true,
            variant: "plural",
          })}
        />

        <Box className="divide-border divide-y-[0.5px]">
          {/* Enable Automation Toggle */}
          <Flex align="center" className="px-6 py-4" justify="between">
            <Box>
              <Text className="font-medium">Enable automation</Text>
              <Text className="line-clamp-2" color="muted">
                Turn on automated {getTermDisplay("sprintTerm")} management and
                scheduling
              </Text>
            </Box>
            <Switch
              checked={sprintSettings?.autoCreateSprints}
              onCheckedChange={(checked) => {
                updateSprintSettings.mutate({ autoCreateSprints: checked });
              }}
            />
          </Flex>

          {/* Number of Sprints to Create */}
          {sprintSettings?.autoCreateSprints ? (
            <Flex align="center" className="px-6 py-4" justify="between">
              <Box>
                <Text className="font-medium">
                  Number of upcoming{" "}
                  {getTermDisplay("sprintTerm", { variant: "plural" })} to
                  create
                </Text>
                <Text className="line-clamp-2" color="muted">
                  How many {getTermDisplay("sprintTerm", { variant: "plural" })}{" "}
                  to create in advance for planning
                </Text>
              </Box>
              <Select
                onValueChange={(value) => {
                  updateSprintSettings.mutate({
                    upcomingSprintsCount: parseInt(value),
                  });
                }}
                value={sprintSettings.upcomingSprintsCount.toString()}
              >
                <Select.Trigger className="w-32 text-[0.9rem] md:text-base">
                  <Select.Input />
                </Select.Trigger>
                <Select.Content>
                  <Select.Option className="text-base" value="1">
                    1 {getTermDisplay("sprintTerm")}
                  </Select.Option>
                  <Select.Option className="text-base" value="2">
                    2 {getTermDisplay("sprintTerm", { variant: "plural" })}
                  </Select.Option>
                  <Select.Option className="text-base" value="3">
                    3 {getTermDisplay("sprintTerm", { variant: "plural" })}
                  </Select.Option>
                  <Select.Option className="text-base" value="4">
                    4 {getTermDisplay("sprintTerm", { variant: "plural" })}
                  </Select.Option>
                </Select.Content>
              </Select>
            </Flex>
          ) : null}

          {/* Sprint Length */}
          {sprintSettings?.autoCreateSprints ? (
            <Flex align="center" className="px-6 py-4" justify="between">
              <Box>
                <Text className="font-medium">
                  Each {getTermDisplay("sprintTerm")} lasts
                </Text>
                <Text className="line-clamp-2" color="muted">
                  Duration of each {getTermDisplay("sprintTerm")} cycle
                </Text>
              </Box>
              <Select
                onValueChange={(value) => {
                  updateSprintSettings.mutate({
                    sprintDurationWeeks: parseInt(value),
                  });
                }}
                value={sprintSettings.sprintDurationWeeks.toString()}
              >
                <Select.Trigger className="w-32 text-[0.9rem] md:text-base">
                  <Select.Input />
                </Select.Trigger>
                <Select.Content>
                  <Select.Option className="text-base" value="1">
                    1 week
                  </Select.Option>
                  <Select.Option className="text-base" value="2">
                    2 weeks
                  </Select.Option>
                  <Select.Option className="text-base" value="3">
                    3 weeks
                  </Select.Option>
                  <Select.Option className="text-base" value="4">
                    4 weeks
                  </Select.Option>
                </Select.Content>
              </Select>
            </Flex>
          ) : null}

          {/* Sprints Start On */}
          {sprintSettings?.autoCreateSprints ? (
            <Flex align="center" className="px-6 py-4" justify="between">
              <Box>
                <Text className="font-medium">
                  {getTermDisplay("sprintTerm", {
                    capitalize: true,
                    variant: "plural",
                  })}{" "}
                  start on
                </Text>
                <Text className="line-clamp-2" color="muted">
                  Which day of the week new{" "}
                  {getTermDisplay("sprintTerm", { variant: "plural" })} begin
                </Text>
              </Box>
              <Select
                onValueChange={(value) => {
                  updateSprintSettings.mutate({
                    sprintStartDay: value,
                  });
                }}
                value={sprintSettings.sprintStartDay}
              >
                <Select.Trigger className="w-32 text-[0.9rem] md:text-base">
                  <Select.Input />
                </Select.Trigger>
                <Select.Content>
                  <Select.Option className="text-base" value="Monday">
                    Monday
                  </Select.Option>
                  <Select.Option className="text-base" value="Tuesday">
                    Tuesday
                  </Select.Option>
                  <Select.Option className="text-base" value="Wednesday">
                    Wednesday
                  </Select.Option>
                  <Select.Option className="text-base" value="Thursday">
                    Thursday
                  </Select.Option>
                  <Select.Option className="text-base" value="Friday">
                    Friday
                  </Select.Option>
                </Select.Content>
              </Select>
            </Flex>
          ) : null}

          {/* Move Incomplete Stories */}
          {sprintSettings?.autoCreateSprints ? (
            <Flex align="center" className="px-6 py-4" justify="between">
              <Box>
                <Text className="font-medium">
                  Move incomplete{" "}
                  {getTermDisplay("storyTerm", { variant: "plural" })}
                </Text>
                <Text className="line-clamp-2" color="muted">
                  Move unfinished{" "}
                  {getTermDisplay("storyTerm", { variant: "plural" })} to the
                  next {getTermDisplay("sprintTerm")} automatically
                </Text>
              </Box>
              <Switch
                checked={sprintSettings.moveIncompleteStoriesEnabled}
                onCheckedChange={(checked) => {
                  updateSprintSettings.mutate({
                    moveIncompleteStoriesEnabled: checked,
                  });
                }}
              />
            </Flex>
          ) : null}
        </Box>
      </Box>

      {/* Story Automations Section */}
      <Box className="border-border bg-surface mt-6 rounded-2xl border">
        <SectionHeader
          description={`Automate cleanup and management of ${getTermDisplay("storyTerm", { variant: "plural" })}.`}
          title={`${getTermDisplay("storyTerm", { capitalize: true, variant: "plural" })} automations`}
        />

        <Box className="divide-border divide-y-[0.5px]">
          {/* Auto-close Inactive Stories */}
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

          {/* Inactive Period */}
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
                    autoCloseInactiveMonths: parseInt(value),
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

          {/* Auto-archive Stories */}
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

          {/* Archive Period */}
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
                    autoArchiveMonths: parseInt(value),
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
    </>
  );
};
