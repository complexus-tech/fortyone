"use client";
import React from "react";
import { Box, Text } from "ui";
import { motion } from "framer-motion";
import { Container } from "@/components/ui";

const Brand = ({ logo }: { logo: string }) => {
  return (
    <img
      alt="brand logo"
      className="3xl:h-20 mr-8 block h-8 w-auto grayscale dark:invert md:mr-16 md:h-12 md:justify-self-center"
      key={logo}
      loading="lazy"
      src={logo}
    />
  );
};

export const SampleClients = () => {
  const brands = [
    "/images/brands/miningo.svg",
    "/images/brands/mds.svg",
    "/images/brands/nesbil.png",
    "/images/brands/zimboriginal.png",
    "/images/brands/digitank.png",
    "/images/brands/wastemate.png",
  ];
  return (
    <Container className="relative">
      <Box className="py-16 md:py-32">
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
            as="h2"
            className="mx-auto max-w-4xl text-center text-3xl md:text-5xl md:leading-tight"
          >
            <Text
              as="span"
              className="text-stroke-white relative opacity-80"
              fontWeight="semibold"
            >
              Join{" "}
            </Text>
            these ambitious teams relying on us to crush their{" "}
            <Text
              as="span"
              className="text-stroke-whitee opacity-80"
              fontWeight="semibold"
            >
              objectives{" "}
            </Text>
            and drive{" "}
            <Text as="span" className="text-stroke-white" fontWeight="semibold">
              growth.
            </Text>
          </Text>
        </motion.div>
        <Box className="mx-auto mt-16 grid max-w-5xl grid-cols-2 gap-x-4 gap-y-10 md:mt-20 md:grid-cols-4 md:gap-y-20">
          {brands.map((logo) => (
            <Brand key={logo} logo={logo} />
          ))}
        </Box>
      </Box>
    </Container>
  );
};
