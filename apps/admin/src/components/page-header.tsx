import type { ReactNode } from "react";
import { Box, Flex, Text } from "ui";

export const PageHeader = ({
  actions,
  description,
  eyebrow,
  title,
}: {
  actions?: ReactNode;
  description?: string;
  eyebrow?: string;
  title: string;
}) => {
  return (
    <Flex
      align="start"
      className="border-border/80 gap-4 border-b-[0.5px] px-5 py-5 md:px-7"
      justify="between"
    >
      <Box className="min-w-0">
        {eyebrow ? (
          <Text
            className="mb-2 text-[0.82rem] tracking-[0.08em]"
            color="muted"
            transform="uppercase"
          >
            {eyebrow}
          </Text>
        ) : null}
        <Text
          as="h1"
          className="font-heading text-[1.7rem] leading-tight"
          fontWeight="semibold"
        >
          {title}
        </Text>
        {description ? (
          <Text className="mt-1 max-w-3xl text-[0.98rem]" color="muted">
            {description}
          </Text>
        ) : null}
      </Box>
      {actions ? <Box className="shrink-0">{actions}</Box> : null}
    </Flex>
  );
};
