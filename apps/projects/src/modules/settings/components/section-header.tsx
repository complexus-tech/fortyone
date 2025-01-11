import { Box, Flex, Text } from "ui";
import type { ReactNode } from "react";

type SectionHeaderProps = {
  title: string;
  description: string;
  action?: ReactNode;
};

export const SectionHeader = ({
  title,
  description,
  action,
}: SectionHeaderProps) => {
  return (
    <Box className="border-b border-gray-100 px-6 py-4 dark:border-dark-100">
      <Flex align="center" justify="between">
        <Box>
          <Text as="h3" className="font-medium">
            {title}
          </Text>
          <Text className="mt-1" color="muted">
            {description}
          </Text>
        </Box>
        {action}
      </Flex>
    </Box>
  );
};
