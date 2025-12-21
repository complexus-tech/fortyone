"use client";
import React from "react";
import { Container } from "@/components/ui";
import { Flex, Text } from "ui";

const Brand = ({ logo }: { logo: string }) => {
  return (
    <img
      alt="brand logo"
      className="3xl:h-20 mr-8 block h-8 w-auto grayscale dark:invert md:mr-16 md:h-9 md:justify-self-center"
      key={logo}
      loading="lazy"
      src={logo}
    />
  );
};

export const SampleClients = () => {
  const brands = [
    "/images/brands/digitank.png",
    "/images/brands/miningo.svg",
    "/images/brands/mds.svg",
    "/images/brands/nesbil.png",
    "/images/brands/zimboriginal.png",
  ];

  return (
    <Container className="relative z-10 mt-16">
      <Flex className="gap-10">
        <Text color="muted" fontSize="sm">
          Trusted by leading <br />
          product teams at
        </Text>
        <Flex>
          {brands.map((logo) => (
            <Brand key={logo} logo={logo} />
          ))}
        </Flex>
      </Flex>
      {/* <Marquee
        // gradient
        // gradientColor={resolvedTheme === "light" ? "#ffffff" : "#08090a"}
        pauseOnHover
        speed={30}
      >
        {brands.map((logo) => (
          <Brand key={logo} logo={logo} />
        ))}
      </Marquee> */}
    </Container>
  );
};
