"use client";
import { Box, Button, Flex, Text } from "ui";
import { Blur, Container } from "@/components/ui";
import { motion } from "framer-motion";

export const CallToAction = () => {
  return (
    <Box className="border-y border-gray-100 bg-gray-50 dark:border-dark-300 dark:bg-[#030303]">
      <Container className="relative max-w-7xl py-16 md:py-32">
        <Flex
          align="center"
          className="md:mt-18 mb-8 text-center"
          direction="column"
        >
          <motion.div
            initial={{ y: 20, opacity: 0 }}
            viewport={{ once: true, amount: 0.5 }}
            transition={{
              duration: 1,
              delay: 0,
            }}
            whileInView={{ y: 0, opacity: 1 }}
          >
            <Text
              as="h1"
              className="mt-6 h-max max-w-5xl pb-2 text-5xl md:text-7xl"
              color="gradient"
            >
              Experience the difference. Try it now!
            </Text>
          </motion.div>
          <motion.div
            initial={{ y: 20, opacity: 0 }}
            viewport={{ once: true, amount: 0.5 }}
            transition={{
              duration: 1,
              delay: 0.3,
            }}
            whileInView={{ y: 0, opacity: 1 }}
          >
            <Text
              className="mt-4 max-w-[600px] md:mt-6"
              color="muted"
              fontSize="xl"
              fontWeight="normal"
            >
              Streamline your workflows, empower your team, and nail every key
              objective all in one seamless platform.
            </Text>
          </motion.div>

          <motion.div
            className="mt-8 flex items-center gap-3"
            initial={{ y: 20, opacity: 0 }}
            viewport={{ once: true, amount: 0.5 }}
            transition={{
              duration: 1,
              delay: 0.6,
            }}
            whileInView={{ y: 0, opacity: 1 }}
          >
            <Button
              href="https://h1o9dmbqz9w.typeform.com/to/FWw2Ni3H"
              rounded="full"
              size="lg"
            >
              Join the waitlist
            </Button>
          </motion.div>
        </Flex>
        <Blur className="-translate-y-1/23 absolute -bottom-20 left-1/2 right-1/2 top-1/2 h-[800px] w-[800px] -translate-x-1/2 bg-warning/5" />
      </Container>
    </Box>
  );
};
