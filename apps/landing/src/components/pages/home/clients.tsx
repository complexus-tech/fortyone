"use client";
import React from "react";
import { Box, Text } from "ui";
import Marquee from "react-fast-marquee";
import { Container, Logo } from "@/components/ui";
import { motion } from "framer-motion";

const Brand = ({ logo }: { logo: string }) => {
  return (
    // eslint-disable-next-line @next/next/no-img-element
    <img
      key={logo}
      src={logo}
      loading="lazy"
      alt="brand logo"
      className="3xl:h-20 mr-8 block h-8 w-auto grayscale invert md:mr-16 md:h-10 md:justify-self-center"
    />
  );
};

export const SampleClients = () => {
  const brands = [
    "/images/brands/miningo.svg",
    "/images/brands/mds.svg",
    "/images/brands/zimboriginal.png",
    "/images/brands/digitank.png",
    "/images/brands/miningo.svg",
    "/images/brands/wastemate.png",
    "/images/brands/nesbil.png",
  ];
  return (
    <Container className="relative md:mt-16">
      <Box className="py-16 md:py-28">
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
            as="h3"
            className="text-center text-xl md:text-2xl"
            fontWeight="normal"
          >
            Hundreds of teams rely on us to crush their{" "}
            <Text as="span" color="primary" fontWeight="semibold">
              objectives.
            </Text>
          </Text>
        </motion.div>
        <Marquee className="mt-20" pauseOnHover speed={40}>
          {brands.map((logo) => (
            <Brand key={logo} logo={logo} />
          ))}
        </Marquee>
      </Box>
    </Container>
  );
};
