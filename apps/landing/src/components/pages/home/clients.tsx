"use client";
import React from "react";
import { Box, Text } from "ui";
import Marquee from "react-fast-marquee";
import { Container, Logo } from "@/components/ui";

export const SampleClients = () => {
  return (
    <Container className="relative md:mt-16">
      <Box className="py-16 md:py-28">
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
        <Marquee className="mt-20" pauseOnHover speed={40}>
          {Array.from({ length: 12 }).map((_, i) => (
            <Logo className="mr-8 h-8 text-gray-100 md:mr-12 md:h-10" key={i} />
          ))}
        </Marquee>
      </Box>
    </Container>
  );
};
