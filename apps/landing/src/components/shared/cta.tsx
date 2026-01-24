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
    <Box className="border-border/70 from-surface-muted overflow-hidden border-b bg-linear-to-t">
      <Container className="relative max-w-7xl pt-6 md:pt-16">
        <Flex className="mb-8 md:mb-12" direction="column">
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
              className="mt-6 h-max max-w-3xl pb-2 text-5xl font-semibold text-balance md:text-6xl md:leading-[1.1]"
            >
              Ready to <span className="text-stroke-white">10x</span> Your
              Team&apos;s Velocity? Start Free
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
              Plan with Maya, turn ideas into shippable tasks, and watch
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
          <Box className="group border-border bg-background/5 d/70 dark:bg-dark-200/40 relative rounded-t-[0.6rem] border border-b-0 p-0.5 pb-0 shadow-2xl backdrop-blur md:rounded-t-2xl md:px-1.5 md:pt-1.5">
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
              className="border-border rounded-t-lg border border-b-0 md:rounded-t-[0.7rem] dark:hidden"
              src={ctaLight}
            />
            <Image
              alt="CTA"
              className="border-border d hidden rounded-t-lg border border-b-0 md:rounded-t-[0.7rem] dark:block"
              src={ctaDark}
            />
            <Box className="group-hover:bg-background/5 absolute inset-0 flex items-center justify-center rounded-t-[0.6rem] transition-colors duration-300 md:rounded-t-2xl">
              <Button
                className="border-0 px-3 opacity-0 backdrop-blur-lg transition-opacity group-hover:opacity-100 md:pr-4 md:pl-5"
                color="invert"
                href="https://cloud.fortyone.app/signup"
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
