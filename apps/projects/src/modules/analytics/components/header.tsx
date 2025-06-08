"use client";
import { Flex, Text } from "ui";

export const Header = () => {
  return (
    <Flex
      align="center"
      className="border-b border-gray-100 bg-gray-50/30 px-4 py-3 dark:border-dark-100 dark:bg-dark-300/50 md:px-5"
      justify="between"
    >
      <Text fontSize="lg" fontWeight="medium">
        Analytics
      </Text>
    </Flex>
  );
};
