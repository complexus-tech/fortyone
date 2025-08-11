"use client";
import { Box, Button, Flex, Text } from "ui";
import { motion } from "framer-motion";
import { useSession } from "next-auth/react";
import { Container, GoogleIcon } from "@/components/ui";
import { signInWithGoogle } from "@/lib/actions/sign-in";

export const CallToAction = () => {
  const { data: session } = useSession();
  return (
    <Box className="border-b border-gray-100 bg-gradient-to-t from-gray-50">
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
              as="h2"
              className="mt-6 h-max max-w-4xl pb-2 text-5xl font-semibold md:text-7xl"
            >
              Work smarter with AI that’s in the loop.
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

          <Box className="mt-8 flex flex-wrap items-center justify-center gap-3">
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
                color="invert"
                href="/signup"
                rounded="lg"
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
                className="border-gray-200 px-4 md:pl-3.5 md:pr-4"
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
          </Box>
        </Flex>
      </Container>
    </Box>
  );
};
