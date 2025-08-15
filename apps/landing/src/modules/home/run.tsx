"use client";

import { Text, Button, Flex, Box, Divider } from "ui";
import { motion } from "framer-motion";
import { Container, GoogleIcon } from "@/components/ui";
import { signInWithGoogle } from "@/lib/actions/sign-in";

const viewport = { once: true, amount: 0.35 };
const fadeUp = {
  hidden: { y: 20, opacity: 0 },
  show: { y: 0, opacity: 1, transition: { duration: 0.7, ease: "easeOut" } },
};
const scaleXIn = {
  hidden: { scaleX: 0, opacity: 0, transformOrigin: "left" },
  show: {
    scaleX: 1,
    opacity: 1,
    transition: { duration: 0.7, ease: "easeOut" },
  },
};

export const RunEverything = () => {
  const phrases = [
    "Plan sprints.",
    "Track objectives.",
    "Deliver on time.",
    "Work smarter with AI.",
  ];

  return (
    <Box className="bg-gradient-to-b from-white to-gray-50 pb-24 pt-28 dark:from-dark dark:via-black dark:to-black dark:pb-10">
      <Container>
        <motion.div
          initial="hidden"
          variants={fadeUp}
          viewport={viewport}
          whileInView="show"
        >
          <Text className="mb-8 w-max text-xl">It&apos;s time to ship</Text>
        </motion.div>
        <Text
          as="h2"
          className="mb-10 w-max text-5xl font-semibold md:text-[4rem] md:leading-[1.2]"
          color="gradient"
        >
          {phrases.map((line, index) => (
            <motion.span
              initial="hidden"
              key={line}
              transition={{ delay: index * 0.3 }}
              variants={fadeUp}
              viewport={viewport}
              whileInView="show"
            >
              <span className="block">{line}</span>
            </motion.span>
          ))}
        </Text>
        <Flex align="center" className="gap-2 md:gap-4" wrap>
          <motion.span
            initial="hidden"
            variants={fadeUp}
            viewport={viewport}
            whileInView="show"
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
            initial="hidden"
            transition={{ delay: 0.15 }}
            variants={fadeUp}
            viewport={viewport}
            whileInView="show"
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
        <motion.div
          initial="hidden"
          variants={scaleXIn}
          viewport={viewport}
          whileInView="show"
        >
          <Box className="mt-24 hidden dark:block">
            <Divider />
          </Box>
        </motion.div>
      </Container>
    </Box>
  );
};
