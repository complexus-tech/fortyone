"use client";

import { Box, Container, Flex, Text, Button } from "ui";
import { motion } from "framer-motion";
import { ArrowRight2Icon } from "icons";
import { useSession } from "next-auth/react";
import Image from "next/image";
import { GoogleIcon } from "@/components/ui";
import { signInWithGoogle } from "@/lib/actions/sign-in";
import sprintsImg from "../../../../public/images/product/kanban.webp"; // Using existing image; ideally replace with Sprint-specific image

export const Hero = () => {
  const { data: session } = useSession();
  return (
    <Box>
      <Container className="pt-12 md:pt-16">
        <Flex
          align="center"
          className="mb-8 mt-20 text-center"
          direction="column"
        >
          <motion.span
            initial={{ y: -10, opacity: 0 }}
            transition={{
              duration: 1,
              delay: 0,
            }}
            viewport={{ once: true, amount: 0.5 }}
            whileInView={{ y: 0, opacity: 1 }}
          >
            <Button
              className="cursor-text px-3 text-sm md:text-base"
              color="tertiary"
              rounded="full"
              size="sm"
            >
              Sprints
            </Button>
          </motion.span>
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
              className="mt-6 pb-2 text-5xl font-semibold md:max-w-3xl md:text-7xl md:leading-[1.1]"
            >
              <span className="text-stroke-white">Deliver</span> Value with
              Agile Sprints
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
            <Text
              className="mt-8 max-w-3xl text-lg opacity-80 md:text-2xl"
              fontWeight="normal"
            >
              Streamline your development cycles with our powerful sprint
              management tools. Plan, execute, and track sprint progress to
              deliver value consistently and predictably.
            </Text>
          </motion.span>

          <Flex
            align="center"
            className="relative mt-6 justify-center gap-2 md:mt-10 md:gap-4"
            wrap
          >
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
                className="px-3 font-semibold md:pl-5 md:pr-4"
                href="/signup"
                rightIcon={
                  <ArrowRight2Icon className="text-white dark:text-gray-200" />
                }
                rounded="full"
                size="lg"
              >
                <span className="hidden md:inline">
                  Start Your First Sprint
                </span>
                <span className="md:hidden">Get Started</span>
              </Button>
            </motion.span>
            <motion.span
              initial={{ y: -5, opacity: 0 }}
              transition={{
                duration: 1,
                delay: 0.6,
              }}
              viewport={{ once: true, amount: 0.5 }}
              whileInView={{ y: 0, opacity: 1 }}
            >
              <Button
                className="px-3 md:pl-3.5 md:pr-4"
                color="tertiary"
                leftIcon={<GoogleIcon />}
                onClick={async () => {
                  await signInWithGoogle();
                }}
                rounded="full"
                size="lg"
              >
                {session ? "Continue with Google" : "Sign up with Google"}
              </Button>
            </motion.span>
          </Flex>
        </Flex>
        <Box className="relative mx-auto mt-16 max-w-6xl">
          <Image
            alt="Sprint Management Dashboard - Agile sprint planning and tracking"
            className="rounded border-[6px] dark:border-dark-100 md:rounded-2xl"
            placeholder="blur"
            src={sprintsImg}
          />
          <Box className="absolute inset-0 bg-gradient-to-t from-black via-black via-20%" />
        </Box>
      </Container>
    </Box>
  );
};
