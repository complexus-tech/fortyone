"use client";

import { Box, Text, Switch } from "ui";
import { SectionHeader } from "@/modules/settings/components";
import { RowWrapper } from "@/components/ui";
import { useWorkspaceSettings } from "@/lib/hooks/workspace/settings";
import { useUpdateWorkspaceSettingsMutation } from "@/lib/hooks/workspace/update-settings";
import { useTerminology } from "@/hooks";

export const WorkspaceFeatures = () => {
  const { getTermDisplay } = useTerminology();
  const {
    data: settings = {
      sprintEnabled: true,
      objectiveEnabled: true,
      keyResultEnabled: true,
    },
  } = useWorkspaceSettings();
  const { mutate: updateSettings } = useUpdateWorkspaceSettingsMutation();
  const handleToggleFeature = (
    feature: "sprintEnabled" | "objectiveEnabled" | "keyResultEnabled",
  ) => {
    updateSettings({ [feature]: !settings[feature] });
  };

  return (
    <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
      <SectionHeader
        description="Enable or disable features for your workspace."
        title="Features"
      />
      <Box>
        <RowWrapper className="px-6">
          <Box>
            <Text className="font-medium">
              {getTermDisplay("objectiveTerm", {
                variant: "plural",
                capitalize: true,
              })}
            </Text>
            <Text color="muted">
              Track high-level goals and connect work to desired outcomes
            </Text>
          </Box>
          <Switch
            checked={settings.objectiveEnabled}
            onCheckedChange={() => {
              handleToggleFeature("objectiveEnabled");
            }}
          />
        </RowWrapper>

        <RowWrapper className="px-6">
          <Box>
            <Text className="font-medium">
              {getTermDisplay("sprintTerm", {
                variant: "plural",
                capitalize: true,
              })}
            </Text>
            <Text color="muted">
              Plan and organize work into time-boxed periods
            </Text>
          </Box>
          <Switch
            checked={settings.sprintEnabled}
            onCheckedChange={() => {
              handleToggleFeature("sprintEnabled");
            }}
          />
        </RowWrapper>

        <RowWrapper className="border-b-0 px-6">
          <Box>
            <Text className="font-medium">
              {getTermDisplay("keyResultTerm", {
                variant: "plural",
                capitalize: true,
              })}
            </Text>
            <Text color="muted">
              Measure progress towards key{" "}
              {getTermDisplay("objectiveTerm", {
                variant: "plural",
              })}
            </Text>
          </Box>
          <Switch
            checked={settings.keyResultEnabled}
            onCheckedChange={() => {
              handleToggleFeature("keyResultEnabled");
            }}
          />
        </RowWrapper>
      </Box>
    </Box>
  );
};
