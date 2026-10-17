"use client";

import { Box, Text, Switch } from "ui";
import { SectionHeader } from "@/modules/settings/components";
import { RowWrapper } from "@/components/ui";
import { useUpdateWorkspaceSettingsMutation } from "@/lib/hooks/workspace/update-settings";
import { useFeatures, useTerminology } from "@/hooks";

export const WorkspaceFeatures = () => {
  const { getTermDisplay } = useTerminology();
  const features = useFeatures();
  const { mutate: updateSettings, isPending: isSaving } =
    useUpdateWorkspaceSettingsMutation();
  const isDisabled = features.isPending || features.isError || isSaving;
  const handleToggleFeature = (
    feature: "objectiveEnabled" | "keyResultEnabled" | "chatEnabled",
  ) => {
    updateSettings({ [feature]: !features[feature] });
  };

  return (
    <Box className="border-border bg-surface rounded-2xl border">
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
            aria-label={getTermDisplay("objectiveTerm", {
              variant: "plural",
              capitalize: true,
            })}
            checked={features.objectiveEnabled}
            disabled={isDisabled}
            onCheckedChange={() => {
              handleToggleFeature("objectiveEnabled");
            }}
          />
        </RowWrapper>

        <RowWrapper className="md:px-6">
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
            aria-label={getTermDisplay("keyResultTerm", {
              variant: "plural",
              capitalize: true,
            })}
            checked={features.keyResultEnabled}
            disabled={isDisabled}
            onCheckedChange={() => {
              handleToggleFeature("keyResultEnabled");
            }}
          />
        </RowWrapper>

        <RowWrapper className="border-b-0 md:px-6">
          <Box>
            <Text className="font-medium">Chat</Text>
            <Text color="muted">
              Bring team conversations and direct messages into your workspace
            </Text>
          </Box>
          <Switch
            aria-label="Chat"
            checked={features.chatEnabled}
            disabled={isDisabled}
            onCheckedChange={() => {
              handleToggleFeature("chatEnabled");
            }}
          />
        </RowWrapper>
      </Box>
    </Box>
  );
};
