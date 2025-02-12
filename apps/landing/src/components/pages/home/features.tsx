"use client";
import { Box, Text, Flex, Wrapper, Button } from "ui";
import {
  ArrowRightIcon,
  KanbanIcon,
  SprintsIcon,
  StoryIcon,
  AnalyticsIcon,
} from "icons";
import { motion } from "framer-motion";
import { useState } from "react";
import type { ReactNode } from "react";
import Image from "next/image";
import { cn } from "lib";
import { Container, Blur } from "@/components/ui";
import kanbanImg from "../../../../public/kanban.png";
import storyImg from "../../../../public/story1.png";
import sprintImg from "../../../../public/sprints.png";
import analyticsImg from "../../../../public/analytics1.png";

const Intro = () => (
  <Box className="relative">
    <Box as="section" className="my-12 text-center md:my-24">
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
          as="h3"
          className="mx-auto max-w-5xl pb-2 text-5xl md:text-7xl"
          color="gradient"
        >
          Meet your core objectives.
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
          className="mx-auto mt-6 max-w-[700px]"
          fontSize="xl"
          fontWeight="normal"
        >
          Streamline workflows, elevate collaboration, and nail your objectives
          with features like OKR tracking, Kanban boards, Roadmaps, sprints, and
          more.
        </Text>
      </motion.div>
    </Box>
    <Blur className="absolute left-1/2 right-1/2 top-0 z-[4] h-[230vh] w-[70vw] -translate-x-1/2 bg-warning/[0.07] md:h-[70vw]" />
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
      <Box className="h-full">
        <motion.div
          animate={isActive ? { y: -6, x: 6 } : { y: 0, x: 0 }}
          initial={{ y: 0, x: 0 }}
          transition={{ type: "spring", stiffness: 100 }}
        >
          <Wrapper className="h-full rounded-3xl px-6 py-8 shadow-2xl dark:bg-dark-300/30 md:rounded-[2rem]">
            <>{children}</>
          </Wrapper>
        </motion.div>
      </Box>
    </Box>
  );
};

export const Features = () => {
  const features = [
    {
      id: 4,
      name: "Kanban Boards",
      description:
        "Visualize your workflow, track progress, and manage tasks efficiently. Drag and drop tasks to update status.",
      icon: <KanbanIcon className="h-6 w-auto md:h-7" strokeWidth={1.4} />,
      image: {
        src: kanbanImg,
        alt: "Kanban Boards",
      },
      className: "md:col-span-2",
    },
    {
      id: 1,
      name: "Stories",
      description: "Break down complex projects into manageable tasks.",
      icon: <StoryIcon className="h-6 w-auto md:h-7" strokeWidth={1.4} />,
      image: {
        src: storyImg,
        alt: "Stories",
      },
      // className: "md:col-span-2",
    },
    {
      id: 3,
      name: "Sprints",
      description: "Set sprints to achieve your goals and track progress.",
      icon: <SprintsIcon className="h-6 w-auto md:h-7" strokeWidth={1.4} />,
      image: {
        src: sprintImg,
        alt: "Sprints",
      },
    },
    {
      id: 2,
      name: "Analytics & Reporting",
      description:
        "Track your progress and make data-driven decisions. Get insights into your team's performance and productivity.",
      icon: <AnalyticsIcon className="h-6 w-auto md:h-7" strokeWidth={1.4} />,
      image: {
        src: analyticsImg,
        alt: "Stories",
      },
      className: "md:col-span-2",
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
                className={cn(
                  "pointer-events-none mx-auto block rounded-xl border-[3px] border-dark-200 object-cover object-top shadow-xl shadow-dark-300/60 md:aspect-[4/3]",
                  {
                    "md:aspect-[16/5.5]": className === "md:col-span-2",
                  },
                )}
                placeholder="blur"
                src={src}
              />
            </ItemWrapper>
          ),
        )}
      </Box>
    </Container>
  );
};
