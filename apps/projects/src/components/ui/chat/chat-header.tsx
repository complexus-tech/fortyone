import { Text, Flex } from "ui";

export const ChatHeader = () => {
  return (
    <Flex align="center" gap={3}>
      <Text fontSize="lg" fontWeight="medium">
        Maya: Your AI assistant
      </Text>
    </Flex>
  );
};
