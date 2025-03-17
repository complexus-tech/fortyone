import { Box, Flex, Text, Switch } from "ui";
import { SectionHeader } from "@/modules/settings/components";
import { useTerminology } from "@/hooks";

export const Automations = () => {
  const { getTermDisplay } = useTerminology();
  return (
    <Box className="mt-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
      <SectionHeader
        description={`Configure how ${getTermDisplay("storyTerm", { variant: "plural" })} are automatically handled.`}
        title="Automations"
      />
      <Box className="p-6">
        <Flex direction="column" gap={6}>
          <Flex align="center" justify="between">
            <Box>
              <Text className="font-medium">Auto-assign to self</Text>
              <Text color="muted">
                When creating new{" "}
                {getTermDisplay("storyTerm", { variant: "plural" })}, always
                assign them to yourself by default
              </Text>
            </Box>
            <Switch name="autoAssignSelf" />
          </Flex>

          <Flex align="center" justify="between">
            <Box>
              <Text className="font-medium">
                On git branch copy, move {getTermDisplay("storyTerm")} to
                started status
              </Text>
              <Text color="muted">
                After copying the git branch name, {getTermDisplay("storyTerm")}{" "}
                is moved to the started workflow status
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
                After copying the git branch name, {getTermDisplay("storyTerm")}{" "}
                is assigned to yourself
              </Text>
            </Box>
            <Switch name="autoBranchAssign" />
          </Flex>
        </Flex>
      </Box>
    </Box>
  );
};
