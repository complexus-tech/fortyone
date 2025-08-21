import type { ReactNode } from "react";
import { Text, Box, Flex } from "ui";
import { Container } from "@/components/ui";

type FeatureCard = {
  icon: ReactNode;
  title: string;
  description: string;
};

type FeatureGridProps = {
  smallHeading: string;
  mainHeading: string;
  cards: FeatureCard[];
};

const FeatureCardComponent = ({ card }: { card: FeatureCard }) => {
  return (
    <Box className="group border border-transparent bg-gradient-to-b py-10 hover:border-gray-100 hover:from-gray-50 dark:hover:border-dark-100 dark:hover:from-dark-300 md:border-l-gray-100 md:px-7 dark:md:border-l-dark-100">
      <Flex align="center" className="mb-5" justify="between">
        {card.icon}
      </Flex>
      <Text as="h3" className="mb-3 text-xl dark:text-white">
        {card.title}
      </Text>
      <Text className="text-[0.95rem] leading-relaxed opacity-60">
        {card.description}
      </Text>
    </Box>
  );
};

export const FeatureGrid = ({
  smallHeading,
  mainHeading,
  cards,
}: FeatureGridProps) => {
  return (
    <Container className="py-28">
      <Text className="mb-8 text-sm uppercase tracking-wider opacity-70">
        {smallHeading}
      </Text>
      <Text as="h2" className="text-5xl font-semibold md:text-6xl">
        {mainHeading}
      </Text>
      <Box className="mt-24 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
        {cards.map((card, index) => (
          <FeatureCardComponent card={card} key={index} />
        ))}
      </Box>
    </Container>
  );
};
