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
    <Link
      className="group block rounded-xl opacity-80 transition-opacity duration-150 outline-none hover:opacity-100 focus-visible:opacity-100 motion-reduce:transition-none"
      href={href}
    >
      <Wrapper className="group-hover:border-foreground/70 group-hover:ring-foreground/70 group-focus-visible:border-foreground group-focus-visible:ring-foreground border py-4 shadow-none group-hover:ring-1 group-focus-visible:ring-2">
        <Flex align="center" gap={3}>
          <Box className="flex w-10 shrink-0 items-center justify-center">
            {icon}
          </Box>
          <Box className="min-w-0">
            <Text fontWeight="medium">{title}</Text>
            <Text
              className="mt-1 text-[calc(1rem-1px)] leading-5 whitespace-nowrap"
              color="muted"
            >
              {description}
            </Text>
          </Box>
        </Flex>
      </Wrapper>
    </Link>
  );
};
