"use client";

import { Box, Flex, Select, Text } from "ui";
import { useTerminology } from "@/hooks";
import { DEFAULT_ESTIMATE_SCHEME } from "@/lib/estimate";
import { SectionHeader } from "@/modules/settings/components/section-header";
import { useTeamSettings } from "@/modules/teams/hooks/use-team-settings";
import { useUpdateEstimationSettingsMutation } from "@/modules/teams/hooks/update-estimation-settings-mutation";

export const EstimationSettings = ({ teamId }: { teamId: string }) => {
  const { getTermDisplay } = useTerminology();
  const { data: teamSettings } = useTeamSettings(teamId);
  const updateEstimationSettings = useUpdateEstimationSettingsMutation(teamId);
  const estimationSettings = teamSettings?.estimationSettings;

  return (
    <Box className="border-border bg-surface mt-6 rounded-2xl border">
      <SectionHeader
        description={`Choose how your team estimates ${getTermDisplay("storyTerm", { variant: "plural" })}.`}
        title="Estimation"
      />

      <Box className="divide-border divide-y-[0.5px]">
        <Flex align="center" className="gap-4 px-6 py-4" justify="between">
          <Box>
            <Text className="font-medium">Estimation scheme</Text>
            <Text className="line-clamp-2 max-w-md" color="muted">
              Team-wide estimation values are restricted to the selected scheme
            </Text>
          </Box>
          <Select
            onValueChange={(value) => {
              updateEstimationSettings.mutate({
                scheme: value as "points" | "hours" | "tshirt" | "ideal_days",
              });
            }}
            value={estimationSettings?.scheme ?? DEFAULT_ESTIMATE_SCHEME}
          >
            <Select.Trigger className="w-max text-[0.9rem] md:text-base">
              <Select.Input />
            </Select.Trigger>
            <Select.Content>
              <Select.Option className="text-base" value="points">
                Points (1,2,3,5,8)
              </Select.Option>
              <Select.Option className="text-base" value="hours">
                Hours (0.5,1,2,4,8)
              </Select.Option>
              <Select.Option className="text-base" value="tshirt">
                T-Shirt (XS,S,M,L,XL)
              </Select.Option>
              <Select.Option className="text-base" value="ideal_days">
                Ideal Days (0.5,1,2,3,5)
              </Select.Option>
            </Select.Content>
          </Select>
        </Flex>
      </Box>
    </Box>
  );
};
