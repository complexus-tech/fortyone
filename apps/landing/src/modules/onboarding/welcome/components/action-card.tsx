import { Box, Flex, Text, Wrapper } from "ui";
import type { ReactNode } from "react";
import Link from "next/link";

type ActionCardProps = {
  icon: ReactNode;
  title: string;
  description: ReactNode;
  href: string;
};

export const ActionCard = ({
  icon,
  title,
  description,
  href,
}: ActionCardProps) => {
  return (
    <Link href={href}>
      <Wrapper className="border-opacity-40 py-2.5 shadow-none dark:border-opacity-40">
        <Flex gap={3}>
          <Box className="mt-1">{icon}</Box>
          <Box>
            <Text>{title}</Text>
            <Text className="mt-1 text-[0.9rem]" color="muted">
              {description}
            </Text>
          </Box>
        </Flex>
      </Wrapper>
    </Link>
  );
};
