"use client";
import React from "react";
import { Box, Container, Text } from "ui";
import Marquee from "react-fast-marquee";
import { Logo } from "@/components/ui";

export const SampleClients = () => {
  return (
    <Container className="relative mt-10">
      <Box className="py-20">
        <Text as="h3" className="text-center" fontSize="xl">
          Trusted by the hundreds of teams to manage projects.
        </Text>
        <Marquee className="mt-16" pauseOnHover speed={40}>
          {Array.from({ length: 12 }).map((_, i) => (
            <Logo className="mr-12 h-10 text-gray-100" key={i} />
          ))}
        </Marquee>
      </Box>
    </Container>
  );
};
