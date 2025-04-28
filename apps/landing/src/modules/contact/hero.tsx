"use client";
import { Text, Button } from "ui";
import { motion } from "framer-motion";
import { Container } from "@/components/ui";

export const Hero = () => {
  return (
    <Container className="relative max-w-4xl pb-10 pt-36 md:pt-40">
      <Button
        className="mx-auto cursor-text px-3 text-sm md:text-base"
        color="tertiary"
        rounded="full"
        size="sm"
      >
        Get in Touch
      </Button>
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
          align="center"
          as="h1"
          className="mx-auto my-8 max-w-2xl pb-2 text-5xl font-semibold md:text-7xl"
        >
          How Can We Help You{" "}
          <Text as="span" className="text-stroke-white">
            Today?
          </Text>
        </Text>
      </motion.div>
      <Text
        align="center"
        as="h2"
        className="mb-4 text-lg leading-snug opacity-80 md:text-2xl"
        fontWeight="normal"
      >
        Connect with our team for product demos, implementation support, or any
        questions about getting started with complexus.
      </Text>
    </Container>
  );
};
