"use client";

import { Box, Text, Switch } from "ui";
import { SectionHeader } from "@/modules/settings/components";
import { RowWrapper } from "@/components/ui";
import { useUpdateWorkspaceSettingsMutation } from "@/lib/hooks/workspace/update-settings";
import { useFeatures, useTerminology } from "@/hooks";

export const WorkspaceFeatures = () => {
  const { getTermDisplay } = useTerminology();
  const features = useFeatures();
  const { mutate: updateSettings } = useUpdateWorkspaceSettingsMutation();
  const handleToggleFeature = (
    feature: "sprintEnabled" | "objectiveEnabled" | "keyResultEnabled",
  ) => {
    updateSettings({ [feature]: !features[feature] });
  };

  return (
    <Box className="rounded-2xl border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
      <SectionHeader
        description="Enable or disable features for your workspace."
        title="Features"
      />
      <Box>
        <RowWrapper className="md:px-6">
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
            checked={features.objectiveEnabled}
            onCheckedChange={() => {
              handleToggleFeature("objectiveEnabled");
            }}
          />
        </RowWrapper>

        <RowWrapper className="md:px-6">
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
            checked={features.sprintEnabled}
            onCheckedChange={() => {
              handleToggleFeature("sprintEnabled");
            }}
          />
        </RowWrapper>

        <RowWrapper className="border-b-0 md:px-6">
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
            checked={features.keyResultEnabled}
            onCheckedChange={() => {
              handleToggleFeature("keyResultEnabled");
            }}
          />
        </RowWrapper>
      </Box>
    </Box>
  );
};
