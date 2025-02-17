import { Box, Flex, Text, Wrapper } from "ui";
import type { ReactNode } from "react";

type ActionCardProps = {
  icon: ReactNode;
  title: string;
  description: ReactNode;
};

export const ActionCard = ({ icon, title, description }: ActionCardProps) => {
  return (
    <Wrapper className="py-3">
      <Flex gap={3}>
        <Box className="mt-1">{icon}</Box>
        <Box>
          <Text className="font-semibold">{title}</Text>
          <Text className="mt-1 text-[0.95rem]" color="muted">
            {description}
          </Text>
        </Box>
      </Flex>
    </Wrapper>
  );
};
