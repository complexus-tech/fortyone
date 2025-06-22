import { Box, Flex, Text, Switch } from "ui";
import { SectionHeader } from "@/modules/settings/components";
import { useTerminology } from "@/hooks";
import { useAutomationPreferences } from "@/lib/hooks/users/preferences";
import { useUpdateAutomationPreferencesMutation } from "@/lib/hooks/users/update-auto-preferences";

export const Automations = () => {
  const { data: preferences } = useAutomationPreferences();
  const { getTermDisplay } = useTerminology();
  const { mutate: updatePreferences } =
    useUpdateAutomationPreferencesMutation();

  const handleToggle = (field: string, checked: boolean) => {
    updatePreferences({ [field]: checked });
  };

  return (
    <Box className="mt-6 rounded-2xl border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
      <SectionHeader
        description={`Configure how ${getTermDisplay("storyTerm", { variant: "plural" })} are automatically handled.`}
        title="Automations"
      />
      <Box className="p-6">
        <Flex direction="column" gap={6}>
          <Flex align="center" gap={2} justify="between">
            <Box>
              <Text className="font-medium">Auto-assign to self</Text>
              <Text className="line-clamp-2" color="muted">
                When creating new{" "}
                {getTermDisplay("storyTerm", { variant: "plural" })}, always
                assign them to yourself by default
              </Text>
            </Box>
            <Switch
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
              <Text className="line-clamp-1 font-medium">
                On git branch copy, move {getTermDisplay("storyTerm")} to
                started status
              </Text>
              <Text className="line-clamp-2" color="muted">
                After copying the git branch name, {getTermDisplay("storyTerm")}{" "}
                is moved to the started workflow status
              </Text>
            </Box>
            <Switch
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
              <Text className="font-medium">
                On git branch copy, assign to yourself
              </Text>
              <Text className="line-clamp-2" color="muted">
                After copying the git branch name, {getTermDisplay("storyTerm")}{" "}
                is assigned to yourself
              </Text>
            </Box>
            <Switch
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
