"use client";
import { Button, Flex, Text, Box } from "ui";
import { motion } from "framer-motion";
import { cn } from "lib";
import { Container, GoogleIcon } from "@/components/ui";
import { SIGNUP_URL } from "@/lib/app-url";

export const Hero = () => {
  return (
    <Box>
      <Box className="absolute inset-0 bg-[linear-gradient(to_right,#8080802a_1px,transparent_1px),linear-gradient(to_bottom,#8080801a_1px,transparent_1px)] bg-size-[45px_45px]" />
      <Box className="absolute inset-0 bg-radial-[at_50%_75%] from-transparent via-white/80 to-white dark:via-black/80 dark:to-black" />
      <Container className="pt-12">
        <Box className="mt-12 mb-6 flex items-end justify-between md:mt-16">
          <motion.span
            initial={{ y: -15, opacity: 0 }}
            transition={{
              duration: 1,
              delay: 0.15,
            }}
            viewport={{ once: true, amount: 0.5 }}
            whileInView={{ y: 0, opacity: 1 }}
          >
            <Text
              as="h1"
              className={cn(
                "relative z-1 text-5xl text-balance md:max-w-4xl md:text-[3.5rem]",
              )}
            >
              Keep every task connected to a goal.
            </Text>
          </motion.span>

          <motion.span
            initial={{ y: -10, opacity: 0 }}
            transition={{
              duration: 1,
              delay: 0.3,
            }}
            viewport={{ once: true, amount: 0.5 }}
            whileInView={{ y: 0, opacity: 1 }}
          >
            <Text className="mb-0.5 max-w-4xl opacity-80">
              Most teams do not lose on strategy — they lose it between the plan
              and the sprint board. FortyOne keeps the goal visible, while Maya
              drafts tasks, scopes sprints, and flags risks early.
            </Text>
          </motion.span>
        </Box>

        <motion.span
          initial={{ y: -10, opacity: 0 }}
          transition={{
            duration: 1,
            delay: 0.4,
          }}
          viewport={{ once: true, amount: 0.5 }}
          whileInView={{ y: 0, opacity: 1 }}
        >
          <Button
            className="relative z-1 px-3 md:pr-4 md:pl-5"
            color="invert"
            href={SIGNUP_URL}
            rounded="lg"
            size="lg"
          >
            Get Started Free
          </Button>
        </motion.span>
      </Container>
    </Box>
  );
};
