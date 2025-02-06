import { Box, Flex, Text, Switch } from "ui";
import { SectionHeader } from "@/modules/settings/components";

export const Automations = () => {
  return (
    <Box className="mt-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
      <SectionHeader
        description="Configure how stories are automatically handled."
        title="Automation"
      />
      <Box className="p-6">
        <Flex direction="column" gap={6}>
          <Flex align="center" justify="between">
            <Box>
              <Text className="font-medium">Auto-assign to self</Text>
              <Text color="muted">
                When creating new stories, always assign them to yourself by
                default
              </Text>
            </Box>
            <Switch name="autoAssignSelf" />
          </Flex>

          <Flex align="center" justify="between">
            <Box>
              <Text className="font-medium">
                On git branch copy, move story to started status
              </Text>
              <Text color="muted">
                After copying the git branch name, story is moved to the started
                workflow status
              </Text>
            </Box>
            <Switch name="autoBranchMoveStatus" />
          </Flex>

          <Flex align="center" justify="between">
            <Box>
              <Text className="font-medium">
                On git branch copy, assign to yourself
              </Text>
              <Text color="muted">
                After copying the git branch name, story is assigned to yourself
              </Text>
            </Box>
            <Switch name="autoBranchAssign" />
          </Flex>
        </Flex>
      </Box>
    </Box>
  );
};
