"use client";

import { Box, Container, Flex, Text, Button } from "ui";
import { motion } from "framer-motion";
import Image from "next/image";
import { GoogleIcon } from "@/components/ui";
import kanbanImg from "../../../../public/images/product/kanban.webp";
import kanbanImgLight from "../../../../public/images/product/kanban-light.webp";

export const Hero = () => {
  return (
    <Box>
      <Container className="pt-12 md:pt-16">
        <Flex
          align="center"
          className="mt-20 mb-8 text-center"
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
              rounded="lg"
              size="sm"
            >
              Stories
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
              <span className="text-stroke-white">Manage</span> Tasks with User
              Stories
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
              Transform your project management with our intuitive story-based
              workflow system. Create, track, and manage tasks efficiently while
              keeping your team aligned and productive.
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
                className="px-3 md:pr-4 md:pl-5"
                color="invert"
                href="https://cloud.fortyone.app/signup"
                rounded="lg"
                size="lg"
              >
                <span className="hidden md:inline">
                  Create Your First Story
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
                className="px-3 md:pr-4 md:pl-3.5"
                color="tertiary"
                href="https://cloud.fortyone.app/signup"
                leftIcon={<GoogleIcon />}
                rounded="lg"
                size="lg"
              >
                Continue with Google
              </Button>
            </motion.span>
          </Flex>
        </Flex>
        <Box className="relative mx-auto mt-16 max-w-6xl dark:hidden">
          <Image
            alt="Kanban"
            className="border-border rounded border-[6px] md:rounded-2xl"
            placeholder="blur"
            src={kanbanImgLight}
          />
          <Box className="absolute inset-0 bg-linear-to-t from-white via-white via-20%" />
        </Box>
        <Box className="relative mx-auto mt-16 hidden max-w-6xl dark:block">
          <Image
            alt="Kanban"
            className="d rounded border-[6px] md:rounded-2xl"
            placeholder="blur"
            src={kanbanImg}
          />
          <Box className="absolute inset-0 bg-linear-to-t from-black via-black via-20%" />
        </Box>
      </Container>
    </Box>
  );
};
