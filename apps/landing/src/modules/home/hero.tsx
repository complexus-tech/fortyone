"use client";
import { Button, Flex, Text, Box } from "ui";
import { motion } from "framer-motion";
import { useSession } from "next-auth/react";
import { Container, GoogleIcon } from "@/components/ui";
import { signInWithGoogle } from "@/lib/actions/sign-in";

export const Hero = () => {
  const { data: session } = useSession();

  return (
    <Box>
      <Container className="pt-12 md:pt-8">
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
              className="border-0 bg-[#dddddd]/40 px-5 text-sm font-medium backdrop-blur-xl dark:bg-dark-100/40 md:text-[0.95rem]"
              color="tertiary"
              href="/signup"
              rounded="full"
              size="sm"
            >
              Free forever. No credit card required.
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
              className="mt-6 pb-2 text-5xl font-semibold md:max-w-5xl md:text-[4.7rem] md:leading-[1.1]"
            >
              The Everything App for{" "}
              <Text as="span" className="text-stroke-white">
                Projects
              </Text>{" "}
              & OKRs
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
            <Text className="mt-4 max-w-[700px] text-lg opacity-80 dark:font-normal md:text-xl">
              AI‑powered all‑in‑one Projects & OKRs platform that connects daily
              work to strategic goals, tracks real‑time progress, and predicts
              risks
            </Text>
          </motion.span>

          <Flex
            align="center"
            className="relative mt-6 justify-center gap-2 md:mt-6 md:gap-4"
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
                rounded="full"
                size="lg"
              >
                <span className="hidden md:inline">
                  Get Started - It&apos;s free
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
                variant="naked"
              >
                {session ? "Continue with Google" : "Sign up with Google"}
              </Button>
            </motion.span>
          </Flex>
        </Flex>
      </Container>
    </Box>
  );
};
