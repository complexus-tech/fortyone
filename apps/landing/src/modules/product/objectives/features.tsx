"use client";

import { Box, Flex, Text } from "ui";
import { StoryIcon, ObjectiveIcon, OKRIcon, SprintsIcon } from "icons";
import { motion } from "framer-motion";
import { useState, type ReactNode } from "react";
import { cn } from "lib";
import { Container } from "@/components/ui";

const Intro = () => (
  <Box className="relative">
    <Box className="flex flex-col py-12 md:flex-row md:items-end md:py-32">
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
          className="text-5xl font-semibold md:max-w-5xl md:text-7xl"
        >
          Define and achieve{" "}
          <span className="text-stroke-white">strategic goals</span>
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
          Create, track, and achieve your organizational objectives with our
          comprehensive system. Align efforts and drive consistent results
          toward your strategic priorities.
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
          "relative flex min-h-[400px] flex-col justify-between overflow-hidden rounded-3xl border border-dark-50 bg-dark p-6 pb-8 md:h-[420px]",
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
        <div className="pointer-events-none absolute inset-0 z-[3] bg-[url('/noise.png')] bg-repeat opacity-50" />
      </Box>
    </motion.div>
  );
};

export const Features = () => {
  const features = [
    {
      id: 1,
      name: "Objective Creation",
      description:
        "Define clear, measurable objectives aligned with your strategic vision. Set timeframes and assign ownership for accountability.",
      icon: <ObjectiveIcon className="h-6" />,
    },
    {
      id: 2,
      name: "Progress Tracking",
      description:
        "Monitor objective progress with visual dashboards. Get real-time updates on completion status and identify at-risk goals early.",
      icon: <StoryIcon className="h-6" />,
    },
    {
      id: 3,
      name: "Team Alignment",
      description:
        "Cascade objectives throughout your organization. Ensure every team understands how their work contributes to strategic goals.",
      icon: <OKRIcon className="h-6" />,
    },
    {
      id: 4,
      name: "Objective Analytics",
      description:
        "Generate insights on objective completion rates and trends. Identify patterns to improve your strategic planning process.",
      icon: <SprintsIcon className="h-6" />,
    },
  ];

  return (
    <Box className="bg-[radial-gradient(ellipse_at_center,_var(--tw-gradient-stops))] from-dark-200 via-black to-black pb-20 md:pb-48">
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
