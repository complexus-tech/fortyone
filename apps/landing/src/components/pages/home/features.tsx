"use client";
import { Box, Text, Flex, Wrapper, Button } from "ui";
import {
  ArrowRightIcon,
  EpicsIcon,
  KanbanIcon,
  MilestonesIcon,
  ObjectiveIcon,
  StoryIcon,
} from "icons";
import Image from "next/image";
import { motion } from "framer-motion";
import { useState } from "react";
import type { ReactNode } from "react";
import { Container, Blur } from "@/components/ui";

const Intro = () => (
  <Box className="relative">
    <Box as="section" className="my-20 text-center">
      <Text
        as="h3"
        className="mx-auto max-w-4xl pb-2 text-4xl font-medium md:text-7xl"
        color="gradient"
      >
        Say goodbye to wasted time and energy.
      </Text>
      <Text className="mx-auto mt-2 max-w-[700px] md:mt-6" fontSize="lg">
        Simplify workflows, streamline collaboration, and achieve exceptional
        results with Complexus. With features like OKR Tracking, Themes
        Management, Iterations Planning, and Roadmap Visualization, welcome to
        effortless project management.
      </Text>
    </Box>
    <Blur className="absolute left-1/2 right-1/2 top-28 z-[4] h-[800px] w-[800px] -translate-x-1/2 bg-primary/40 dark:bg-secondary/20" />
  </Box>
);

const ItemWrapper = ({
  className,
  children,
}: {
  children: ReactNode;
  className?: string;
}) => {
  const [isActive, setIsActive] = useState(false);
  return (
    <Box
      className={className}
      onMouseEnter={() => {
        setIsActive(true);
      }}
      onMouseLeave={() => {
        setIsActive(false);
      }}
    >
      <motion.div
        animate={isActive ? { y: -6, x: 6 } : { y: 0, x: 0 }}
        className="h-full"
        initial={{ y: 0, x: 0 }}
        transition={{ type: "spring", stiffness: 100 }}
      >
        <Wrapper className="h-full rounded-[2rem] px-6 py-8 shadow-2xl dark:bg-dark-300/30">
          <>{children}</>
        </Wrapper>
      </motion.div>
    </Box>
  );
};

export const Features = () => {
  const features = [
    {
      id: 1,
      name: "Stories",
      description: "Break down complex projects into manageable tasks.",
      icon: <StoryIcon className="h-7 w-auto" />,
      image: {
        src: "/story-card.png",
        alt: "Stories",
        height: 2422,
        width: 1652,
      },
    },
    {
      id: 2,
      name: "Objectives",
      description: "Define your goals, track progress, and measure success.",
      icon: <ObjectiveIcon className="h-7 w-auto" />,
      image: {
        src: "/story-card.png",
        alt: "Stories",
        height: 2422,
        width: 1652,
      },
    },
    {
      id: 3,
      name: "Milestones",
      description: "Set milestones to track progress and celebrate success.",
      icon: <MilestonesIcon className="h-7 w-auto" />,
      image: {
        src: "/story-card.png",
        alt: "Stories",
        height: 2422,
        width: 1652,
      },
    },
    {
      id: 4,
      name: "Kanban Boards",
      description:
        "Visualize your workflow, track progress, and manage tasks efficiently. Drag and drop tasks to update status.",
      icon: <KanbanIcon className="h-7 w-auto" />,
      image: {
        src: "/kanban.png",
        alt: "Kanban Boards",
        height: 2198,
        width: 948,
      },
      className: "col-span-2",
    },

    {
      id: 5,
      name: "Themes",
      description: "Define your goals, track progress, and measure success.",
      icon: <EpicsIcon className="h-7 w-auto" />,
      image: {
        src: "/story-card.png",
        alt: "Stories",
        height: 2422,
        width: 1652,
      },
    },
  ];

  return (
    <Container as="section">
      <Intro />
      <Box className="grid grid-cols-3 gap-8">
        {features.map((feature) => (
          <ItemWrapper className={feature.className} key={feature.id}>
            <Flex align="center" className="mb-3" gap={4} justify="between">
              <Text
                as="h3"
                className="flex items-center gap-2"
                fontSize="2xl"
                fontWeight="medium"
              >
                {feature.icon}
                {feature.name}
              </Text>
              <Button color="tertiary" rounded="full" size="md">
                <ArrowRightIcon className="h-4 w-auto" />
              </Button>
            </Flex>
            <Text className="mb-4" color="muted">
              {feature.description}
            </Text>
            <Image
              alt={feature.image.alt}
              className="mx-auto block rounded-xl border-2 border-dark-200 shadow-xl shadow-dark-300/60"
              height={feature.image.height}
              src={feature.image.src}
              width={feature.image.width}
            />
          </ItemWrapper>
        ))}
      </Box>

      <Image
        alt="Stories"
        className="mx-auto block rounded-xl border-2 border-dark-200 shadow-xl shadow-dark-300/60"
        height={400}
        src="/images/panel.png"
        width={1200}
      />
    </Container>
  );
};
