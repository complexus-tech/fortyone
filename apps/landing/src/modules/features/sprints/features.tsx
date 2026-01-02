"use client";

import { Box, Flex, Text } from "ui";
import { StoryIcon, ObjectiveIcon, OKRIcon, SprintsIcon } from "icons";
import { motion } from "framer-motion";
import { useState, type ReactNode } from "react";
import { cn } from "lib";
import { Container } from "@/components/ui";

const Intro = () => (
  <Box className="relative">
    <Box className="flex flex-col py-12 md:flex-row md:items-end md:justify-between md:py-32">
      <motion.div
        initial={{ y: 20, opacity: 0 }}
        transition={{
          duration: 1,
          delay: 0,
        }}
        viewport={{ once: true, amount: 0.5 }}
        whileInView={{ y: 0, opacity: 1 }}
      >
        <Text
          as="h2"
          className="text-5xl font-semibold md:max-w-3xl md:text-7xl"
        >
          Accelerate delivery with{" "}
          <span className="text-stroke-white">agile sprints</span>
        </Text>
      </motion.div>
      <motion.div
        initial={{ y: 20, opacity: 0 }}
        transition={{
          duration: 1,
          delay: 0.3,
        }}
        viewport={{ once: true, amount: 0.5 }}
        whileInView={{ y: 0, opacity: 1 }}
      >
        <Text
          className="mt-6 max-w-xl opacity-80 md:mt-0"
          fontSize="xl"
          fontWeight="normal"
        >
          Optimize your agile development process with our intuitive sprint
          management system. Plan iterations, track progress, and deliver value
          efficiently with every cycle.
        </Text>
      </motion.div>
    </Box>
  </Box>
);

const Card = ({
  name,
  icon,
  description,
}: {
  name: string;
  icon: ReactNode;
  description: string;
}) => {
  const [isActive, setIsActive] = useState(false);
  return (
    <motion.div
      animate={isActive ? { y: -6, x: 6 } : { y: 0, x: 0 }}
      initial={{ y: 0, x: 0 }}
      transition={{ type: "spring", stiffness: 100 }}
    >
      <Box
        className={cn(
          "relative flex min-h-[400px] flex-col justify-between overflow-hidden rounded-3xl border border-border bg-surface p-6 pb-8 md:h-[420px]",
        )}
        onMouseEnter={() => {
          setIsActive(true);
        }}
        onMouseLeave={() => {
          setIsActive(false);
        }}
      >
        <Flex className="mb-4" justify="between">
          <Text as="h3" className="text-lg font-semibold">
            {name}
          </Text>
          {icon}
        </Flex>
        <Box>
          <Text className="mt-4 opacity-80">{description}</Text>
        </Box>
        <div className="pointer-events-none absolute inset-0 z-3 bg-[url('/noise.png')] bg-repeat opacity-50" />
      </Box>
    </motion.div>
  );
};

export const Features = () => {
  const features = [
    {
      id: 1,
      name: "Sprint Planning",
      description:
        "Plan sprints with precision. Allocate tasks, set capacity, and prioritize work items to maximize team efficiency and delivery.",
      icon: <SprintsIcon className="h-6" />,
    },
    {
      id: 2,
      name: "Progress Tracking",
      description:
        "Monitor sprint progress with real-time burndown charts and velocity metrics. Identify bottlenecks early to ensure on-time delivery.",
      icon: <StoryIcon className="h-6" />,
    },
    {
      id: 3,
      name: "Sprint Reviews",
      description:
        "Conduct effective sprint reviews with comprehensive retrospective tools. Identify improvements and refine your agile process.",
      icon: <OKRIcon className="h-6" />,
    },
    {
      id: 4,
      name: "Sprint Forecasting",
      description:
        "Predict sprint outcomes using historical data. Forecast completion dates and capacity needs for better planning.",
      icon: <ObjectiveIcon className="h-6" />,
    },
  ];

  return (
    <Box className="dark bg-[radial-gradient(ellipse_at_center,var(--tw-gradient-stops))] from-dark-200 via-black to-black pb-20 md:pb-48">
      <Container as="section">
        <Intro />
        <Box className="mx-auto grid grid-cols-1 gap-6 md:grid-cols-4">
          {features.map((feature) => (
            <Card key={feature.id} {...feature} />
          ))}
        </Box>
      </Container>
    </Box>
  );
};
