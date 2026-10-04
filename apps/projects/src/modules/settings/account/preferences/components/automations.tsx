import { useId } from "react";
import { Box, Flex, Text, Switch } from "ui";
import { SectionHeader } from "@/modules/settings/components";
import { useTerminology } from "@/hooks";
import { useAutomationPreferences } from "@/lib/hooks/users/preferences";
import { useUpdateAutomationPreferencesMutation } from "@/lib/hooks/users/update-auto-preferences";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";

export const Automations = () => {
  const id = useId();
  const { data: preferences } = useAutomationPreferences();
  const { getTermDisplay } = useTerminology();
  const { hasFeature } = useSubscriptionFeatures();
  const { mutate: updatePreferences } =
    useUpdateAutomationPreferencesMutation();
  const canUseBackgroundMaya = hasFeature("backgroundMaya");

  const handleToggle = (
    field:
      | "autoAssignSelf"
      | "autoScheduling"
      | "assignSelfOnBranchCopy"
      | "moveStoryToStartedOnBranch",
    checked: boolean,
  ) => {
    updatePreferences({ [field]: checked });
  };

  return (
    <Box className="border-border bg-surface mt-6 rounded-2xl border">
      <SectionHeader
        description={`Configure how ${getTermDisplay("storyTerm", { variant: "plural" })} are automatically handled.`}
        title="Automations"
      />
      <Box className="p-6">
        <Flex direction="column" gap={6}>
          <Flex align="center" gap={2} justify="between">
            <Box>
              <Text className="font-medium" id={`${id}-auto-assign-label`}>
                Auto-assign to self
              </Text>
              <Text
                className="line-clamp-2"
                color="muted"
                id={`${id}-auto-assign-description`}
              >
                When creating new{" "}
                {getTermDisplay("storyTerm", { variant: "plural" })}, always
                assign them to yourself by default
              </Text>
            </Box>
            <Switch
              aria-describedby={`${id}-auto-assign-description`}
              aria-labelledby={`${id}-auto-assign-label`}
              checked={preferences?.autoAssignSelf}
              className="shrink-0"
              name="autoAssignSelf"
              onCheckedChange={(checked) => {
                handleToggle("autoAssignSelf", checked);
              }}
            />
          </Flex>

          <Flex align="center" gap={2} justify="between">
            <Box>
              <Text className="font-medium" id={`${id}-auto-scheduling-label`}>
                Auto-scheduling
              </Text>
              <Text
                className="line-clamp-2"
                color="muted"
                id={`${id}-auto-scheduling-description`}
              >
                When creating new{" "}
                {getTermDisplay("storyTerm", { variant: "plural" })}, enable
                auto-scheduling by default. You can turn it off on individual{" "}
                {getTermDisplay("storyTerm", { variant: "plural" })}.
              </Text>
              {!canUseBackgroundMaya && (
                <Text
                  className="mt-1"
                  color="muted"
                  id={`${id}-auto-scheduling-availability`}
                >
                  Available on paid plans.
                </Text>
              )}
            </Box>
            <Switch
              aria-describedby={
                canUseBackgroundMaya
                  ? `${id}-auto-scheduling-description`
                  : `${id}-auto-scheduling-description ${id}-auto-scheduling-availability`
              }
              aria-labelledby={`${id}-auto-scheduling-label`}
              checked={Boolean(
                canUseBackgroundMaya && preferences?.autoScheduling,
              )}
              className="shrink-0"
              disabled={!canUseBackgroundMaya}
              name="autoScheduling"
              onCheckedChange={(checked) => {
                handleToggle("autoScheduling", checked);
              }}
            />
          </Flex>

          <Flex align="center" gap={2} justify="between">
            <Box>
              <Text
                className="line-clamp-1 font-medium"
                id={`${id}-branch-status-label`}
              >
                On git branch copy, move {getTermDisplay("storyTerm")} to
                started status
              </Text>
              <Text
                className="line-clamp-2"
                color="muted"
                id={`${id}-branch-status-description`}
              >
                After copying the git branch name, {getTermDisplay("storyTerm")}{" "}
                is moved to the started workflow status
              </Text>
            </Box>
            <Switch
              aria-describedby={`${id}-branch-status-description`}
              aria-labelledby={`${id}-branch-status-label`}
              checked={preferences?.moveStoryToStartedOnBranch}
              className="shrink-0"
              name="autoBranchMoveStatus"
              onCheckedChange={(checked) => {
                handleToggle("moveStoryToStartedOnBranch", checked);
              }}
            />
          </Flex>

          <Flex align="center" gap={2} justify="between">
            <Box>
              <Text className="font-medium" id={`${id}-branch-assign-label`}>
                On git branch copy, assign to yourself
              </Text>
              <Text
                className="line-clamp-2"
                color="muted"
                id={`${id}-branch-assign-description`}
              >
                After copying the git branch name, {getTermDisplay("storyTerm")}{" "}
                is assigned to yourself
              </Text>
            </Box>
            <Switch
              aria-describedby={`${id}-branch-assign-description`}
              aria-labelledby={`${id}-branch-assign-label`}
              checked={preferences?.assignSelfOnBranchCopy}
              className="shrink-0"
              name="autoBranchAssign"
              onCheckedChange={(checked) => {
                handleToggle("assignSelfOnBranchCopy", checked);
              }}
            />
          </Flex>
        </Flex>
      </Box>
    </Box>
  );
};
