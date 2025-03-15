import { useState } from "react";
import { Box, Text, Switch } from "ui";
import { SectionHeader } from "@/modules/settings/components";
import { RowWrapper } from "@/components/ui";

export const WorkspaceFeatures = () => {
  const [features, setFeatures] = useState({
    objectives: true,
    sprints: true,
  });

  const handleToggleFeature = (feature: keyof typeof features) => {
    setFeatures((prev) => ({
      ...prev,
      [feature]: !prev[feature],
    }));
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
            <Text className="font-medium">Objectives</Text>
            <Text color="muted">
              Track high-level goals and connect work to desired outcomes
            </Text>
          </Box>
          <Switch
            checked={features.objectives}
            onCheckedChange={() => {
              handleToggleFeature("objectives");
            }}
          />
        </RowWrapper>

        <RowWrapper className="px-6">
          <Box>
            <Text className="font-medium">Sprints</Text>
            <Text color="muted">
              Plan and organize work into time-boxed periods
            </Text>
          </Box>
          <Switch
            checked={features.sprints}
            onCheckedChange={() => {
              handleToggleFeature("sprints");
            }}
          />
        </RowWrapper>
      </Box>
    </Box>
  );
};
