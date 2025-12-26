"use client";

import { Box, Container, Flex, Text, Button } from "ui";
import { motion } from "framer-motion";
import { useSession } from "next-auth/react";
import Image from "next/image";
import { GoogleIcon } from "@/components/ui";
import { signInWithGoogle } from "@/lib/actions/sign-in";
import okrsImg from "../../../../public/images/product/objective.webp";
import okrsImgLight from "../../../../public/images/product/objective-light.webp";

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
              rounded="lg"
              size="sm"
            >
              OKRs
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
              <span className="text-stroke-white">Achieve</span> Goals with
              Objectives & Key Results
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
              Transform your organization with our powerful OKR framework. Set
              inspiring objectives, track measurable key results, and align your
              entire team around what matters most.
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
                className="px-3 md:pl-5 md:pr-4"
                color="invert"
                href="/signup"
                rounded="lg"
                size="lg"
              >
                <span className="hidden md:inline">Set Goals in Minutes</span>
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
                rounded="lg"
                size="lg"
              >
                {session ? "Continue with Google" : "Sign up with Google"}
              </Button>
            </motion.span>
          </Flex>
        </Flex>
        <Box className="relative mx-auto mt-16 max-w-6xl dark:hidden">
          <Image
            alt="OKRs Dashboard - Objectives and Key Results tracking"
            className="rounded border-[6px] border-gray-100 md:rounded-2xl"
            placeholder="blur"
            src={okrsImgLight}
          />
          <Box className="absolute inset-0 bg-linear-to-t from-white via-white via-20%" />
        </Box>
        <Box className="relative mx-auto mt-16 hidden max-w-6xl dark:block">
          <Image
            alt="OKRs Dashboard - Objectives and Key Results tracking"
            className="rounded border-[6px] dark:border-dark-100 md:rounded-2xl"
            placeholder="blur"
            src={okrsImg}
          />
          <Box className="absolute inset-0 bg-linear-to-t from-white via-white via-20%" />
        </Box>
      </Container>
    </Box>
  );
};
