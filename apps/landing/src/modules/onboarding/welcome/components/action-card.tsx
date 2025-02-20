import { Box, Flex, Text, Wrapper } from "ui";
import type { ReactNode } from "react";

type ActionCardProps = {
  icon: ReactNode;
  title: string;
  description: ReactNode;
};

export const ActionCard = ({ icon, title, description }: ActionCardProps) => {
  return (
    <Wrapper className="py-2.5">
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
  );
};
