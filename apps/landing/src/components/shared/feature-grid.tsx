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
  smallHeading?: string;
  mainHeading: string;
  description?: string;
  cards: FeatureCard[];
};

const FeatureCardComponent = ({
  card,
  index,
}: {
  card: FeatureCard;
  index: number;
}) => {
  return (
    <Box
      className={cn(
        "group border-border/80 hover:from-surface-muted dark:border-border/80 dark:hover:from-surface-elevated border-[0.5px] bg-linear-to-b px-6 py-8 md:px-7 md:py-10",
      )}
      data-landing-reveal
      style={{ transitionDelay: `${index * 60}ms` }}
    >
      <Flex align="center" className="mb-6" justify="between">
        {card.icon}
      </Flex>
      <Text as="h3" className="mb-4 text-xl md:text-2xl">
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
  description,
  cards,
}: FeatureGridProps) => {
  return (
    <Container className="py-10 md:py-28">
      {smallHeading ? (
        <Box data-landing-reveal>
          <Text className="mb-6 font-mono text-sm tracking-wider uppercase opacity-80">
            {smallHeading}
          </Text>
        </Box>
      ) : null}

      <Box className="flex flex-col gap-6 md:flex-row md:items-baseline md:justify-between md:gap-16">
        <Box data-landing-reveal>
          <Text as="h2" className="max-w-4xl pb-1 text-4xl md:text-5xl">
            {mainHeading}
          </Text>
        </Box>
        <Box data-landing-reveal style={{ transitionDelay: "70ms" }}>
          {description ? (
            <Text className="w-full max-w-xl leading-relaxed opacity-70 md:mb-0.5">
              {description}
            </Text>
          ) : null}
        </Box>
      </Box>
      <Box className="border-border/80 d mt-6 grid grid-cols-1 border-[0.5px] md:mt-16 md:grid-cols-2 lg:grid-cols-3">
        {cards.map((card, index) => (
          <FeatureCardComponent card={card} index={index} key={card.title} />
        ))}
      </Box>
    </Container>
  );
};
