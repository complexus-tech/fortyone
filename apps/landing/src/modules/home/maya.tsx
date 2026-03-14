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
        <Text className="mb-4 font-mono text-sm tracking-wider uppercase opacity-75">
          See Maya in action
        </Text>
        <Text
          as="h3"
          className="mb-10 max-w-4xl pb-2 text-4xl md:mb-4 md:text-5xl"
          color="gradientDark"
        >
          From &ldquo;we should build this&rdquo; to shipped — without the chaos
          in between.
        </Text>
        <Text
          className="mb-10 max-w-3xl leading-relaxed opacity-70 md:mb-14"
          color="muted"
        >
          Type a rough idea. Maya structures it, adds it to the right sprint,
          connects it to your goals, and tracks it to done. You get the credit.
          She does the paperwork.
        </Text>
      </motion.div>
      <motion.div
        initial="hidden"
        variants={scaleIn}
        viewport={viewport}
        whileInView="show"
      >
        <Box className="border-border bg-background/5 shadow-shadow rounded-lg border border-b-0 p-0.5 shadow-2xl md:rounded-2xl md:p-1.5">
          <Flex
            align="center"
            className="mt-1 mb-2 px-1.5 dark:mb-2.5"
            justify="between"
          >
            <Flex className="gap-1.5">
              <Dot className="text-primary size-2.5" />
              <Dot className="text-warning size-2.5" />
              <Dot className="text-success size-2.5" />
            </Flex>
            <ArrowDown2Icon className="h-3.5" strokeWidth={2.5} />
          </Flex>
          <video
            autoPlay
            className="hidden aspect-[6/3.9] h-full w-full object-cover md:rounded-xl dark:block"
            loop
            muted
            src="/videos/intro-dark.mp4"
          />
          <video
            autoPlay
            className="aspect-[6/3.9] h-full w-full object-cover md:rounded-xl dark:hidden"
            loop
            muted
            src="/videos/intro-light.mp4"
          />
        </Box>
      </motion.div>
      <Box className="mt-6 grid grid-cols-1 gap-3 md:mt-8 md:grid-cols-3">
        {[
          "Draft tasks from plain language",
          "Scope sprints around team capacity",
          "Push progress straight into goals",
        ].map((point) => (
          <Box
            className="border-border bg-surface/70 rounded-2xl border px-4 py-4"
            key={point}
          >
            <Text className="font-medium opacity-85">{point}</Text>
          </Box>
        ))}
      </Box>

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
