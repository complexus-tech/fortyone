"use client";
import Image from "next/image";
import { Box, Button, Flex, Text } from "ui";
import { motion } from "framer-motion";
import { Container } from "@/components/ui";
import ctaLight from "../../../public/images/product/cta.webp";
import ctaDark from "../../../public/images/product/cta-dark.webp";

export const CallToAction = () => {
  return (
    <Box className="overflow-hidden border-b border-gray-100/70 bg-gradient-to-t from-gray-50 dark:border-dark-100 dark:from-dark">
      <Container className="relative max-w-7xl pt-16 md:pt-16">
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
              className="mt-6 h-max max-w-4xl pb-2 text-5xl font-semibold md:text-7xl"
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
              className="mt-4 max-w-[650px] md:mt-10"
              color="muted"
              fontSize="xl"
            >
              Plan with Maya, turn ideas into shippable stories, and watch
              progress roll into OKRs automatically.
            </Text>
          </motion.div>
        </Flex>
        <Box className="group relative rounded-t-[0.6rem] border border-b-0 border-[#8080802a] bg-white/50 p-0.5 shadow-2xl backdrop-blur dark:border-dark-50/70 dark:bg-dark-200/40 md:rounded-t-3xl md:px-1.5 md:pb-0 md:pt-1.5">
          <Image
            alt="CTA"
            className="rounded-t-[1.1rem] border border-b-0 border-gray-100 dark:hidden"
            src={ctaLight}
          />
          <Image
            alt="CTA"
            className="hidden rounded-t-[1.1rem] border border-b-0 border-gray-100 dark:block dark:border-dark-100"
            src={ctaDark}
          />
          <Box className="absolute inset-0 flex items-center justify-center opacity-0 backdrop-blur-[2px] transition-opacity duration-300 group-hover:opacity-100">
            <Button
              className="px-3 md:pl-5 md:pr-4"
              color="invert"
              href="/signup"
              rounded="full"
              size="lg"
            >
              Get Started - It&apos;s free
            </Button>
          </Box>
        </Box>
      </Container>
    </Box>
  );
};
