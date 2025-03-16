import { Box, Flex, Text, Switch } from "ui";

export const Entry = () => {
  return (
    <Flex align="center" justify="between">
      <Box>
        <Text className="font-medium">Story updates</Text>
        <Text color="muted">
          Get notified when a story you&apos;re involved with is updated
        </Text>
      </Box>
      <Switch checked />
    </Flex>
  );
};
