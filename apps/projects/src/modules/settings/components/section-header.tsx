import { Box, Flex, Text } from "ui";
import type { ReactNode } from "react";
import { cn } from "lib";

type SectionHeaderProps = {
  title?: string;
  description: string;
  action?: ReactNode;
  className?: string;
};

export const SectionHeader = ({
  title,
  description,
  action,
  className,
}: SectionHeaderProps) => {
  return (
    <Box className={cn("border-border border-b px-6 py-4", className)}>
      <Flex align="center" gap={2} justify="between">
        <Box>
          {title ? (
            <Text as="h3" className="font-medium">
              {title}
            </Text>
          ) : null}
          <Text className="mt-1 line-clamp-2" color="muted">
            {description}
          </Text>
        </Box>
        {action}
      </Flex>
    </Box>
  );
};
