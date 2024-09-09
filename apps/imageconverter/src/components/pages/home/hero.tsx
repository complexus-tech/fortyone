"use client";
import { Button, Flex, Text, Box } from "ui";
import { ArrowRightIcon } from "icons";
import { motion } from "framer-motion";
import { Container, Blur } from "@/components/ui";
import { FilePicker } from "./filepicker";

export const Hero = () => {
  return (
    <Box className="relative">
      <Blur className="absolute -top-1/2 left-1/2 right-1/2 z-[4] h-[80vh] w-screen -translate-x-1/2 bg-warning/10 dark:bg-warning/[0.05]" />
      <Container className="max-w-3xl pt-12 md:pt-20">
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
              className="px-3 text-sm md:text-base"
              color="tertiary"
              rightIcon={<ArrowRightIcon className="h-3 w-auto" />}
              rounded="full"
              size="sm"
            >
              New: WebP Conversion Available
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
              className="mt-8 max-w-5xl pb-2 font-satoshi text-6xl font-extrabold md:text-[4rem] md:leading-[1.1]"
            >
              Convert Images In <br />
              <Text as="span" color="gradient">
                Seconds.
              </Text>
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
              className="mt-4 max-w-[750px] text-lg font-medium md:mt-8 md:text-[1.35rem]"
              color="muted"
              as="h2"
            >
              Transform any image instantly. Our fast converter handles all
              formats, solving your conversion needs in one click.
            </Text>
          </motion.span>

          <Box className="mx-auto mt-12 w-full max-w-3xl">
            <motion.div
              initial={{ y: -10, opacity: 0 }}
              transition={{
                duration: 1,
                delay: 0.4,
              }}
              viewport={{ once: true, amount: 0.5 }}
              whileInView={{ y: 0, opacity: 1 }}
              style={{
                width: "100%",
              }}
            >
              <FilePicker />
            </motion.div>
          </Box>

          <Text className="mt-3" color="muted">
            By proceeding, you agree to our Terms of Use.
          </Text>
        </Flex>
      </Container>
    </Box>
  );
};
