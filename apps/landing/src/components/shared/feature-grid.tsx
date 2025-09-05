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
  cards: FeatureCard[];
};

const FeatureCardComponent = ({ card }: { card: FeatureCard }) => {
  return (
    <Box
      className={cn(
        "group border-[0.5px] border-gray-100 bg-gradient-to-b px-6 py-8 hover:from-gray-50 dark:border-dark-50 dark:hover:from-dark-200 md:px-7 md:py-16",
      )}
    >
      <Flex align="center" className="mb-6" justify="between">
        {card.icon}
      </Flex>
      <Text as="h3" className="mb-3 text-xl dark:text-white md:text-2xl">
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
      <motion.div
        initial={{ y: 20, opacity: 0 }}
        transition={{ duration: 0.6, delay: 0 }}
        viewport={{ once: true, amount: 0.5 }}
        whileInView={{ y: 0, opacity: 1 }}
      >
        <Text className="mb-8 text-sm uppercase tracking-wider opacity-70">
          {smallHeading}
        </Text>
      </motion.div>
      <motion.div
        initial={{ y: 20, opacity: 0 }}
        transition={{ duration: 0.8, delay: 0.2 }}
        viewport={{ once: true, amount: 0.5 }}
        whileInView={{ y: 0, opacity: 1 }}
      >
        <Text
          as="h2"
          className="pb-4 text-4xl font-semibold md:text-5xl"
          color="gradientDark"
        >
          {mainHeading}
        </Text>
      </motion.div>
      <motion.div
        initial={{ y: 20, opacity: 0 }}
        transition={{ duration: 0.8, delay: 0.2 }}
        viewport={{ once: true, amount: 0.2 }}
        whileInView={{ y: 0, opacity: 1 }}
      >
        <Box className="mt-6 grid grid-cols-1 border-[0.5px] border-gray-100 dark:border-dark-50 md:mt-16 md:grid-cols-2 lg:grid-cols-3">
          {cards.map((card, index) => (
            <FeatureCardComponent card={card} key={index} />
          ))}
        </Box>
      </motion.div>
    </Container>
  );
};
