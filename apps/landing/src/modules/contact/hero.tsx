"use client";
import { Text } from "ui";
import { motion } from "framer-motion";
import { Container } from "@/components/ui";

export const Hero = () => {
  return (
    <Container className="relative max-w-4xl pb-10 pt-28">
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
          className="mx-auto my-8 max-w-xl pb-2 text-5xl font-bold md:text-6xl"
        >
          How can we help you{" "}
          <Text as="span" className="text-stroke-white">
            today?
          </Text>
        </Text>
      </motion.div>
      <Text
        align="center"
        as="h2"
        className="mx-auto mb-4 max-w-3xl text-lg leading-snug opacity-80 md:text-2xl"
        fontWeight="normal"
      >
        Connect with our team for product demos, implementation support, or any
        questions about getting started with complexus.
      </Text>
    </Container>
  );
};
