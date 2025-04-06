"use client";
import React from "react";
import { Box, Text } from "ui";
import { motion } from "framer-motion";
import { Container } from "@/components/ui";

const Brand = ({ logo }: { logo: string }) => {
  return (
    <img
      alt="brand logo"
      className="3xl:h-20 mr-8 block h-8 w-auto grayscale invert md:mr-16 md:h-12 md:justify-self-center"
      key={logo}
      loading="lazy"
      src={logo}
    />
  );
};

export const SampleClients = () => {
  const brands = [
    "/images/brands/miningo.svg",
    // "/images/brands/mds.svg",
    "/images/brands/nesbil.png",
    "/images/brands/zimboriginal.png",
    "/images/brands/digitank.png",
    "/images/brands/miningo.svg",
    "/images/brands/wastemate.png",
  ];
  return (
    <Container className="relative md:mt-16">
      <img
        alt=""
        className="absolute left-0 top-0 hidden h-20 w-auto -rotate-12 opacity-20 invert md:-bottom-28 md:inline-block"
        src="/svgs/xx.svg"
      />
      <img
        alt=""
        className="absolute right-0 top-0 hidden h-24 w-auto -rotate-[40deg] opacity-20 invert md:-bottom-28 md:inline-block"
        src="/svgs/arrow-2.svg"
      />
      <Box className="py-16 md:py-28">
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
            as="h3"
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
              className="text-stroke-white relative opacity-80"
              fontWeight="semibold"
            >
              objectives{" "}
            </Text>
            and drive{" "}
            <Text
              as="span"
              className="relative"
              color="gradient"
              fontWeight="semibold"
            >
              growth
              <img
                alt=""
                className="absolute -bottom-20 left-0 h-auto w-full -rotate-12 opacity-80 invert md:-bottom-20"
                src="/svgs/arrow.svg"
              />
            </Text>
            .
          </Text>
        </motion.div>
        <Box className="mx-auto mt-36 grid max-w-5xl grid-cols-2 gap-x-4 gap-y-20 md:grid-cols-3">
          {brands.map((logo) => (
            <Brand key={logo} logo={logo} />
          ))}
        </Box>
      </Box>
      <Box className="absolute bottom-0 right-0 opacity-5">
        <svg
          className="3xl:w-40 h-auto w-20 rotate-6 text-white xl:w-32 2xl:w-36"
          fill="none"
          height="470"
          viewBox="0 0 266 470"
          width="266"
          xmlns="http://www.w3.org/2000/svg"
        >
          <line
            stroke="currentColor"
            x1="202.44"
            x2="0.43993"
            y1="33.2376"
            y2="407.238"
          />
          <line
            stroke="currentColor"
            x1="265.439"
            x2="10.4393"
            y1="0.238835"
            y2="469.239"
          />
          <circle cx="123.5" cy="212.5" r="86" stroke="currentColor" />
        </svg>
      </Box>
    </Container>
  );
};
