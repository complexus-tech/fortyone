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
import { motion } from "framer-motion";
import { useState } from "react";
import type { ReactNode } from "react";
import Image from "next/image";
import { Container, Blur } from "@/components/ui";
import storyCard from "../../../../public/story-card.png";
import kanbanImg from "../../../../public/kanban.png";

const Intro = () => (
  <Box className="relative">
    <Box as="section" className="my-12 text-center md:my-24">
      <motion.div
        initial={{ y: 20, opacity: 0 }}
        viewport={{ once: true, amount: 0.5 }}
        transition={{
          duration: 1,
          delay: 0,
        }}
        whileInView={{ y: 0, opacity: 1 }}
      >
        <Text
          as="h3"
          className="mx-auto max-w-5xl pb-2 text-5xl md:text-7xl"
          color="gradient"
        >
          Meet your core objectives.
        </Text>
      </motion.div>
      <motion.div
        initial={{ y: 20, opacity: 0 }}
        viewport={{ once: true, amount: 0.5 }}
        transition={{
          duration: 1,
          delay: 0.3,
        }}
        whileInView={{ y: 0, opacity: 1 }}
      >
        <Text
          className="mx-auto mt-6 max-w-[700px]"
          fontSize="xl"
          fontWeight="normal"
        >
          Streamline workflows, elevate collaboration, and nail your objectives
          with features like OKR tracking, Kanban boards, Roadmaps, milestones,
          and more.
        </Text>
      </motion.div>
    </Box>
    <Blur className="absolute left-1/2 right-1/2 top-28 z-[4] h-[800px] w-[800px] -translate-x-1/2 bg-warning/5" />
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
        <Wrapper className="h-full rounded-3xl px-6 py-8 shadow-2xl dark:bg-dark-300/30 md:rounded-[2rem]">
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
      icon: <StoryIcon strokeWidth={1.4} className="h-6 w-auto md:h-7" />,
      image: {
        src: storyCard,
        alt: "Stories",
        height: 2422,
        width: 1652,
      },
    },
    {
      id: 2,
      name: "Objectives",
      description: "Define your goals, track progress, and measure success.",
      icon: <ObjectiveIcon strokeWidth={1.4} className="h-6 w-auto md:h-7" />,
      image: {
        src: storyCard,
        alt: "Stories",
        height: 2422,
        width: 1652,
      },
    },
    {
      id: 3,
      name: "Sprints",
      description: "Set sprints to achieve your goals and track progress.",
      icon: <MilestonesIcon strokeWidth={1.4} className="h-6 w-auto md:h-7" />,
      image: {
        src: storyCard,
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
      icon: <KanbanIcon strokeWidth={1.4} className="h-6 w-auto md:h-7" />,
      image: {
        src: kanbanImg,
        alt: "Kanban Boards",
        height: 2198,
        width: 948,
      },
      className: "md:col-span-2",
    },

    {
      id: 5,
      name: "Epics",
      description: "Group related stories together to manage large projects.",
      icon: <EpicsIcon strokeWidth={1.4} className="h-6 w-auto md:h-7" />,
      image: {
        src: storyCard,
        alt: "Stories",
        height: 2422,
        width: 1652,
      },
    },
  ];

  return (
    <Container as="section" className="mb-20 md:mb-40">
      <Intro />
      <Box className="grid grid-cols-1 gap-8 md:grid-cols-3">
        {features.map(
          ({ id, icon, name, description, className, image: { src, alt } }) => (
            <ItemWrapper className={className} key={id}>
              <Flex align="center" className="mb-3" gap={4} justify="between">
                <Text
                  as="h3"
                  className="flex items-center gap-2"
                  fontSize="2xl"
                  fontWeight="medium"
                >
                  {icon}
                  {name}
                </Text>
                <Button color="tertiary" rounded="full" size="sm">
                  <ArrowRightIcon className="h-4 w-auto" />
                </Button>
              </Flex>
              <Text className="mb-4" color="muted">
                {description}
              </Text>
              <Image
                alt={alt}
                placeholder="blur"
                className="pointer-events-none mx-auto block rounded-xl border-2 border-dark-200 shadow-xl shadow-dark-300/60"
                src={src}
              />
            </ItemWrapper>
          ),
        )}
      </Box>
    </Container>
  );
};
