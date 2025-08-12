"use client";

import { Text, Button, Flex, Box } from "ui";
import { motion } from "framer-motion";
import { Container, GoogleIcon } from "@/components/ui";
import { signInWithGoogle } from "@/lib/actions/sign-in";

export const RunEverything = () => {
  return (
    <Box className="bg-gradient-to-b from-white to-gray-50 py-28">
      <Container>
        <Text className="mb-10 text-5xl">It&apos;s time to ship</Text>
        <Text
          className="mb-10 text-5xl font-semibold md:text-[5rem] md:leading-[1.3]"
          color="gradientDark"
        >
          Plan sprints. Track objectives. <br /> Hit OKRs. Deliver on time.{" "}
          <br /> Work smarter with AI.
        </Text>
        <Flex align="center" className="gap-2 md:gap-4" wrap>
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
              Continue with Google
            </Button>
          </motion.span>
        </Flex>
      </Container>
    </Box>
  );
};
