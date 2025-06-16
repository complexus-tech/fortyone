"use client";
import { Box, Flex, Text } from "ui";
import { CommandIcon, DashboardIcon, GitIcon, OKRIcon } from "icons";
import { motion } from "framer-motion";
import { useState, type ReactNode } from "react";
import type { StaticImageData } from "next/image";
import Image from "next/image";
import { cn } from "lib";
import { Container } from "@/components/ui";
import keyboardImg from "../../../public/features/keyboard.png";
import analyticsImg from "../../../public/features/analytics.png";
import workflowImg from "../../../public/features/workflow.png";
import okrImg from "../../../public/features/okr.png";

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
          Built on <span className="text-stroke-white">strong</span>{" "}
          foundations.
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
          Complexus is so simple to use, it&apos;s easy to overlook the wealth
          of complex technologies packed under the hood that keep it robust,
          safe, and blazing fast.
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
            <Box className="absolute inset-0 bg-[radial-gradient(ellipse_at_center,_var(--tw-gradient-stops))] from-transparent via-dark/90 to-dark" />
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
      name: "Advanced workflows",
      description:
        "Define complex workflows. Customize statuses, and conditions to meet your team's needs.",
      icon: <GitIcon className="h-6" />,
      image: {
        src: workflowImg,
        alt: "Kanban Board View",
      },
    },
    {
      id: 2,
      name: "Analytics & Insights",
      description:
        "Track progress, identify bottlenecks, and optimize your workflow with detailed analytics.",
      icon: <DashboardIcon className="h-6" />,
      image: {
        src: analyticsImg,
        alt: "Analytics & Insights",
      },
    },
    {
      id: 3,
      name: "Track OKRs",
      description:
        "Set and track OKRs to align your team and measure progress towards goals.",
      icon: <OKRIcon className="h-6" />,
      image: {
        src: okrImg,
        alt: "OKR Tracking",
      },
    },
    {
      id: 4,
      name: "Keyboard shortcuts",
      description:
        "Manage your work faster with intuitive keyboard shortcuts for enhanced productivity.",
      icon: <CommandIcon className="h-6" />,
      image: {
        src: keyboardImg,
        alt: "Keyboard Shortcuts",
      },
    },
  ];

  return (
    <Box className="dark bg-[radial-gradient(ellipse_at_center,_var(--tw-gradient-stops))] from-dark-200 via-black to-black pb-20 md:pb-48">
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
