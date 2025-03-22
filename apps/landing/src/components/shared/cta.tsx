"use client";
import { Box, Button, Flex, Text } from "ui";
import { motion } from "framer-motion";
import { useSession } from "next-auth/react";
import { Container, GoogleIcon } from "@/components/ui";
import { signInWithGoogle } from "@/lib/actions/sign-in";

export const CallToAction = () => {
  const { data: session } = useSession();
  return (
    <Box className="border-y border-gray-100 bg-gray-50 dark:border-dark-300 dark:bg-dark/80">
      <Container className="relative max-w-7xl py-16 md:py-32">
        <Flex
          align="center"
          className="md:mt-18 mb-8 text-center"
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
              as="h1"
              className="mt-6 h-max max-w-4xl pb-2 text-5xl font-semibold md:text-7xl"
              color="gradient"
            >
              Set Objectives. Drive Outcomes.
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
              className="mt-4 max-w-[600px] md:mt-6"
              color="muted"
              fontSize="xl"
            >
              Bring your objectives, OKRs, and sprints together. The modern way
              to align teams and deliver meaningful outcomes.
            </Text>
          </motion.div>

          <Box className="mt-8 flex items-center gap-3">
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
                className="md:px-4"
                href="/signup"
                rounded="full"
                size="lg"
              >
                Manage Projects Free
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
                className="px-4 md:pl-3.5 md:pr-4"
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
          </Box>
        </Flex>
      </Container>
    </Box>
  );
};
