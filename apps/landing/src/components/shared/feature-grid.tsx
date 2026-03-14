"use client";
import type { ReactNode } from "react";
import { Text, Box, Flex } from "ui";
import { motion } from "framer-motion";
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
  description?: string;
  cards: FeatureCard[];
};

const FeatureCardComponent = ({ card }: { card: FeatureCard }) => {
  return (
    <Box
      className={cn(
        "group border-border/80 hover:from-surface-muted dark:border-border/80 dark:hover:from-surface-elevated border-[0.5px] bg-linear-to-b px-6 py-8 md:px-7 md:py-10",
      )}
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
    <Container className="py-10 md:py-20">
      <motion.div
        initial={{ y: 20, opacity: 0 }}
        transition={{ duration: 0.6, delay: 0 }}
        viewport={{ once: true, amount: 0.5 }}
        whileInView={{ y: 0, opacity: 1 }}
      >
        <Text className="mb-6 font-mono text-sm tracking-wider uppercase opacity-80">
          {smallHeading}
        </Text>
      </motion.div>
      <Box className="flex items-end justify-between gap-16">
        <motion.div
          initial={{ y: 20, opacity: 0 }}
          transition={{ duration: 0.8, delay: 0.2 }}
          viewport={{ once: true, amount: 0.5 }}
          whileInView={{ y: 0, opacity: 1 }}
        >
          <Text
            as="h2"
            className="max-w-3xl pb-1 text-4xl md:text-5xl"
            color="gradientDark"
          >
            {mainHeading}
          </Text>
        </motion.div>
        <motion.div
          initial={{ y: 20, opacity: 0 }}
          transition={{ duration: 0.8, delay: 0.2 }}
          viewport={{ once: true, amount: 0.5 }}
          whileInView={{ y: 0, opacity: 1 }}
        >
          {description ? (
            <Text className="mb-0.5 max-w-3xl leading-relaxed opacity-70">
              {description}
            </Text>
          ) : null}
        </motion.div>
      </Box>
      <motion.div
        initial={{ y: 20, opacity: 0 }}
        transition={{ duration: 0.8, delay: 0.2 }}
        viewport={{ once: true, amount: 0.2 }}
        whileInView={{ y: 0, opacity: 1 }}
      >
        <Box className="border-border/80 d mt-6 grid grid-cols-1 border-[0.5px] md:mt-16 md:grid-cols-2 lg:grid-cols-3">
          {cards.map((card, index) => (
            <FeatureCardComponent card={card} key={index} />
          ))}
        </Box>
      </motion.div>
    </Container>
  );
};
