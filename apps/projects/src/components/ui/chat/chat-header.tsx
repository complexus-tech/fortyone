import { AiIcon } from "icons";
import { Avatar, Text, Flex } from "ui";

export const ChatHeader = () => {
  return (
    <Flex align="center" gap={3}>
      <Avatar
        className="bg-gradient-to-br from-primary to-secondary"
        color="primary"
        size="sm"
      >
        <AiIcon />
      </Avatar>
      <Text fontSize="lg" fontWeight="medium">
        Maya AI Assistant
      </Text>
    </Flex>
  );
};
