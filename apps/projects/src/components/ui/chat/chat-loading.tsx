import { AiIcon } from "icons";
import { Avatar, Box, Flex } from "ui";

export const ChatLoading = () => {
  return (
    <Flex gap={3}>
      <Avatar
        className="bg-gradient-to-br from-primary to-secondary"
        color="primary"
        size="md"
      >
        <AiIcon />
      </Avatar>
      <Box className="rounded-2xl bg-gray-50 p-4 dark:bg-dark-100">
        <Flex gap={1}>
          <Box className="size-2 animate-bounce rounded-full bg-gray" />
          <Box
            className="size-2 animate-bounce rounded-full bg-gray"
            style={{ animationDelay: "0.1s" }}
          />
          <Box
            className="size-2 animate-bounce rounded-full bg-gray"
            style={{ animationDelay: "0.2s" }}
          />
        </Flex>
      </Box>
    </Flex>
  );
};
