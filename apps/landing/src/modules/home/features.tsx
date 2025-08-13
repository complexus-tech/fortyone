"use client";
import { Box, Text } from "ui";
import { motion } from "framer-motion";
import { useState, type ReactNode } from "react";
import type { StaticImageData } from "next/image";
import { Container } from "@/components/ui";
import analyticsImg from "../../../public/features/analytics.png";
import workflowImg from "../../../public/features/workflow.png";
import okrImg from "../../../public/features/okr.png";

const Intro = () => (
  <Box className="relative">
    <Box className="flex flex-col gap-8 pb-12 md:flex-row md:gap-12 md:py-16">
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
          className="text-5xl font-semibold md:max-w-2xl md:text-6xl"
        >
          Built for ambitious teams who want{" "}
          <span className="text-stroke-white">results</span>
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
        <Text className="mt-6 max-w-xl md:mt-1" color="muted" fontSize="xl">
          Complexus puts an AI assistant at the heart of your workflow helping
          you create stories, plan sprints, track OKRs, and keep everything
          moving, so your team stays aligned and delivers without the chaos.
        </Text>
      </motion.div>
    </Box>
  </Box>
);

const Card = ({
  description,
  image: { alt, src },
}: {
  description: ReactNode;
  image: { src: StaticImageData; alt: string };
}) => {
  const [isActive, setIsActive] = useState(false);
  return (
    <motion.div
      animate={isActive ? { y: -6, x: 6 } : { y: 0, x: 0 }}
      initial={{ y: 0, x: 0 }}
      transition={{ type: "spring", stiffness: 100 }}
    >
      <Box
        onMouseEnter={() => {
          setIsActive(true);
        }}
        onMouseLeave={() => {
          setIsActive(false);
        }}
      >
        <Box className="mb-6 aspect-square rounded-2xl bg-gray-50" />

        {description}
      </Box>
    </motion.div>
  );
};

export const Features = () => {
  const features = [
    {
      id: 1,
      description: (
        <Text color="muted">
          <Text as="span" color="gradient" fontWeight="bold">
            Turn goals into action.
          </Text>{" "}
          Set clear objectives and connect them to the tasks. Track progress in
          real time and adjust fast.
        </Text>
      ),
      image: {
        src: workflowImg,
        alt: "Kanban Board View",
      },
    },
    {
      id: 2,
      description: (
        <Text color="muted">
          <Text as="span" color="gradient" fontWeight="bold">
            Plan with confidence.
          </Text>{" "}
          Organise sprints with clear priorities and owners. Spot risks early
          and keep deadlines on track.
        </Text>
      ),
      image: {
        src: analyticsImg,
        alt: "Analytics & Insights",
      },
    },
    {
      id: 3,
      description: (
        <Text color="muted">
          <Text as="span" color="gradient" fontWeight="bold">
            Keep everyone aligned.
          </Text>{" "}
          One shared plan for all your work. Less confusion, faster decisions,
          better results.
        </Text>
      ),
      image: {
        src: okrImg,
        alt: "OKR Tracking",
      },
    },
  ];

  return (
    <Box className="pb-20 dark:bg-[radial-gradient(ellipse_at_center,_var(--tw-gradient-stops))] dark:from-dark-200 dark:via-black dark:to-black md:pb-36">
      <Container as="section">
        <Intro />
        <Box className="mx-auto grid grid-cols-1 gap-8 md:grid-cols-3">
          {features.map((feature) => (
            <Card key={feature.id} {...feature} />
          ))}
        </Box>
      </Container>
    </Box>
  );
};
