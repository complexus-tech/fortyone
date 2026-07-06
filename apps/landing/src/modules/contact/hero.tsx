"use client";
import { Text } from "ui";
import { motion } from "framer-motion";
import { Container } from "@/components/ui";

export const Hero = () => {
  return (
    <Container className="relative pt-28 pb-10">
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
          as="h1"
          className="mb-6 max-w-4xl pb-2 text-5xl font-medium text-balance md:text-[3.5rem]"
        >
          Talk to the team behind FortyOne.
        </Text>
      </motion.div>
      <Text
        as="h2"
        className="text-text-muted max-w-2xl text-lg leading-8 md:text-xl"
        fontWeight="normal"
      >
        Ask about pricing, implementation, integrations, support, or whether
        FortyOne is the right fit for the way your team plans and tracks work.
      </Text>
    </Container>
  );
};
