"use client";
import { Button, Flex, Text, Box } from "ui";
import { motion } from "framer-motion";
import { Container } from "@/components/ui";

export const Hero = () => {
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
              className="px-3 text-sm md:text-base"
              color="tertiary"
              rounded="full"
              size="sm"
            >
              Announcing Private Beta
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
              className="font-satosh mt-6 max-w-4xl pb-2 text-6xl font-semibold antialiased md:text-7xl md:leading-[1.1]"
              color="gradient"
            >
              Nail every objective on time with complexus.
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
              Empower your team to crush every key objective with our seamless
              project management platform.
            </Text>
          </motion.span>

          <Flex align="center" className="mt-10" gap={4}>
            {/* <motion.span
              initial={{ y: -10, opacity: 0 }}
              viewport={{ once: true, amount: 0.5 }}
              transition={{
                duration: 1,
                delay: 0.4,
              }}
              whileInView={{ y: 0, opacity: 1 }}
            >
              <Button
                className="border border-primary"
                rounded="full"
                size="lg"
                variant="outline"
              >
                Talk to us
              </Button>
            </motion.span> */}
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
                href="https://forms.gle/NmG4XFS5GhvRjUxu6"
                rounded="full"
                size="lg"
              >
                Join the waitlist
              </Button>
            </motion.span>
          </Flex>
          <Text className="mt-8" color="muted" fontSize="sm">
            No credit card required.
          </Text>
        </Flex>
      </Container>
    </Box>
  );
};
