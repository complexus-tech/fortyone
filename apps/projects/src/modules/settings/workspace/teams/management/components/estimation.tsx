"use client";

import { Box, Flex, Select, Text } from "ui";
import { useTerminology } from "@/hooks";
import { DEFAULT_ESTIMATE_SCHEME, type EstimateScheme } from "@/lib/estimate";
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
        description={`Choose one relative complexity scale for your team's ${getTermDisplay("storyTerm", { variant: "plural" })}. Time needed is set separately on each item.`}
        title="Complexity"
      />

      <Box className="divide-border divide-y-[0.5px]">
        <Flex align="center" className="gap-4 px-6 py-4" justify="between">
          <Box>
            <Text className="font-medium">Complexity scale</Text>
            <Text className="line-clamp-2 max-w-md" color="muted">
              Keep comparisons consistent with one scale across the team.
            </Text>
          </Box>
          <Select
            onValueChange={(value) => {
              updateEstimationSettings.mutate({
                scheme: value as EstimateScheme,
              });
            }}
            value={estimationSettings?.scheme ?? DEFAULT_ESTIMATE_SCHEME}
          >
            <Select.Trigger className="w-max text-[0.9rem] md:text-base">
              <Select.Input />
            </Select.Trigger>
            <Select.Content>
              <Select.Option className="text-base" value="tshirt">
                T-Shirt — recommended (XS, S, M, L, XL)
              </Select.Option>
              <Select.Option className="text-base" value="points">
                Points (1, 2, 3, 5, 8)
              </Select.Option>
            </Select.Content>
          </Select>
        </Flex>
      </Box>
    </Box>
  );
};
