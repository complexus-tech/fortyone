"use client";

import { Box, Flex, Text } from "ui";
import { StoryIcon, UserIcon, WorkflowIcon, DashboardIcon } from "icons";
import { motion } from "framer-motion";
import { useState, type ReactNode } from "react";
import { cn } from "lib";
import type { StaticImageData } from "next/image";
import Image from "next/image";
import { Container } from "@/components/ui";
import createImg from "../../../../public/product/stories/create.png";
import assignImg from "../../../../public/product/stories/assign.png";
import automateImg from "../../../../public/product/stories/automate.png";
import progressImg from "../../../../public/product/stories/progress.png";

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
          Powerful story management{" "}
          <span className="text-stroke-white">tools</span>
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
          Streamline your workflow with our intuitive story management system.
          Track progress and deliver value consistently.
        </Text>
      </motion.div>
    </Box>
  </Box>
);

const Card = ({
  name,
  icon,
  description,
  image: { src, alt },
}: {
  name: string;
  icon: ReactNode;
  description: string;
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
          <Box className="relative">
            <Image
              alt={alt}
              className={cn(
                "pointer-events-none mx-auto block aspect-square rounded-xl object-contain",
                {
                  "aspect-[6/4]": name.toLowerCase().includes("keyboard"),
                  "object-bottom": name.toLowerCase().includes("analytics"),
                },
              )}
              placeholder="blur"
              quality={100}
              src={src}
            />
            <Box className="absolute inset-0 bg-[radial-gradient(ellipse_at_center,_var(--tw-gradient-stops))] from-transparent via-dark/90 via-50% to-dark" />
          </Box>
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
      name: "Story Creation",
      description:
        "Create stories with clear criteria. Organize them into sprints and objectives for better flow.",
      icon: <StoryIcon className="h-6" />,
      image: {
        src: createImg,
        alt: "Story Creation",
      },
    },
    {
      id: 2,
      name: "Story Assignment",
      description:
        "Assign stories to team members and track progress in real-time. Keep everyone aligned.",
      icon: <UserIcon className="h-6" />,
      image: {
        src: assignImg,
        alt: "Story assignment",
      },
    },
    {
      id: 3,
      name: "Workflow Automation",
      description:
        "Automate routine tasks with custom workflows. Set up smart notifications and reminders.",
      icon: <WorkflowIcon className="h-6" />,
      image: {
        src: automateImg,
        alt: "Workflow automation",
      },
    },
    {
      id: 4,
      name: "Progress Tracking",
      description:
        "Track progress with burndown charts and velocity metrics. Make data-driven decisions.",
      icon: <DashboardIcon className="h-6" />,
      image: {
        src: progressImg,
        alt: "Progress tracking",
      },
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
