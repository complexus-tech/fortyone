"use client";
import { Box, BlurImage, Text, Flex, Button } from "ui";
import { motion } from "framer-motion";
import { useSession } from "next-auth/react";
import { Container, GoogleIcon } from "@/components/ui";
import { signInWithGoogle } from "@/lib/actions/sign-in";

export const Transform = () => {
  const { data: session } = useSession();
  return (
    <Box className="relative grid bg-gradient-to-r from-gray-100 via-gray-50 to-gray-50 py-20 dark:from-dark dark:via-black dark:to-black md:grid-cols-2 md:py-0">
      <Box />
      <Box className="relative hidden md:block">
        <BlurImage
          alt="Login"
          className="h-[85vh] w-full dark:hidden"
          imageClassName="object-top"
          quality={100}
          src="/images/login-light.webp"
        />
        <BlurImage
          alt="Login"
          className="hidden h-[85vh] w-full dark:block"
          imageClassName="object-top"
          quality={100}
          src="/images/login.webp"
        />
        <Box className="absolute inset-0 bg-gradient-to-t from-gray-50 via-gray-50 via-25% dark:from-black dark:via-black" />
      </Box>
      {/* <div className="pointer-events-none absolute inset-0 z-[3] bg-[url('/noise.png')] bg-repeat opacity-70" /> */}
      <Box className="z-[3] md:absolute md:inset-0">
        <Container className="grid-cols-2 gap-10 md:grid md:h-full">
          <Flex direction="column" justify="center">
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
                className="mx-auto border-gray-200 px-3 text-sm md:mx-0 md:text-base"
                color="tertiary"
                href="/signup"
                rounded="lg"
                size="sm"
              >
                Ready to get started?
              </Button>
            </motion.span>
            <Text
              as="h2"
              className="mt-6 text-center text-5xl font-semibold md:text-left md:text-7xl"
            >
              <span className="text-stroke-white">Transform</span> how your team
              works <span className="text-stroke-white relative">today.</span>
            </Text>
            <Text
              className="mt-10 max-w-[600px] text-center text-lg opacity-80 md:mt-16 md:text-left md:text-2xl"
              fontWeight="normal"
            >
              Join innovative teams who use complexus to deliver projects faster
              and with better results.
            </Text>
            <Flex
              align="center"
              className="relative mt-6 justify-center gap-2 md:mt-10 md:justify-start md:gap-4"
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
                  className="px-3 md:px-5"
                  color="invert"
                  href="/signup"
                  rounded="lg"
                  size="lg"
                >
                  Get Started Free
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
                  className="border-gray-200 px-3 md:pl-3.5 md:pr-4"
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
        </Container>
      </Box>
    </Box>
  );
};
