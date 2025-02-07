"use client";

import { Box, Flex, Text, Wrapper } from "ui";
import type { ReactNode } from "react";

type ActionCardProps = {
  icon: ReactNode;
  title: string;
  description: string;
};

export const ActionCard = ({ icon, title, description }: ActionCardProps) => {
  return (
    <Wrapper>
      <Flex gap={3}>
        <Box className="text-gray-500 dark:text-gray-400 mt-1">{icon}</Box>
        <Box>
          <Text className="font-medium">{title}</Text>
          <Text className="mt-1" color="muted">
            {description}
          </Text>
        </Box>
      </Flex>
    </Wrapper>
  );
};
