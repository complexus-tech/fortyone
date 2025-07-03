import React from "react";
import { Box, Flex, Text } from "ui";

export const Thinking = ({
  message = "Maya is thinking",
}: {
  message?: string;
}) => {
  return (
    <Text as="div" className="flex items-baseline gap-0.5" color="muted">
      {message}
      <Flex className="gap-0.5">
        <Box className="size-[2.5px] animate-bounce rounded-full bg-current" />
        <Box
          className="size-[2.5px] animate-bounce rounded-full bg-current"
          style={{ animationDelay: "0.1s" }}
        />
        <Box
          className="size-[2.5px] animate-bounce rounded-full bg-current"
          style={{ animationDelay: "0.2s" }}
        />
      </Flex>
    </Text>
  );
};
