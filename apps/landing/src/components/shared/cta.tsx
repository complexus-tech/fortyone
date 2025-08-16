"use client";
import Image from "next/image";
import { Box, Button, Flex, Text } from "ui";
import { motion } from "framer-motion";
import { ArrowDown2Icon } from "icons";
import { Container, Dot } from "@/components/ui";
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
    <Box className="overflow-hidden border-b border-gray-100/70 bg-gradient-to-t from-gray-50 dark:border-dark-200 dark:from-dark-300/70">
      <Container className="relative max-w-7xl pt-6 md:pt-16">
        <Flex
          align="center"
          className="mb-8 text-center md:mb-12"
          direction="column"
        >
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
              className="mt-6 h-max max-w-2xl pb-2 text-5xl font-semibold md:text-6xl md:leading-[1.1]"
            >
              Work <span className="text-stroke-white">smarter</span> with AI
              that’s in the loop.
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
              className="mt-4 max-w-[650px] text-lg md:mt-5 md:text-lg"
              color="muted"
            >
              Plan with Maya, turn ideas into shippable stories, and watch
              progress roll into OKRs automatically.
            </Text>
          </motion.div>
        </Flex>
        <motion.div
          initial="hidden"
          variants={scaleIn}
          viewport={viewport}
          whileInView="show"
        >
          <Box className="group relative rounded-t-[0.6rem] border border-b-0 border-gray-100 bg-dark/5 p-0.5 pb-0 shadow-2xl backdrop-blur dark:border-dark-50/70 dark:bg-dark-200/40 md:rounded-t-2xl md:px-1.5 md:pt-1.5">
            <Flex align="center" className="mb-2 mt-1 px-1.5" justify="between">
              <Flex className="gap-1.5">
                <Dot className="size-2.5 text-primary" />
                <Dot className="size-2.5 text-warning" />
                <Dot className="size-2.5 text-success" />
              </Flex>
              <ArrowDown2Icon className="h-3.5" strokeWidth={2.5} />
            </Flex>
            <Image
              alt="CTA"
              className="rounded-t-[0.5rem] border border-b-0 border-gray-100 dark:hidden md:rounded-t-[0.7rem]"
              src={ctaLight}
            />
            <Image
              alt="CTA"
              className="hidden rounded-t-[0.5rem] border border-b-0 border-gray-100 dark:block dark:border-dark-100 md:rounded-t-[0.7rem]"
              src={ctaDark}
            />
            <Box className="absolute inset-0 flex items-center justify-center rounded-t-[0.6rem] transition-colors duration-300 group-hover:bg-dark/5 md:rounded-t-2xl">
              <Button
                className="border-0 px-3 opacity-0 backdrop-blur-lg transition-opacity group-hover:opacity-100 md:pl-5 md:pr-4"
                color="invert"
                href="/signup"
                rounded="lg"
                size="lg"
              >
                Get Started - It&apos;s free
              </Button>
            </Box>
          </Box>
        </motion.div>
      </Container>
    </Box>
  );
};
