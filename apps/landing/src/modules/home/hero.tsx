"use client";
import { Box, Button, Text } from "ui";
import { domAnimation, LazyMotion, m, useReducedMotion } from "framer-motion";
import { cn } from "lib";
import { Container } from "@/components/ui";
import { SIGNUP_URL } from "@/lib/app-url";

const createRevealMotion = (delay: number, reduceMotion: boolean) => ({
  initial: reduceMotion
    ? { opacity: 0 }
    : {
        y: 22,
        scale: 0.96,
        opacity: 0,
        filter: "blur(14px)",
      },
  whileInView: { y: 0, scale: 1, opacity: 1, filter: "blur(0px)" },
  transition: reduceMotion
    ? {
        duration: 0.5,
        delay,
      }
    : {
        delay,
        y: {
          duration: 1.15,
          ease: [0.16, 1, 0.3, 1],
        },
        scale: {
          duration: 1.15,
          ease: [0.16, 1, 0.3, 1],
        },
        opacity: {
          duration: 0.8,
          ease: "easeOut",
        },
        filter: {
          duration: 1.05,
          ease: "easeOut",
        },
      },
  viewport: { once: true, amount: 0.5 },
});

export const Hero = () => {
  const shouldReduceMotion = useReducedMotion() ?? false;

  return (
    <LazyMotion features={domAnimation}>
      <Box>
        <Box className="absolute inset-0 hidden bg-[linear-gradient(to_right,#8080802a_1px,transparent_1px),linear-gradient(to_bottom,#8080801a_1px,transparent_1px)] bg-size-[45px_45px] md:block" />
        <Box className="absolute inset-0 hidden bg-radial-[at_50%_75%] from-transparent via-white/80 to-white md:block dark:via-black/70 dark:to-black" />
        <Container className="pt-12">
          <Box className="mt-12 mb-6 flex flex-col gap-6 md:mt-24 md:flex-row md:items-end md:justify-between md:gap-12">
            <m.div {...createRevealMotion(0.15, shouldReduceMotion)}>
              <Text
                as="h1"
                className={cn(
                  "relative z-1 text-5xl font-medium text-balance md:max-w-7xl md:text-6xl",
                )}
              >
                Keep every task connected to a goal.
              </Text>
            </m.div>

            <m.div {...createRevealMotion(0.3, shouldReduceMotion)}>
              <Text className="w-full max-w-xl opacity-80 md:mb-0.5">
                Most teams do not lose on strategy — they lose it between the
                plan and the sprint board. FortyOne keeps the goal visible,
                while Maya drafts tasks, scopes sprints, and flags risks early.
              </Text>
            </m.div>
          </Box>

          <m.div {...createRevealMotion(0.4, shouldReduceMotion)}>
            <Button
              className="relative z-1 px-3 md:pr-4 md:pl-5"
              color="invert"
              href={SIGNUP_URL}
              rounded="lg"
              size="lg"
            >
              Get Started Free
            </Button>
          </m.div>
        </Container>
      </Box>
    </LazyMotion>
  );
};
