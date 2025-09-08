"use client";
import { Button, Flex, Text, Box } from "ui";
import { motion } from "framer-motion";
import { useSession } from "next-auth/react";
import { cn } from "lib";
import { Container, GoogleIcon } from "@/components/ui";
import { signInWithGoogle } from "@/lib/actions/sign-in";

export const Hero = () => {
  const { data: session } = useSession();

  return (
    <Box>
      <Box className="absolute inset-0 bg-[linear-gradient(to_right,#8080802a_1px,transparent_1px),linear-gradient(to_bottom,#8080801a_1px,transparent_1px)] bg-[size:45px_45px] dark:block" />
      <Box className="absolute inset-0 bg-[radial-gradient(ellipse_at_center,_var(--tw-gradient-stops))] from-transparent via-white/80 to-white dark:via-black/80 dark:to-black" />
      <Container className="pt-12">
        <Flex
          align="center"
          className="mb-8 mt-12 text-center md:mt-20"
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
              className="border-0 bg-[#dddddd]/30 px-5 text-sm backdrop-blur-xl dark:bg-dark-100/40 md:text-[0.95rem]"
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
              className={cn(
                "relative z-[1] mt-8 text-balance pb-2 text-5xl font-semibold md:max-w-5xl md:text-[4.25rem] md:leading-[1.1]",
              )}
            >
              The AI-Powered Project Management Platform for OKRs & Teams
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
            <Text className="mt-8 max-w-[700px] text-lg opacity-80 md:text-xl">
              Plan sprints, manage stories, and connect OKRs to daily work with
              built-in AI for summaries, updates, and automation.
            </Text>
          </motion.span>

          <Flex
            align="center"
            className="relative mt-6 justify-center gap-2 md:mt-8 md:gap-4"
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
                Get Started - It&apos;s free
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
