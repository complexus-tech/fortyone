import type { ReactNode } from "react";
import { Text, Box, Flex } from "ui";
import { cn } from "lib";
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

const FeatureCardComponent = ({
  card,
  index,
  totalCards,
}: {
  card: FeatureCard;
  index: number;
  totalCards: number;
}) => {
  const isLastInRow = (index + 1) % 3 === 0;
  const isInLastRow = index >= totalCards - 3;

  return (
    <Box
      className={cn(
        "group bg-gradient-to-b px-7 py-16 hover:from-gray-50 dark:hover:from-dark-200",
        "border-l border-t dark:border-dark-50",
        {
          "border-r": isLastInRow,
          "border-b": isInLastRow,
        },
      )}
    >
      <Flex align="center" className="mb-6" justify="between">
        {card.icon}
      </Flex>
      <Text as="h3" className="mb-3 text-2xl dark:text-white">
        {card.title}
      </Text>
      <Text className="text-[0.95rem] leading-relaxed opacity-60 group-hover:opacity-100">
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
    <Container className="py-10 md:py-28">
      <Text className="mb-8 text-sm uppercase tracking-wider opacity-70">
        {smallHeading}
      </Text>
      <Text
        as="h2"
        className="pb-4 text-4xl font-semibold md:text-6xl"
        color="gradientDark"
      >
        {mainHeading}
      </Text>
      <Box className="mt-6 grid grid-cols-1 md:mt-16 md:grid-cols-2 lg:grid-cols-3">
        {cards.map((card, index) => (
          <FeatureCardComponent
            card={card}
            index={index}
            key={index}
            totalCards={cards.length}
          />
        ))}
      </Box>
    </Container>
  );
};
