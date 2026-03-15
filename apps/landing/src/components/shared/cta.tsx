"use client";
import Image from "next/image";
import { Box, Button, Flex, Text } from "ui";
import { motion } from "framer-motion";
import { ArrowDown2Icon } from "icons";
import { Container, Dot } from "@/components/ui";
import { SIGNUP_URL } from "@/lib/app-url";
import ctaLight from "../../../public/images/product/cta.webp";
import ctaDark from "../../../public/images/product/cta-dark.webp";

const viewport = { once: true, amount: 0.35 };
const scaleIn = {
  hidden: { scale: 0.98, opacity: 0 },
  show: {
    scale: 1,
    opacity: 1,
    transition: { duration: 0.7, ease: "easeOut" },
  },
};

export const CallToAction = () => {
  return (
    <Box className="border-border/70 from-warning/12 overflow-hidden border-b bg-linear-to-t dark:from-white/20">
      <Container className="relative max-w-7xl pt-6 md:pt-16">
        <Box className="mb-6 flex flex-col gap-6 md:flex-row md:items-center md:justify-between md:gap-12">
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
              as="h2"
              className="mt-6 h-max max-w-3xl pb-2 text-4xl text-balance md:text-5xl"
            >
              Stop managing your project manager. Let Maya do it.
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
            <Text className="w-full max-w-xl md:mt-4 md:mb-0.5" color="muted">
              Your team already knows what to build. FortyOne makes sure every
              task, sprint, and decision is working toward the same goal — and
              Maya keeps it that way, automatically.
            </Text>
          </motion.div>
        </Box>
        <Button
          className="border-0 px-3 backdrop-blur-lg transition-opacity md:pr-4 md:pl-5"
          color="invert"
          href={SIGNUP_URL}
          rounded="lg"
          size="lg"
        >
          Get Started Free
        </Button>
        <motion.div
          initial="hidden"
          variants={scaleIn}
          viewport={viewport}
          whileInView="show"
        >
          <Box className="group border-border/70 bg-background/30 dark:bg-dark-200/40 relative mt-12 rounded-t-lg border border-b-0 p-0.5 pb-0 shadow-2xl backdrop-blur md:rounded-t-xl md:px-1.5 md:pt-1.5">
            <Flex align="center" className="mt-1 mb-2 px-1.5" justify="between">
              <Flex className="gap-1.5">
                <Dot className="text-primary size-2.5" />
                <Dot className="text-warning size-2.5" />
                <Dot className="text-success size-2.5" />
              </Flex>
              <ArrowDown2Icon className="h-3.5" strokeWidth={2.5} />
            </Flex>
            <Image
              alt="CTA"
              className="border-border/70 rounded-t-lg border border-b-0 md:rounded-t-xl dark:hidden"
              src={ctaLight}
              quality={100}
            />
            <Image
              alt="CTA"
              className="border-border/70 hidden rounded-t-lg border border-b-0 dark:block"
              src={ctaDark}
              quality={100}
            />
          </Box>
        </motion.div>
      </Container>
    </Box>
  );
};
