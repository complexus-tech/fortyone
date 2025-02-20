"use client";
import { Button, Flex, Text, Box } from "ui";
import { motion } from "framer-motion";
import { Container, Blur, GoogleIcon } from "@/components/ui";
import { signInWithGoogle } from "@/lib/actions/sign-in";

export const Hero = () => {
  return (
    <Box className="relative">
      <Blur className="absolute -top-96 left-1/2 right-1/2 h-[300px] w-[300px] -translate-x-1/2 bg-warning/5 md:h-[700px] md:w-[90vw]" />
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
              href="https://forms.gle/NmG4XFS5GhvRjUxu6"
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
              color="gradient"
            >
              Transform Goals into Victories, Right on Schedule
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
              className="mt-6 max-w-[600px] text-lg opacity-80 md:text-2xl"
              fontWeight="normal"
            >
              Track, align, and achieve team objectives with powerful insights
              that keep everyone moving forward.
            </Text>
          </motion.span>

          <Flex align="center" className="mt-10" gap={4}>
            <motion.span
              initial={{ y: -10, opacity: 0 }}
              transition={{
                duration: 1,
                delay: 0.4,
              }}
              viewport={{ once: true, amount: 0.5 }}
              whileInView={{ y: 0, opacity: 1 }}
            >
              <Button href="/signup" rounded="full" size="lg">
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
                className="px-4 md:pl-4 md:pr-5"
                color="tertiary"
                leftIcon={<GoogleIcon />}
                onClick={async () => {
                  await signInWithGoogle();
                }}
                rounded="full"
                size="lg"
              >
                Sign up with Google
              </Button>
            </motion.span>
          </Flex>
          <Text className="mt-6" color="muted">
            No credit card required.
          </Text>
        </Flex>
      </Container>
    </Box>
  );
};
