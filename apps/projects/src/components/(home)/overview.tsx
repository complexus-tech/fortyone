"use client";
import { Box, Text, Flex, Wrapper } from "ui";
import { ArrowUpRight } from "lucide-react";

type CardProps = {
  title: string;
  count: number;
  description: string;
};
const Card = ({ title, count, description }: CardProps) => (
  <Wrapper>
    <Flex align="center" justify="between">
      <Text className="mb-2" color="muted">
        {title}
      </Text>
      <ArrowUpRight className="h-5 w-auto text-primary" />
    </Flex>
    <Text className="mb-2" fontSize="2xl" fontWeight="semibold">
      {count}
    </Text>
    <Text color="muted">{description}</Text>
  </Wrapper>
);

export const Overview = () => {
  return (
    <Box>
      <Text as="h2" fontSize="3xl" fontWeight="medium">
        Good afternoon, Joseph.
      </Text>
      <Text color="muted">
        Here&rsquo;s what&rsquo;s happening with your projects today.
      </Text>

      <Box className="my-4 grid grid-cols-5 gap-4">
        {new Array(5).fill(0).map((_, i) => (
          <Card
            count={27}
            description="+5% from last month"
            key={i}
            title="Issues closed"
          />
        ))}
      </Box>
    </Box>
  );
};
