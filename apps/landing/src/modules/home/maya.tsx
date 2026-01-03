"use client";

import { Text, Box, Flex } from "ui";
import { motion } from "framer-motion";
import { ArrowDown2Icon } from "icons";
import { Container, Dot } from "@/components/ui";

export const Maya = () => {
  const viewport = { once: true, amount: 0.35 };
  const fadeUp = {
    hidden: { y: 20, opacity: 0 },
    show: { y: 0, opacity: 1, transition: { duration: 0.7, ease: "easeOut" } },
  };
  const scaleIn = {
    hidden: { scale: 0.97, opacity: 0 },
    show: {
      scale: 1,
      opacity: 1,
      transition: { duration: 0.8, ease: "easeOut" },
    },
  };
  return (
    <Container className="dark:py-10">
      <motion.div
        initial="hidden"
        variants={fadeUp}
        viewport={viewport}
        whileInView="show"
      >
        <Text
          as="h3"
          className="mb-10 max-w-5xl pb-2 text-4xl font-semibold leading-[1.15] md:mb-16 md:text-5xl"
          color="gradientDark"
        >
          Your AI assistant, helps your team plan sprints, track objectives, and
          catch bottlenecks before they slow you down.
        </Text>
      </motion.div>
      <motion.div
        initial="hidden"
        variants={scaleIn}
        viewport={viewport}
        whileInView="show"
      >
        <Box className="rounded-[0.6rem] border border-b-0 border-border bg-background/5 p-0.5 shadow-2xl shadow-shadow md:rounded-2xl md:p-1.5">
          <Flex
            align="center"
            className="mb-2 mt-1 px-1.5 dark:mb-2.5"
            justify="between"
          >
            <Flex className="gap-1.5">
              <Dot className="size-2.5 text-primary" />
              <Dot className="size-2.5 text-warning" />
              <Dot className="size-2.5 text-success" />
            </Flex>
            <ArrowDown2Icon className="h-3.5" strokeWidth={2.5} />
          </Flex>
          <video
            autoPlay
            className="hidden aspect-[6/3.9] h-full w-full object-cover dark:block md:rounded-[0.7rem]"
            loop
            muted
            src="/videos/intro-dark.mp4"
          />
          <video
            autoPlay
            className="aspect-[6/3.9] h-full w-full object-cover dark:hidden md:rounded-[0.7rem]"
            loop
            muted
            src="/videos/intro-light.mp4"
          />
        </Box>
      </motion.div>

      {/* <Box className="mt-20 grid grid-cols-1 gap-16 md:grid-cols-2">
        <motion.div
          initial="hidden"
          variants={fadeUp}
          viewport={viewport}
          whileInView="show"
        >
          <Box>
            <Text className="mb-5 text-4xl font-semibold">
              AI that works like a teammate
            </Text>
            <Text className="max-w-lg text-lg" color="muted">
              Maya keeps projects moving, anticipating next steps and removing
              blockers before they slow you down.
            </Text>
            <Box className="mt-8 h-60 rounded-3xl bg-surface-muted" />
          </Box>
        </motion.div>
        <motion.div
          initial="hidden"
          transition={{ delay: 0.1 }}
          variants={fadeUp}
          viewport={viewport}
          whileInView="show"
        >
          <Box>
            <Text className="mb-5 text-4xl font-semibold">
              Plans that write themselves
            </Text>
            <Text className="max-w-lg text-lg" color="muted">
              Turn rough ideas into fully scoped sprints in seconds, with AI
              handling the details.
            </Text>
            <Box className="mt-8 h-60 rounded-3xl bg-surface-muted" />
          </Box>
        </motion.div>
      </Box> */}
    </Container>
  );
};
