"use client";

import { Text, Box } from "ui";
import { motion } from "framer-motion";
import { Container } from "@/components/ui";

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
          className="mb-16 max-w-5xl text-6xl font-semibold leading-[1.15] md:mb-20"
          color="gradient"
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
        <video
          autoPlay
          className="aspect-video h-full w-full object-cover md:rounded-3xl"
          loop
          muted
          src="/videos/intro-dark.mp4"
        />
      </motion.div>

      <Box className="mt-20 grid grid-cols-1 gap-16 md:grid-cols-2">
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
            <Box className="mt-8 h-60 rounded-3xl bg-gray-100 dark:bg-dark-300/80" />
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
            <Box className="mt-8 h-60 rounded-3xl bg-gray-100 dark:bg-dark-300/80" />
          </Box>
        </motion.div>
      </Box>
    </Container>
  );
};
