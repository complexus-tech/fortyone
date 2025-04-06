"use client";
import { Button, Flex, Text, Box } from "ui";
import { motion } from "framer-motion";
import { useSession } from "next-auth/react";
import { ArrowRight2Icon } from "icons";
import { Container, GoogleIcon, Blur } from "@/components/ui";
import { signInWithGoogle } from "@/lib/actions/sign-in";

export const Hero = () => {
  const { data: session } = useSession();
  return (
    <Box className="relative">
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
              className="px-3 text-sm md:text-base"
              color="tertiary"
              href="/signup"
              rounded="full"
              size="sm"
            >
              Join Our Exclusive Beta
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
              className="mt-6 pb-2 text-5xl font-semibold antialiased md:max-w-4xl md:text-7xl md:leading-[1.1]"
            >
              <span className="text-stroke-white font-bold">Project</span>{" "}
              Management That Works for{" "}
              <Text as="span" className="relative" color="gradient">
                You
                <img
                  alt=""
                  className="absolute -bottom-20 left-0 h-auto w-full -rotate-12 opacity-80 invert md:-bottom-28"
                  src="/svgs/arrow.svg"
                />
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
              className="mt-8 max-w-[600px] text-lg opacity-80 md:text-2xl"
              fontWeight="normal"
            >
              Track, align, and achieve team objectives with powerful insights
              that keep everyone moving forward.
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
                rightIcon={<ArrowRight2Icon className="dark:text-gray-200" />}
                rounded="full"
                size="lg"
              >
                <span className="hidden md:inline">Manage in 3 minutes</span>
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
          <Text className="mt-6" color="muted">
            No credit card required.
          </Text>
        </Flex>
      </Container>
      <Blur className="absolute bottom-1/2 left-1/2 right-1/2 top-1/2 hidden h-[600px] w-[600px] -translate-x-1/2 -translate-y-1/2 bg-warning/[0.07] md:block" />
    </Box>
  );
};
