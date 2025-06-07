"use client";

import { useState } from "react";
import { Box, Text, Switch, Select, Flex } from "ui";
import { SectionHeader } from "@/modules/settings/components/section-header";
import { useTerminology } from "@/hooks";

export const Automations = () => {
  const { getTermDisplay } = useTerminology();

  // State for all automation settings
  const [settings, setSettings] = useState({
    sprintsEnabled: false,
    numberOfSprints: "2",
    sprintLength: "2weeks",
    startsOn: "monday",
    autoMoveIncompleteStories: false,
    autoCloseInactiveStories: false,
    inactivePeriod: "6months",
    autoArchiveStories: false,
    archivePeriod: "6months",
  });

  return (
    <>
      {/* Sprint Configuration Section */}
      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description={`Configure ${getTermDisplay("sprintTerm", { variant: "plural" })} and automation settings for your team.`}
          title={getTermDisplay("sprintTerm", {
            capitalize: true,
            variant: "plural",
          })}
        />

        <Box className="divide-y-[0.5px] divide-gray-100 dark:divide-dark-100">
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
              checked={settings.sprintsEnabled}
              onCheckedChange={(checked) => {
                setSettings((prev) => ({ ...prev, sprintsEnabled: checked }));
              }}
            />
          </Flex>

          {/* Number of Sprints to Create */}
          {settings.sprintsEnabled ? (
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
                  setSettings((prev) => ({ ...prev, numberOfSprints: value }));
                }}
                value={settings.numberOfSprints}
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
          {settings.sprintsEnabled ? (
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
                  setSettings((prev) => ({ ...prev, sprintLength: value }));
                }}
                value={settings.sprintLength}
              >
                <Select.Trigger className="w-32 text-[0.9rem] md:text-base">
                  <Select.Input />
                </Select.Trigger>
                <Select.Content>
                  <Select.Option className="text-base" value="1week">
                    1 week
                  </Select.Option>
                  <Select.Option className="text-base" value="2weeks">
                    2 weeks
                  </Select.Option>
                  <Select.Option className="text-base" value="3weeks">
                    3 weeks
                  </Select.Option>
                  <Select.Option className="text-base" value="4weeks">
                    4 weeks
                  </Select.Option>
                </Select.Content>
              </Select>
            </Flex>
          ) : null}

          {/* Sprints Start On */}
          {settings.sprintsEnabled ? (
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
                  setSettings((prev) => ({ ...prev, startsOn: value }));
                }}
                value={settings.startsOn}
              >
                <Select.Trigger className="w-32 text-[0.9rem] md:text-base">
                  <Select.Input />
                </Select.Trigger>
                <Select.Content>
                  <Select.Option className="text-base" value="monday">
                    Monday
                  </Select.Option>
                  <Select.Option className="text-base" value="tuesday">
                    Tuesday
                  </Select.Option>
                  <Select.Option className="text-base" value="wednesday">
                    Wednesday
                  </Select.Option>
                  <Select.Option className="text-base" value="thursday">
                    Thursday
                  </Select.Option>
                  <Select.Option className="text-base" value="friday">
                    Friday
                  </Select.Option>
                </Select.Content>
              </Select>
            </Flex>
          ) : null}

          {/* Move Incomplete Stories */}
          {settings.sprintsEnabled ? (
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
                checked={settings.autoMoveIncompleteStories}
                onCheckedChange={(checked) => {
                  setSettings((prev) => ({
                    ...prev,
                    autoMoveIncompleteStories: checked,
                  }));
                }}
              />
            </Flex>
          ) : null}
        </Box>
      </Box>

      {/* Story Automations Section */}
      <Box className="mt-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description={`Automate cleanup and management of ${getTermDisplay("storyTerm", { variant: "plural" })}.`}
          title={`${getTermDisplay("storyTerm", { capitalize: true, variant: "plural" })} automations`}
        />

        <Box className="divide-y-[0.5px] divide-gray-100 dark:divide-dark-100">
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
              checked={settings.autoCloseInactiveStories}
              onCheckedChange={(checked) => {
                setSettings((prev) => ({
                  ...prev,
                  autoCloseInactiveStories: checked,
                }));
              }}
            />
          </Flex>

          {/* Inactive Period */}
          {settings.autoCloseInactiveStories ? (
            <Flex align="center" className="px-6 py-4" justify="between">
              <Box>
                <Text className="font-medium">
                  Close after being inactive for
                </Text>
              </Box>
              <Select
                onValueChange={(value) => {
                  setSettings((prev) => ({ ...prev, inactivePeriod: value }));
                }}
                value={settings.inactivePeriod}
              >
                <Select.Trigger className="w-32 text-[0.9rem] md:text-base">
                  <Select.Input />
                </Select.Trigger>
                <Select.Content>
                  <Select.Option className="text-base" value="3months">
                    3 months
                  </Select.Option>
                  <Select.Option className="text-base" value="6months">
                    6 months
                  </Select.Option>
                  <Select.Option className="text-base" value="1year">
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
                Automatically archive completed{" "}
                {getTermDisplay("storyTerm", { variant: "plural" })} after a
                period of time
              </Text>
            </Box>
            <Switch
              checked={settings.autoArchiveStories}
              onCheckedChange={(checked) => {
                setSettings((prev) => ({
                  ...prev,
                  autoArchiveStories: checked,
                }));
              }}
            />
          </Flex>

          {/* Archive Period */}
          {settings.autoArchiveStories ? (
            <Flex align="center" className="px-6 py-4" justify="between">
              <Box>
                <Text className="font-medium">Archive after</Text>
              </Box>
              <Select
                onValueChange={(value) => {
                  setSettings((prev) => ({ ...prev, archivePeriod: value }));
                }}
                value={settings.archivePeriod}
              >
                <Select.Trigger className="w-32 text-[0.9rem] md:text-base">
                  <Select.Input />
                </Select.Trigger>
                <Select.Content>
                  <Select.Option className="text-base" value="3months">
                    3 months
                  </Select.Option>
                  <Select.Option className="text-base" value="6months">
                    6 months
                  </Select.Option>
                  <Select.Option className="text-base" value="1year">
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
