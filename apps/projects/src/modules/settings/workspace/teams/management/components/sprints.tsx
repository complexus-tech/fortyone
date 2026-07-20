"use client";

import { useParams } from "next/navigation";
import { useState } from "react";
import { isAfter, startOfDay, startOfToday } from "date-fns";
import { Box, Text, Switch, Select, Flex, Input, Button, Dialog } from "ui";
import { SectionHeader } from "@/modules/settings/components/section-header";
import { useTerminology } from "@/hooks";
import { useTeamSettings } from "@/modules/teams/hooks/use-team-settings";
import { useUpdateSprintSettingsMutation } from "@/modules/teams/hooks/update-sprint-settings-mutation";
import { useTeamSprints } from "@/modules/sprints/hooks/team-sprints";
import type { UpdateSprintSettingsInput } from "@/modules/teams/types";
import { WorkingDaysSetting } from "./working-days-setting";

export const SprintSettings = () => {
  const { teamId } = useParams<{ teamId: string }>();
  const { getTermDisplay } = useTerminology();
  const { data: teamSettings } = useTeamSettings(teamId);
  const { data: sprints = [] } = useTeamSprints(teamId);
  const updateSprintSettings = useUpdateSprintSettingsMutation(teamId);
  const sprintSettings = teamSettings?.sprintSettings;
  const [isNextSprintDialogOpen, setIsNextSprintDialogOpen] = useState(false);
  const [nextSprintNumber, setNextSprintNumber] = useState("");
  const [pendingCadenceUpdate, setPendingCadenceUpdate] =
    useState<UpdateSprintSettingsInput | null>(null);
  const currentNextSprintNumber = sprintSettings?.nextAutoSprintNumber ?? 1;
  const upcomingManagedSprints = sprints.filter(
    (sprint) =>
      sprint.scheduleManagedByAutomation &&
      isAfter(startOfDay(new Date(sprint.startDate)), startOfToday()),
  );
  const assignedStoryCount = upcomingManagedSprints.reduce(
    (total, sprint) => total + sprint.stats.total,
    0,
  );

  const handleNextSprintDialogOpenChange = (open: boolean) => {
    if (open) {
      setNextSprintNumber(currentNextSprintNumber.toString());
    }
    setIsNextSprintDialogOpen(open);
  };

  const handleUpdateNextSprintNumber = () => {
    const nextValue = Number.parseInt(nextSprintNumber, 10);
    if (
      Number.isNaN(nextValue) ||
      nextValue < 1 ||
      nextValue > 10000 ||
      nextValue === currentNextSprintNumber
    ) {
      setIsNextSprintDialogOpen(false);
      setNextSprintNumber(currentNextSprintNumber.toString());
      return;
    }

    updateSprintSettings.mutate({ nextAutoSprintNumber: nextValue });
    setIsNextSprintDialogOpen(false);
  };

  const handleConfirmCadenceUpdate = () => {
    if (!pendingCadenceUpdate) return;

    updateSprintSettings.mutate(pendingCadenceUpdate, {
      onSuccess: () => {
        setPendingCadenceUpdate(null);
      },
    });
  };

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

          {!sprintSettings?.autoCreateSprints &&
          sprintSettings?.autoCreateDisabledReason ? (
            <Box className="px-6 py-4">
              <Text className="font-medium">Automation paused</Text>
              <Text className="line-clamp-2" color="muted">
                {sprintSettings.autoCreateDisabledReason}
              </Text>
            </Box>
          ) : null}

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

          {/* Next Auto Sprint Number */}
          {sprintSettings?.autoCreateSprints ? (
            <Flex align="center" className="px-6 py-4" justify="between">
              <Box>
                <Text className="font-medium">
                  Next {getTermDisplay("sprintTerm")} number
                </Text>
                <Text className="line-clamp-2" color="muted">
                  The next automated {getTermDisplay("sprintTerm")} will be
                  named {getTermDisplay("sprintTerm", { capitalize: true })}{" "}
                  {sprintSettings.nextAutoSprintNumber}
                </Text>
              </Box>
              <Button
                className="dark:bg-surface-elevated"
                color="tertiary"
                onClick={() => {
                  handleNextSprintDialogOpenChange(true);
                }}
                size="sm"
                variant="outline"
              >
                {getTermDisplay("sprintTerm", { capitalize: true })}{" "}
                {sprintSettings.nextAutoSprintNumber}
              </Button>
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
                  setPendingCadenceUpdate({
                    sprintDurationWeeks: Number.parseInt(value, 10),
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
                  setPendingCadenceUpdate({
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

          <WorkingDaysSetting
            isPending={updateSprintSettings.isPending}
            onSave={(workingDays, onSuccess) => {
              updateSprintSettings.mutate({ workingDays }, { onSuccess });
            }}
            sprintTerm={getTermDisplay("sprintTerm")}
            value={sprintSettings?.workingDays}
          />

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

      <Dialog
        onOpenChange={handleNextSprintDialogOpenChange}
        open={isNextSprintDialogOpen}
      >
        <Dialog.Content size="sm">
          <Dialog.Header className="px-6 pt-6 pb-2">
            <Dialog.Title className="text-lg">
              Next {getTermDisplay("sprintTerm")} number
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="px-6 pt-2 pb-4">
            <Input
              autoFocus
              label={`Next ${getTermDisplay("sprintTerm")} number`}
              max={10000}
              min={1}
              onChange={(event) => {
                setNextSprintNumber(event.target.value);
              }}
              onKeyDown={(event) => {
                if (event.key === "Enter") {
                  handleUpdateNextSprintNumber();
                }
              }}
              type="number"
              value={nextSprintNumber}
            />
          </Dialog.Body>
          <Dialog.Footer className="justify-end gap-2">
            <Button
              color="tertiary"
              onClick={() => {
                setIsNextSprintDialogOpen(false);
              }}
              size="sm"
              variant="outline"
            >
              Cancel
            </Button>
            <Button onClick={handleUpdateNextSprintNumber} size="sm">
              Update
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>

      <Dialog
        onOpenChange={(open) => {
          if (!open && !updateSprintSettings.isPending) {
            setPendingCadenceUpdate(null);
          }
        }}
        open={pendingCadenceUpdate !== null}
      >
        <Dialog.Content size="sm">
          <Dialog.Header className="px-6 pt-6 pb-2">
            <Dialog.Title className="text-lg">
              {pendingCadenceUpdate?.sprintDurationWeeks
                ? `Change ${getTermDisplay("sprintTerm")} length?`
                : `Change ${getTermDisplay("sprintTerm")} start day?`}
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="space-y-3 px-6 pt-2 pb-4">
            <Text>
              Your current {getTermDisplay("sprintTerm")} will keep its existing
              dates. The new cadence begins with the next{" "}
              {getTermDisplay("sprintTerm")}.
            </Text>
            <Text color="muted">
              {upcomingManagedSprints.length > 0
                ? `${upcomingManagedSprints.length} upcoming automated ${getTermDisplay("sprintTerm", { variant: upcomingManagedSprints.length === 1 ? "singular" : "plural" })} will be rescheduled in place.`
                : `No existing ${getTermDisplay("sprintTerm", { variant: "plural" })} need to be rescheduled.`}{" "}
              {assignedStoryCount > 0
                ? `${assignedStoryCount} assigned ${getTermDisplay("storyTerm", { variant: assignedStoryCount === 1 ? "singular" : "plural" })} will remain assigned.`
                : "Assignments will not be changed."}{" "}
              {getTermDisplay("sprintTerm", {
                capitalize: true,
                variant: "plural",
              })}{" "}
              with custom dates will not be changed.
            </Text>
          </Dialog.Body>
          <Dialog.Footer className="justify-end gap-2">
            <Button
              color="tertiary"
              disabled={updateSprintSettings.isPending}
              onClick={() => {
                setPendingCadenceUpdate(null);
              }}
              size="sm"
              variant="outline"
            >
              Cancel
            </Button>
            <Button
              loading={updateSprintSettings.isPending}
              onClick={handleConfirmCadenceUpdate}
              size="sm"
            >
              Update schedule
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>
    </>
  );
};
