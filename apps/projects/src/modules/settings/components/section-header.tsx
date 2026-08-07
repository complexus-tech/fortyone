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
    <Box className={cn("border-border border-b-[0.5px] px-6 py-4", className)}>
      <Flex align="center" gap={3} justify="between" wrap>
        <Box className="min-w-64 flex-1">
          {title ? (
            <Text as="h3" className="font-medium">
              {title}
            </Text>
          ) : null}
          <Text className="mt-1 line-clamp-2" color="muted">
            {description}
          </Text>
        </Box>
        {action ? <Box className="shrink-0">{action}</Box> : null}
      </Flex>
    </Box>
  );
};
